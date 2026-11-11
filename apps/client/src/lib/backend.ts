import { ApiError, createApi, redirectToLogin, type CallOptions } from '@facile/client';

export { ApiError };

export const API_BASE = '/api';

export const COOKIE_SESSION = 'cookie';

const apiFetch = createApi({ base: API_BASE, onUnauthorized: redirectToLogin });

function request<T>(method: string, path: string, body?: unknown, token?: string | null, options?: CallOptions): Promise<T> {
	const headers: Record<string, string> = {};
	if (token && token !== COOKIE_SESSION) headers['Authorization'] = `Bearer ${token}`;
	return apiFetch<T>(path, { method, headers, body: body ? JSON.stringify(body) : undefined }, options);
}

/**
 * The API, always relative to /api. A 401 sends the browser to the login page,
 * except on a call made with `{ quiet: true }`, such as the session probe.
 */
export const api = {
	get: <T>(path: string, token?: string | null, options?: CallOptions) => request<T>('GET', path, undefined, token, options),
	post: <T>(path: string, body: unknown, token?: string | null) => request<T>('POST', path, body, token),
	put: <T>(path: string, body: unknown, token?: string | null) => request<T>('PUT', path, body, token),
	del: <T>(path: string, token?: string | null) => request<T>('DELETE', path, undefined, token),
};

/** GET /me/invitations: Registre's pending space invitations for the user menu. */
export type Invitations = {
	count: number;
	href?: string;
};

export type AuthConfig = {
	sso_only: boolean;
	oidc_enabled: boolean;
};

/** The author name the API stores on rows written by someone whose account was erased. */
export const DELETED_USER = 'Deleted user';

export type Person = {
	email: string;
	name: string;
	avatar_url: string;
};

export type Me = Person & { id: string; locale: string };

export type ProjectRole = 'owner' | 'admin' | 'member' | 'viewer';

export type Project = {
	id: number;
	name: string;
	slug: string;
	description: string;
	role: ProjectRole;
	rotation_days: number;
	created_at: string;
	updated_at: string;
};

export type ProjectsResponse = {
	projects: Project[] | null;
};

export type Environment = {
	id: number;
	name: string;
	slug: string;
	position: number;
	protected: boolean;
	created_at: string;
	updated_at: string;
};

export type EnvironmentsResponse = {
	environments: Environment[] | null;
};

export type RotationStatus = '' | 'ok' | 'due_soon' | 'overdue';

// `value` is absent unless it was explicitly revealed. Listing no longer ships
// plaintext, so masking in the browser is no longer theatre over a value that
// was in the payload all along.
export type Secret = {
	id: number;
	key: string;
	value?: string;
	version: number;
	tags: string[] | null;
	created_at: string;
	updated_at: string;
	rotated_at: string;
	rotation_status: RotationStatus;
};

export type SecretsResponse = {
	secrets: Secret[] | null;
	rotation_days: number;
};

export type TrashedSecret = {
	id: number;
	key: string;
	tags: string[] | null;
	deleted_at: string;
	deleted_by: string;
	actor: Person | null;
	purge_at: string;
};

export type TrashResponse = {
	secrets: TrashedSecret[] | null;
	retention_days: number;
};

export type EnvironmentRotation = {
	slug: string;
	name: string;
	overdue: number;
	due_soon: number;
};

export type RotationSummary = {
	rotation_days: number;
	environments: EnvironmentRotation[] | null;
};

export type SecretVersion = {
	version: number;
	changed_by: string;
	actor: Person | null;
	created_at: string;
};

export type SecretVersionsResponse = {
	versions: SecretVersion[] | null;
};

export type ImportResult = {
	created: number;
	updated: number;
	skipped: number;
};

export type ExportResult = {
	content: string;
};

export type Member = Person & {
	id: number;
	role: ProjectRole;
};

export type MembersResponse = {
	members: Member[] | null;
};

export type ApiToken = {
	id: number;
	token?: string;
	name: string;
	environments: string;
	permissions: string;
	expires_at: string;
	last_used_at: string;
	revoked_at: string | null;
	created_at: string;
	owner: Person | null;
};

export type ApiTokensResponse = {
	tokens: ApiToken[] | null;
};

export type AuditLogEntry = {
	id: number;
	user_email: string;
	actor: Person | null;
	action: string;
	resource: string;
	detail: string;
	client_ip: string;
	created_at: string;
};

export type AuditLogsResponse = {
	logs: AuditLogEntry[] | null;
};

/** One secret a pending MCP action names, as it is now. Never a value. */
export type PendingSecret = {
	key: string;
	version: number;
	updated_at: string;
};

/** GET /mcp/pending: an action an AI client asked for, waiting for the person. */
export type PendingAction = {
	id: string;
	kind: 'reveal_secrets' | 'delete_secret';
	client_id: string;
	summary: string;
	created_at: string;
	expires_at: string;
	state: 'ready' | 'changed' | 'missing';
	project?: string;
	environment?: string;
	secrets?: PendingSecret[];
};

export type PendingList = {
	actions: PendingAction[];
};

/** The confirm page's answer: approved for a reveal, trashed for a delete, or dismissed. */
export type PendingResult = {
	status: 'approved' | 'trashed' | 'dismissed';
};
