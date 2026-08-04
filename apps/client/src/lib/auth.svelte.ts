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
			sessionStorage.setItem('casier_token', token);
			sessionStorage.setItem('casier_user_id', userId);
			sessionStorage.setItem('casier_email', email);
		}
	}

	logout() {
		this.token = null;
		this.userId = null;
		this.email = null;
		if (browser) {
			sessionStorage.removeItem('casier_token');
			sessionStorage.removeItem('casier_user_id');
			sessionStorage.removeItem('casier_email');
		}
	}

	restore() {
		if (browser) {
			this.token = sessionStorage.getItem('casier_token');
			this.userId = sessionStorage.getItem('casier_user_id');
			this.email = sessionStorage.getItem('casier_email');
		}
	}
}

export const auth = new AuthState();
