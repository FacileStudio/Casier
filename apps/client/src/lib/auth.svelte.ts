import { browser } from '$app/environment';

class AuthState {
	token = $state<string | null>(null);
	userId = $state<string | null>(null);
	email = $state<string | null>(null);

	get isAuthenticated() {
		return this.token !== null;
	}

	login(token: string, userId: string, email: string) {
		this.token = token;
		this.userId = userId;
		this.email = email;
		if (browser) {
			sessionStorage.setItem('clef_token', token);
			sessionStorage.setItem('clef_user_id', userId);
			sessionStorage.setItem('clef_email', email);
		}
	}

	logout() {
		this.token = null;
		this.userId = null;
		this.email = null;
		if (browser) {
			sessionStorage.removeItem('clef_token');
			sessionStorage.removeItem('clef_user_id');
			sessionStorage.removeItem('clef_email');
		}
	}

	restore() {
		if (browser) {
			this.token = sessionStorage.getItem('clef_token');
			this.userId = sessionStorage.getItem('clef_user_id');
			this.email = sessionStorage.getItem('clef_email');
		}
	}
}

export const auth = new AuthState();
