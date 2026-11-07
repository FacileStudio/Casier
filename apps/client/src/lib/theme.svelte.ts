import { browser } from '$app/environment';
import { applyTheme } from '@facile/muse';

export type ThemePreference = 'light' | 'dark' | 'system';

const STORAGE_KEY = 'casier.theme';

function isPreference(value: string | null): value is ThemePreference {
	return value === 'light' || value === 'dark' || value === 'system';
}

function stored(): string | null {
	try {
		return localStorage.getItem(STORAGE_KEY);
	} catch {
		return null;
	}
}

class ThemeState {
	preference = $state<ThemePreference>('system');

	restore() {
		if (!browser) return;
		const value = stored();
		this.preference = isPreference(value) ? value : 'system';
		applyTheme(this.preference);
	}

	set(next: ThemePreference) {
		this.preference = next;
		if (!browser) return;
		try {
			localStorage.setItem(STORAGE_KEY, next);
		} catch {
			return;
		} finally {
			applyTheme(next);
		}
	}
}

export const theme = new ThemeState();
