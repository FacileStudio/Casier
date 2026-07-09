export const API_BASE = import.meta.env.VITE_API_BASE_URL ?? 'http://localhost:4000';

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

	if (options.token) {
		headers['Authorization'] = `Bearer ${options.token}`;
	}

	const response = await fetch(`${API_BASE}${path}`, {
		method: options.method ?? 'GET',
		headers,
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

export type Space = {
	id: string;
	name: string;
	slug: string;
	description: string;
	created_at: string;
	updated_at: string;
	member_count: number;
};

export type Secret = {
	id: string;
	space_id: string;
	environment: string;
	key: string;
	value: string;
	created_at: string;
	updated_at: string;
};

export type Member = {
	id: string;
	user_id: string;
	space_id: string;
	role: string;
	email: string;
	created_at: string;
};

export type User = {
	id: string;
	email: string;
	created_at: string;
};
