import { browser } from '$app/environment';
import { api } from './backend';

class AuthState {
	userId = $state<string | null>(null);
	email = $state<string | null>(null);
	loaded = $state(false);

	get isAuthenticated() {
		return this.userId !== null;
	}

	async check() {
		if (!browser) return;
		try {
			const me = await api.get<{ id: string; email: string }>('/auth/me');
			this.userId = me.id;
			this.email = me.email;
		} catch {
			this.userId = null;
			this.email = null;
		}
		this.loaded = true;
	}

	logout() {
		this.userId = null;
		this.email = null;
	}
}

export const auth = new AuthState();
