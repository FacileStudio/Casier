-- Baseline: Casier's schema as the imperative migrations built it,
-- captured from a scratch database provisioned through the pre-goose
-- schemas.Migrate path (AutoMigrate + space->project rename + token-scope
-- backfill + porte adoption) on 2026-08-13 with pg_dump --schema-only and
-- stripped of environment noise (SET preamble, \restrict markers, ownership,
-- and the public. qualification so the migration follows search_path rather
-- than pinning one schema).
--
-- Production is stamped at this version rather than running it; see
-- Wiki/MIGRATIONS.md. It runs for real only against an empty database.

-- +goose Up
CREATE TABLE api_token_environments (
    token_id bigint NOT NULL,
    environment_id bigint NOT NULL
);
CREATE TABLE api_tokens (
    id bigint NOT NULL,
    token_hash text,
    name text,
    project_id bigint,
    all_environments boolean DEFAULT false,
    permissions text DEFAULT 'read'::text,
    created_by bigint,
    expires_at timestamp with time zone,
    last_used_at timestamp with time zone,
    revoked_at timestamp with time zone,
    created_at timestamp with time zone
);
CREATE SEQUENCE api_tokens_id_seq
    START WITH 1
    INCREMENT BY 1
    NO MINVALUE
    NO MAXVALUE
    CACHE 1;
ALTER SEQUENCE api_tokens_id_seq OWNED BY api_tokens.id;
CREATE TABLE audit_logs (
    id bigint NOT NULL,
    project_id bigint,
    user_id text,
    user_email text,
    action text,
    resource text,
    detail text,
    client_ip text,
    created_at timestamp with time zone
);
CREATE SEQUENCE audit_logs_id_seq
    START WITH 1
    INCREMENT BY 1
    NO MINVALUE
    NO MAXVALUE
    CACHE 1;
ALTER SEQUENCE audit_logs_id_seq OWNED BY audit_logs.id;
CREATE TABLE environments (
    id bigint NOT NULL,
    project_id bigint,
    name text,
    slug text,
    "position" bigint,
    protected boolean DEFAULT false,
    created_at timestamp with time zone,
    updated_at timestamp with time zone
);
CREATE SEQUENCE environments_id_seq
    START WITH 1
    INCREMENT BY 1
    NO MINVALUE
    NO MAXVALUE
    CACHE 1;
ALTER SEQUENCE environments_id_seq OWNED BY environments.id;
CREATE TABLE invites (
    id bigint NOT NULL,
    token text,
    project_id bigint,
    role text,
    max_uses bigint DEFAULT 0,
    use_count bigint DEFAULT 0,
    expires_at timestamp with time zone,
    created_by bigint,
    created_at timestamp with time zone
);
CREATE SEQUENCE invites_id_seq
    START WITH 1
    INCREMENT BY 1
    NO MINVALUE
    NO MAXVALUE
    CACHE 1;
ALTER SEQUENCE invites_id_seq OWNED BY invites.id;
CREATE TABLE password_resets (
    token_hash text NOT NULL,
    user_id bigint,
    expires_at timestamp with time zone,
    used boolean DEFAULT false,
    created_at timestamp with time zone
);
CREATE TABLE porte_identities (
    user_id bigint NOT NULL,
    provider text NOT NULL,
    subject text NOT NULL,
    password_hash text DEFAULT ''::text NOT NULL,
    access_token text DEFAULT ''::text NOT NULL,
    refresh_token text DEFAULT ''::text NOT NULL,
    token_expiry timestamp with time zone,
    roles jsonb,
    roles_synced_at timestamp with time zone,
    synced_at timestamp with time zone,
    created_at timestamp with time zone DEFAULT now()
);
CREATE TABLE porte_login_codes (
    code_hash text NOT NULL,
    user_id bigint NOT NULL,
    expires_at timestamp with time zone NOT NULL,
    consumed_at timestamp with time zone
);
CREATE TABLE porte_sessions (
    id bigint NOT NULL,
    token_hash text NOT NULL,
    user_id bigint NOT NULL,
    label text DEFAULT ''::text NOT NULL,
    created_at timestamp with time zone DEFAULT now() NOT NULL,
    last_used_at timestamp with time zone DEFAULT now() NOT NULL,
    expires_at timestamp with time zone
);
CREATE SEQUENCE porte_sessions_id_seq
    START WITH 1
    INCREMENT BY 1
    NO MINVALUE
    NO MAXVALUE
    CACHE 1;
ALTER SEQUENCE porte_sessions_id_seq OWNED BY porte_sessions.id;
CREATE TABLE project_members (
    id bigint NOT NULL,
    project_id bigint,
    user_id bigint,
    role text,
    created_at timestamp with time zone
);
CREATE SEQUENCE project_members_id_seq
    START WITH 1
    INCREMENT BY 1
    NO MINVALUE
    NO MAXVALUE
    CACHE 1;
ALTER SEQUENCE project_members_id_seq OWNED BY project_members.id;
CREATE TABLE projects (
    id bigint NOT NULL,
    name text,
    slug text,
    description text,
    encrypted_data_key bytea,
    data_key_nonce bytea,
    rotation_days bigint DEFAULT 90,
    created_at timestamp with time zone,
    updated_at timestamp with time zone
);
CREATE SEQUENCE projects_id_seq
    START WITH 1
    INCREMENT BY 1
    NO MINVALUE
    NO MAXVALUE
    CACHE 1;
ALTER SEQUENCE projects_id_seq OWNED BY projects.id;
CREATE TABLE secret_acls (
    id bigint NOT NULL,
    secret_id bigint,
    user_id bigint,
    can_read boolean DEFAULT true,
    can_write boolean DEFAULT false,
    created_at timestamp with time zone
);
CREATE SEQUENCE secret_acls_id_seq
    START WITH 1
    INCREMENT BY 1
    NO MINVALUE
    NO MAXVALUE
    CACHE 1;
ALTER SEQUENCE secret_acls_id_seq OWNED BY secret_acls.id;
CREATE TABLE secret_versions (
    id bigint NOT NULL,
    secret_id bigint,
    version bigint,
    encrypted_value bytea,
    nonce bytea,
    changed_by text,
    created_at timestamp with time zone
);
CREATE SEQUENCE secret_versions_id_seq
    START WITH 1
    INCREMENT BY 1
    NO MINVALUE
    NO MAXVALUE
    CACHE 1;
ALTER SEQUENCE secret_versions_id_seq OWNED BY secret_versions.id;
CREATE TABLE secrets (
    id bigint NOT NULL,
    environment_id bigint,
    key text,
    encrypted_value bytea,
    nonce bytea,
    version bigint DEFAULT 1,
    tags text DEFAULT ''::text,
    created_at timestamp with time zone,
    updated_at timestamp with time zone,
    deleted_at timestamp with time zone,
    deleted_by text DEFAULT ''::text
);
CREATE SEQUENCE secrets_id_seq
    START WITH 1
    INCREMENT BY 1
    NO MINVALUE
    NO MAXVALUE
    CACHE 1;
ALTER SEQUENCE secrets_id_seq OWNED BY secrets.id;
CREATE TABLE users (
    id bigint NOT NULL,
    email text,
    name text,
    created_at timestamp with time zone,
    updated_at timestamp with time zone
);
CREATE SEQUENCE users_id_seq
    START WITH 1
    INCREMENT BY 1
    NO MINVALUE
    NO MAXVALUE
    CACHE 1;
ALTER SEQUENCE users_id_seq OWNED BY users.id;
CREATE TABLE webhooks (
    id bigint NOT NULL,
    project_id bigint,
    url text,
    events text,
    secret text,
    active boolean DEFAULT true,
    created_by bigint,
    created_at timestamp with time zone,
    updated_at timestamp with time zone
);
CREATE SEQUENCE webhooks_id_seq
    START WITH 1
    INCREMENT BY 1
    NO MINVALUE
    NO MAXVALUE
    CACHE 1;
ALTER SEQUENCE webhooks_id_seq OWNED BY webhooks.id;
ALTER TABLE ONLY api_tokens ALTER COLUMN id SET DEFAULT nextval('api_tokens_id_seq'::regclass);
ALTER TABLE ONLY audit_logs ALTER COLUMN id SET DEFAULT nextval('audit_logs_id_seq'::regclass);
ALTER TABLE ONLY environments ALTER COLUMN id SET DEFAULT nextval('environments_id_seq'::regclass);
ALTER TABLE ONLY invites ALTER COLUMN id SET DEFAULT nextval('invites_id_seq'::regclass);
ALTER TABLE ONLY porte_sessions ALTER COLUMN id SET DEFAULT nextval('porte_sessions_id_seq'::regclass);
ALTER TABLE ONLY project_members ALTER COLUMN id SET DEFAULT nextval('project_members_id_seq'::regclass);
ALTER TABLE ONLY projects ALTER COLUMN id SET DEFAULT nextval('projects_id_seq'::regclass);
ALTER TABLE ONLY secret_acls ALTER COLUMN id SET DEFAULT nextval('secret_acls_id_seq'::regclass);
ALTER TABLE ONLY secret_versions ALTER COLUMN id SET DEFAULT nextval('secret_versions_id_seq'::regclass);
ALTER TABLE ONLY secrets ALTER COLUMN id SET DEFAULT nextval('secrets_id_seq'::regclass);
ALTER TABLE ONLY users ALTER COLUMN id SET DEFAULT nextval('users_id_seq'::regclass);
ALTER TABLE ONLY webhooks ALTER COLUMN id SET DEFAULT nextval('webhooks_id_seq'::regclass);
ALTER TABLE ONLY api_token_environments
    ADD CONSTRAINT api_token_environments_pkey PRIMARY KEY (token_id, environment_id);
ALTER TABLE ONLY api_tokens
    ADD CONSTRAINT api_tokens_pkey PRIMARY KEY (id);
ALTER TABLE ONLY audit_logs
    ADD CONSTRAINT audit_logs_pkey PRIMARY KEY (id);
ALTER TABLE ONLY environments
    ADD CONSTRAINT environments_pkey PRIMARY KEY (id);
ALTER TABLE ONLY invites
    ADD CONSTRAINT invites_pkey PRIMARY KEY (id);
ALTER TABLE ONLY password_resets
    ADD CONSTRAINT password_resets_pkey PRIMARY KEY (token_hash);
ALTER TABLE ONLY porte_identities
    ADD CONSTRAINT porte_identities_pkey PRIMARY KEY (provider, subject);
ALTER TABLE ONLY porte_login_codes
    ADD CONSTRAINT porte_login_codes_pkey PRIMARY KEY (code_hash);
ALTER TABLE ONLY porte_sessions
    ADD CONSTRAINT porte_sessions_pkey PRIMARY KEY (id);
ALTER TABLE ONLY porte_sessions
    ADD CONSTRAINT porte_sessions_token_hash_key UNIQUE (token_hash);
ALTER TABLE ONLY project_members
    ADD CONSTRAINT project_members_pkey PRIMARY KEY (id);
ALTER TABLE ONLY projects
    ADD CONSTRAINT projects_pkey PRIMARY KEY (id);
ALTER TABLE ONLY secret_acls
    ADD CONSTRAINT secret_acls_pkey PRIMARY KEY (id);
ALTER TABLE ONLY secret_versions
    ADD CONSTRAINT secret_versions_pkey PRIMARY KEY (id);
ALTER TABLE ONLY secrets
    ADD CONSTRAINT secrets_pkey PRIMARY KEY (id);
ALTER TABLE ONLY users
    ADD CONSTRAINT users_pkey PRIMARY KEY (id);
ALTER TABLE ONLY webhooks
    ADD CONSTRAINT webhooks_pkey PRIMARY KEY (id);
CREATE INDEX idx_api_tokens_project_id ON api_tokens USING btree (project_id);
CREATE INDEX idx_api_tokens_revoked_at ON api_tokens USING btree (revoked_at);
CREATE UNIQUE INDEX idx_api_tokens_token_hash ON api_tokens USING btree (token_hash);
CREATE INDEX idx_audit_logs_created_at ON audit_logs USING btree (created_at);
CREATE INDEX idx_audit_logs_project_id ON audit_logs USING btree (project_id);
CREATE UNIQUE INDEX idx_env_key ON secrets USING btree (environment_id, key);
CREATE INDEX idx_invites_project_id ON invites USING btree (project_id);
CREATE UNIQUE INDEX idx_invites_token ON invites USING btree (token);
CREATE INDEX idx_password_resets_user_id ON password_resets USING btree (user_id);
CREATE UNIQUE INDEX idx_project_env_slug ON environments USING btree (project_id, slug);
CREATE UNIQUE INDEX idx_project_user ON project_members USING btree (project_id, user_id);
CREATE UNIQUE INDEX idx_projects_slug ON projects USING btree (slug);
CREATE UNIQUE INDEX idx_secret_user ON secret_acls USING btree (secret_id, user_id);
CREATE INDEX idx_secret_versions_secret_id ON secret_versions USING btree (secret_id);
CREATE INDEX idx_secrets_deleted_at ON secrets USING btree (deleted_at);
CREATE UNIQUE INDEX idx_users_email ON users USING btree (email);
CREATE INDEX idx_webhooks_project_id ON webhooks USING btree (project_id);
ALTER TABLE ONLY api_token_environments
    ADD CONSTRAINT fk_api_token_environments_environment FOREIGN KEY (environment_id) REFERENCES environments(id) ON DELETE CASCADE;
ALTER TABLE ONLY api_token_environments
    ADD CONSTRAINT fk_api_token_environments_token FOREIGN KEY (token_id) REFERENCES api_tokens(id) ON DELETE CASCADE;
ALTER TABLE ONLY api_tokens
    ADD CONSTRAINT fk_api_tokens_project FOREIGN KEY (project_id) REFERENCES projects(id) ON DELETE CASCADE;
ALTER TABLE ONLY environments
    ADD CONSTRAINT fk_environments_project FOREIGN KEY (project_id) REFERENCES projects(id) ON DELETE CASCADE;
ALTER TABLE ONLY invites
    ADD CONSTRAINT fk_invites_project FOREIGN KEY (project_id) REFERENCES projects(id) ON DELETE CASCADE;
ALTER TABLE ONLY project_members
    ADD CONSTRAINT fk_project_members_project FOREIGN KEY (project_id) REFERENCES projects(id) ON DELETE CASCADE;
ALTER TABLE ONLY secret_acls
    ADD CONSTRAINT fk_secret_acls_secret FOREIGN KEY (secret_id) REFERENCES secrets(id) ON DELETE CASCADE;
ALTER TABLE ONLY secret_versions
    ADD CONSTRAINT fk_secret_versions_secret FOREIGN KEY (secret_id) REFERENCES secrets(id) ON DELETE CASCADE;
ALTER TABLE ONLY secrets
    ADD CONSTRAINT fk_secrets_environment FOREIGN KEY (environment_id) REFERENCES environments(id) ON DELETE CASCADE;
ALTER TABLE ONLY webhooks
    ADD CONSTRAINT fk_webhooks_project FOREIGN KEY (project_id) REFERENCES projects(id) ON DELETE CASCADE;
ALTER TABLE ONLY porte_identities
    ADD CONSTRAINT porte_identities_user_id_fkey FOREIGN KEY (user_id) REFERENCES users(id) ON DELETE CASCADE;
ALTER TABLE ONLY porte_login_codes
    ADD CONSTRAINT porte_login_codes_user_id_fkey FOREIGN KEY (user_id) REFERENCES users(id) ON DELETE CASCADE;
ALTER TABLE ONLY porte_sessions
    ADD CONSTRAINT porte_sessions_user_id_fkey FOREIGN KEY (user_id) REFERENCES users(id) ON DELETE CASCADE;

-- +goose Down
DROP TABLE IF EXISTS api_token_environments CASCADE;
DROP TABLE IF EXISTS secrets CASCADE;
DROP TABLE IF EXISTS secret_versions CASCADE;
DROP TABLE IF EXISTS secret_acls CASCADE;
DROP TABLE IF EXISTS password_resets CASCADE;
DROP TABLE IF EXISTS webhooks CASCADE;
DROP TABLE IF EXISTS invites CASCADE;
DROP TABLE IF EXISTS audit_logs CASCADE;
DROP TABLE IF EXISTS project_members CASCADE;
DROP TABLE IF EXISTS porte_sessions CASCADE;
DROP TABLE IF EXISTS porte_identities CASCADE;
DROP TABLE IF EXISTS porte_login_codes CASCADE;
DROP TABLE IF EXISTS environments CASCADE;
DROP TABLE IF EXISTS api_tokens CASCADE;
DROP TABLE IF EXISTS projects CASCADE;
DROP TABLE IF EXISTS users CASCADE;
