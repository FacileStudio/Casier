import { browser } from '$app/environment';
import { api } from './backend';

class AuthState {
	userId = $state<string | null>(null);
	email = $state<string | null>(null);
	loaded = $state(false);
	private checkPromise: Promise<void> | null = null;

	get isAuthenticated() {
		return this.userId !== null;
	}

	async check() {
		if (!browser) return;
		if (this.checkPromise) return this.checkPromise;
		this.checkPromise = (async () => {
			try {
				const me = await api.get<{ id: string; email: string }>('/auth/me');
				this.userId = me.id;
				this.email = me.email;
			} catch {
				this.userId = null;
				this.email = null;
			} finally {
				this.loaded = true;
				this.checkPromise = null;
			}
		})();
		return this.checkPromise;
	}

	logout() {
		this.userId = null;
		this.email = null;
		this.loaded = true;
		this.checkPromise = null;
	}
}

export const auth = new AuthState();
