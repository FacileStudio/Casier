export const API_BASE = '/api';

export const COOKIE_SESSION = 'cookie';

type FetchOptions = {
	method?: string;
	body?: unknown;
	token?: string | null;
};

async function request<T>(path: string, options: FetchOptions = {}): Promise<T> {
	const headers: Record<string, string> = {};

	if (options.body) {
		headers['Content-Type'] = 'application/json';
	}

	if (options.token && options.token !== COOKIE_SESSION) {
		headers['Authorization'] = `Bearer ${options.token}`;
	}

	const response = await fetch(`${API_BASE}${path}`, {
		method: options.method ?? 'GET',
		headers,
		credentials: 'include',
		body: options.body ? JSON.stringify(options.body) : undefined,
	});

	if (!response.ok) {
		const error = await response.json().catch(() => ({ message: 'Request failed' }));
		throw new Error(error.message ?? `HTTP ${response.status}`);
	}

	if (response.status === 204) return undefined as T;
	return response.json();
}

export const api = {
	get: <T>(path: string, token?: string | null) => request<T>(path, { token }),
	post: <T>(path: string, body: unknown, token?: string | null) => request<T>(path, { method: 'POST', body, token }),
	put: <T>(path: string, body: unknown, token?: string | null) => request<T>(path, { method: 'PUT', body, token }),
	del: <T>(path: string, token?: string | null) => request<T>(path, { method: 'DELETE', token }),
};

export type AuthConfig = {
	sso_only: boolean;
	oidc_enabled: boolean;
};

export type AuthResponse = {
	user_id: string;
	token: string;
};

export type Me = {
	id: string;
	email: string;
};

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
	created_at: string;
	updated_at: string;
};

export type EnvironmentsResponse = {
	environments: Environment[] | null;
};

export type RotationStatus = '' | 'ok' | 'due_soon' | 'overdue';

export type Secret = {
	id: number;
	key: string;
	value: string;
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

export type Member = {
	id: number;
	email: string;
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
	created_at: string;
};

export type ApiTokensResponse = {
	tokens: ApiToken[] | null;
};

export type AuditLogEntry = {
	id: number;
	user_email: string;
	action: string;
	resource: string;
	detail: string;
	client_ip: string;
	created_at: string;
};

export type AuditLogsResponse = {
	logs: AuditLogEntry[] | null;
};
