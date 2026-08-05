use anyhow::{bail, Context, Result};
use reqwest::Client;
use serde::Deserialize;

pub struct ApiClient {
    base_url: String,
    token: Option<String>,
    client: Client,
}

#[derive(Deserialize)]
pub struct AuthResponse {
    pub token: String,
}

#[derive(Deserialize)]
pub struct AuthConfigResponse {
    #[serde(default)]
    pub sso_only: bool,
    #[serde(default)]
    pub oidc_enabled: bool,
}

#[derive(Deserialize)]
pub struct MeResponse {
    pub email: String,
}

#[derive(Deserialize)]
pub struct Project {
    pub name: String,
    pub slug: String,
    #[serde(default)]
    pub description: String,
}

/// `value` is absent unless the request asked to reveal it. It is an
/// `Option<String>` rather than a defaulted `String` on purpose: a server that
/// withholds a value and a secret whose value is genuinely empty must not
/// deserialize into the same thing, or `casier run` would inject an empty
/// variable and the child process would fail somewhere far away from the cause.
#[derive(Debug, Deserialize)]
pub struct Secret {
    pub key: String,
    #[serde(default)]
    pub value: Option<String>,
    #[serde(default)]
    pub version: i32,
}

#[derive(Deserialize)]
struct ProjectsResponse {
    #[serde(default)]
    projects: Option<Vec<Project>>,
}

#[derive(Deserialize)]
struct SecretsResponse {
    #[serde(default)]
    secrets: Option<Vec<Secret>>,
}

#[derive(Deserialize)]
pub struct ImportResponse {
    pub created: i32,
    pub updated: i32,
    pub skipped: i32,
}

#[derive(Deserialize)]
pub struct ExportResponse {
    pub content: String,
}

impl ApiClient {
    pub fn new(base_url: &str, token: Option<String>) -> Self {
        Self {
            base_url: base_url.trim_end_matches('/').to_string(),
            token,
            client: Client::new(),
        }
    }

    fn request(&self, method: reqwest::Method, path: &str) -> reqwest::RequestBuilder {
        let url = format!("{}{}", self.base_url, path);
        let mut req = self.client.request(method, &url);
        if let Some(ref token) = self.token {
            req = req.bearer_auth(token);
        }
        req
    }

    pub async fn auth_config(&self) -> Result<AuthConfigResponse> {
        let resp = self
            .request(reqwest::Method::GET, "/auth/config")
            .send()
            .await
            .with_context(|| format!("failed to reach {}", self.base_url))?;
        if !resp.status().is_success() {
            bail!(
                "GET {}/auth/config returned {}",
                self.base_url,
                resp.status()
            );
        }
        let body = resp.text().await?;
        serde_json::from_str(&body).with_context(|| {
            format!(
                "{} did not answer with a Casier API response",
                self.base_url
            )
        })
    }

    pub async fn login(&self, email: &str, password: &str) -> Result<AuthResponse> {
        let resp = self
            .request(reqwest::Method::POST, "/auth/login")
            .json(&serde_json::json!({ "email": email, "password": password }))
            .send()
            .await
            .context("failed to reach server")?;
        if !resp.status().is_success() {
            let status = resp.status();
            let body = resp.text().await.unwrap_or_default();
            bail!("login failed ({}): {}", status, body);
        }
        Ok(resp.json().await?)
    }

    pub async fn me(&self) -> Result<MeResponse> {
        let resp = self
            .request(reqwest::Method::GET, "/auth/me")
            .send()
            .await
            .context("failed to reach server")?;
        if !resp.status().is_success() {
            bail!("GET /auth/me returned {}", resp.status());
        }
        Ok(resp.json().await?)
    }

    pub async fn list_projects(&self) -> Result<Vec<Project>> {
        let resp = self
            .request(reqwest::Method::GET, "/projects")
            .send()
            .await
            .context("failed to reach server")?;
        if !resp.status().is_success() {
            bail!("GET /projects returned {}", resp.status());
        }
        let body: ProjectsResponse = resp.json().await?;
        Ok(body.projects.unwrap_or_default())
    }

    /// Lists secrets without their values. Use it when only the keys matter.
    pub async fn list_secrets(&self, project: &str, env: &str) -> Result<Vec<Secret>> {
        self.fetch_secrets(project, env, false).await
    }

    /// Lists secrets with their values, which is what every command that hands
    /// secrets to something else needs. The server withholds values unless
    /// asked, so this asks, and then insists it was answered.
    pub async fn list_secrets_with_values(&self, project: &str, env: &str) -> Result<Vec<Secret>> {
        require_values(self.fetch_secrets(project, env, true).await?)
    }

    async fn fetch_secrets(&self, project: &str, env: &str, reveal: bool) -> Result<Vec<Secret>> {
        let mut path = format!("/projects/{}/environments/{}/secrets", project, env);
        if reveal {
            path.push_str("?reveal=true");
        }
        let resp = self
            .request(reqwest::Method::GET, &path)
            .send()
            .await
            .context("failed to reach server")?;
        if !resp.status().is_success() {
            bail!("GET {} returned {}", path, resp.status());
        }
        let body: SecretsResponse = resp.json().await?;
        Ok(body.secrets.unwrap_or_default())
    }

    pub async fn set_secret(
        &self,
        project: &str,
        env: &str,
        key: &str,
        value: &str,
    ) -> Result<Secret> {
        let path = format!("/projects/{}/environments/{}/secrets", project, env);
        let resp = self
            .request(reqwest::Method::POST, &path)
            .json(&serde_json::json!({ "key": key, "value": value }))
            .send()
            .await
            .context("failed to reach server")?;
        if !resp.status().is_success() {
            let status = resp.status();
            let body = resp.text().await.unwrap_or_default();
            bail!("POST {} failed ({}): {}", path, status, body);
        }
        Ok(resp.json().await?)
    }

    pub async fn get_secret(&self, project: &str, env: &str, key: &str) -> Result<Secret> {
        let secrets = self.list_secrets_with_values(project, env).await?;
        secrets
            .into_iter()
            .find(|s| s.key == key)
            .context(format!("secret '{}' not found", key))
    }

    pub async fn import_env(
        &self,
        project: &str,
        env: &str,
        content: &str,
    ) -> Result<ImportResponse> {
        let path = format!("/projects/{}/environments/{}/secrets/import", project, env);
        let resp = self
            .request(reqwest::Method::POST, &path)
            .json(&serde_json::json!({ "content": content }))
            .send()
            .await
            .context("failed to reach server")?;
        if !resp.status().is_success() {
            let status = resp.status();
            let body = resp.text().await.unwrap_or_default();
            bail!("POST {} failed ({}): {}", path, status, body);
        }
        Ok(resp.json().await?)
    }

    pub async fn export_env(&self, project: &str, env: &str) -> Result<ExportResponse> {
        let path = format!("/projects/{}/environments/{}/secrets/export", project, env);
        let resp = self
            .request(reqwest::Method::GET, &path)
            .send()
            .await
            .context("failed to reach server")?;
        if !resp.status().is_success() {
            bail!("GET {} returned {}", path, resp.status());
        }
        Ok(resp.json().await?)
    }

    pub async fn delete_secret(&self, project: &str, env: &str, key: &str) -> Result<()> {
        let path = format!("/projects/{}/environments/{}/secrets/{}", project, env, key);
        let resp = self
            .request(reqwest::Method::DELETE, &path)
            .send()
            .await
            .context("failed to reach server")?;
        if !resp.status().is_success() {
            let status = resp.status();
            let body = resp.text().await.unwrap_or_default();
            bail!("DELETE {} failed ({}): {}", path, status, body);
        }
        Ok(())
    }
}

/// Refuses a response that named secrets but withheld their values. A server
/// only does that when it was not asked to reveal them, which after the reveal
/// gating landed means this binary predates the server. Injecting an empty
/// environment would let the child process start and fail somewhere unrelated,
/// so nothing downstream is allowed to proceed on a partial answer.
fn require_values(secrets: Vec<Secret>) -> Result<Vec<Secret>> {
    let withheld: Vec<&str> = secrets
        .iter()
        .filter(|s| s.value.is_none())
        .map(|s| s.key.as_str())
        .collect();
    if !withheld.is_empty() {
        bail!(
            "the server returned {} secret(s) without their values ({}).\n\
             This CLI is older than the server it is talking to — reinstall it, \
             then try again.\n\
             Refusing to continue with an incomplete environment.",
            withheld.len(),
            withheld.join(", ")
        );
    }
    Ok(secrets)
}

#[cfg(test)]
mod tests {
    use super::*;

    fn secret(key: &str, value: Option<&str>) -> Secret {
        Secret {
            key: key.to_string(),
            value: value.map(|v| v.to_string()),
            version: 1,
        }
    }

    #[test]
    fn values_present_pass_through() {
        let secrets = require_values(vec![secret("A", Some("1")), secret("B", Some(""))]).unwrap();
        assert_eq!(secrets.len(), 2);
    }

    #[test]
    fn an_empty_value_is_not_a_missing_one() {
        let secrets = require_values(vec![secret("EMPTY", Some(""))]).unwrap();
        assert_eq!(secrets[0].value.as_deref(), Some(""));
    }

    #[test]
    fn a_withheld_value_refuses_and_names_the_key() {
        let err = require_values(vec![secret("A", Some("1")), secret("MISSING", None)])
            .unwrap_err()
            .to_string();
        assert!(err.contains("MISSING"), "error did not name the key: {err}");
        assert!(err.contains("Refusing to continue"), "{err}");
    }

    #[test]
    fn no_secrets_at_all_is_not_an_error() {
        assert!(require_values(vec![]).unwrap().is_empty());
    }
}
