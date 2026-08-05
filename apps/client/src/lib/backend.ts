export const API_BASE = '/api';

export const COOKIE_SESSION = 'cookie';

type FetchOptions = {
	method?: string;
	body?: unknown;
	token?: string | null;
};

export class ApiError extends Error {
	status: number;
	code: string;

	constructor(message: string, status: number, code = '') {
		super(message);
		this.name = 'ApiError';
		this.status = status;
		this.code = code;
	}
}

function readString(source: Record<string, unknown>, key: string): string {
	const value = source[key];
	return typeof value === 'string' ? value : '';
}

function parseError(body: string, status: number): ApiError {
	let message = '';
	let code = '';

	try {
		const parsed: unknown = JSON.parse(body);
		if (parsed && typeof parsed === 'object') {
			const payload = parsed as Record<string, unknown>;
			const nested = payload.error;
			if (nested && typeof nested === 'object') {
				const inner = nested as Record<string, unknown>;
				message = readString(inner, 'message');
				code = readString(inner, 'code');
			} else if (typeof nested === 'string') {
				message = nested;
			}
			if (!message) message = readString(payload, 'message');
			if (!code) code = readString(payload, 'code');
		}
	} catch {
		message = '';
	}

	return new ApiError(message || `HTTP ${status}`, status, code);
}

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
		const body = await response.text().catch(() => '');
		throw parseError(body, response.status);
	}

	if (response.status === 204) return undefined as T;

	const body = await response.text().catch(() => '');
	if (!body) return undefined as T;

	try {
		return JSON.parse(body) as T;
	} catch {
		throw new ApiError('Unexpected response from server', response.status, 'invalid_response');
	}
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
	protected: boolean;
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
	revoked_at: string | null;
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
