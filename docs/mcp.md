# MCP server

`POST /mcp`, outside `/api`, is Casier's remote MCP server. Any MCP client holding a Registre
access token for `https://casier.facile.studio/mcp` can call it;
`/.well-known/oauth-protected-resource/mcp` says how to get one. Scopes are `casier.read` and
`casier.write`. Every tool calls the REST controllers as the person, so project roles, secret
ACLs and protected environments apply exactly as in the web app.

Bulle has no Casier access and gets none: secret values never reach the suite assistant.

## Tools

| Tool | Does | Destructive | Idempotent |
|---|---|---|---|
| `casier_list_projects` | Projects, roles and environment slugs | no | yes |
| `casier_list_secrets` | Secret metadata of one environment, paged | no | yes |
| `casier_get_secret` | One secret's metadata and version history | no | yes |
| `casier_diff_environments` | Keys added, removed or changed between two environments | no | yes |
| `casier_list_trash` | Deleted secrets that can still be restored | no | yes |
| `casier_list_audit_events` | A project's audit trail | no | yes |
| `casier_create_project` | Creates a project | no | no |
| `casier_create_environment` | Adds an environment | no | no |
| `casier_set_secret` | Writes a value as a new version | yes | no |
| `casier_tag_secret` | Replaces a secret's tags | yes | yes |
| `casier_rollback_secret` | Makes an earlier version current, as a new version | yes | no |
| `casier_restore_secret` | Brings a secret back from the trash | no | yes |
| `casier_reveal_secrets` | Values of up to 50 keys, after the person confirms | no | no |
| `casier_delete_secret` | Moves a secret to the trash, after the person confirms | yes | no |

The hints follow the suite rule: a tool that overwrites or removes data is destructive, and only
a tool whose repeat call has no extra effect is idempotent. The reveal and the delete are not
idempotent: a tool that asks the person to confirm never is, even though a repeat call from the
same client reuses the open confirmation instead of opening another.
`TestEveryToolStatesTheSuiteHints` pins this table.

## Reveal

Only `casier_reveal_secrets` returns values, and only after the person confirms: in the client
when it can elicit, otherwise on Casier's `/pending` page. A confirmation on the page shows no
value. The client that asked then calls again with `confirmation_id`, within 15 minutes, and
gets the values once.

- **Result budget.** One answer stays within tronc's `ResultBudget` (75,000 bytes). Casier reads
  the keys in the order asked and stops before the first value that would overflow it. The keys
  not sent are listed in `omitted`; the client asks for them again, and the person confirms
  again. That one value is read, so the audit trail shows a `secrets.reveal` for it, but it is
  not sent.
- **One value too large.** When the first value alone overflows the budget, nothing is sent and
  the confirmation stays open. Read that value with the CLI or the web app.
- **Spent once, and only on success.** `collect` locks the confirmed action, reads the values,
  and marks the action `revealed` in the same transaction. A failed read, a changed secret or an
  oversized value leaves it unspent. A second collection waits for the first and is then
  refused.
- **A retry audits again.** Casier audits each value as it reads it, before it knows whether
  the answer will be sent. When a read fails after some keys were read, or the first value is
  too large, the confirmation stays open and each retry adds another `secrets.reveal` row for
  the same keys. The trail over-reports; it never hides a read.

## Not exposed, on purpose

Casier is a secrets manager, so the MCP surface stays minimal.

- **Deleting a project or an environment.** Both are hard deletes: every secret and every
  version in them is gone, with no trash and no restore. An agent that picks the wrong slug
  would destroy what `casier_delete_secret` is built to keep recoverable. Delete them in the web
  app.
- **Updating a project or an environment.** The environment update carries the protection
  flag, and unprotecting an environment is the step before an unprotected write to it. The
  project update carries the rotation policy, which `0` turns off.
- **Purging a secret.** A purge is a hard delete of every version, with no restore, like the
  project and environment deletes. REST lets an owner or admin purge with a write-scoped API
  token; MCP leaves it out so an agent can only trash a secret, which
  `casier_restore_secret` undoes.
- **ACLs, members, invites, API tokens, webhooks and key rotation.** These are administration.
  REST accepts them only from a person's browser session (`middleware.RequireUserSession`), and
  MCP keeps that line.

## Confirm page API

`/api/mcp/pending` serves the confirm page: `GET /api/mcp/pending`,
`POST /api/mcp/pending/{id}/confirm` and `POST /api/mcp/pending/{id}/dismiss`. Only the session
cookie a browser sign-in minted passes (`mcpauth.RequireOwnSession`). A bearer, a CLI session, a
`casier_` token, a delegated token or any request with an `Authorization` header answers 403,
and the page then asks the person to sign in again in the browser. A cookie from before porte
v0.20.0 carries no browser mark, so each person signs in once more before their first confirm.

A reveal names the client that asked, so two clients asking for the same keys each get their
own confirmation. Resolved and expired actions are pruned 30 days later, hourly. Erasing a
person deletes their `mcp_actions` rows.
