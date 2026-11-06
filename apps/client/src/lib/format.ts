import { locale } from '$lib/i18n/index.svelte';

const GO_ZERO_TIME = '0001-01-01T00:00:00Z';

export function isZeroTime(value: string): boolean {
	return !value || value === GO_ZERO_TIME;
}

export function formatDate(value: string): string {
	if (isZeroTime(value)) return '—';
	return new Intl.DateTimeFormat(locale(), { dateStyle: 'medium' }).format(new Date(value));
}

export function daysSince(value: string): number | null {
	if (isZeroTime(value)) return null;
	const elapsed = Date.now() - new Date(value).getTime();
	if (Number.isNaN(elapsed)) return null;
	return Math.max(0, Math.floor(elapsed / 86_400_000));
}

export function formatDateTime(value: string): string {
	if (isZeroTime(value)) return '—';
	return new Intl.DateTimeFormat(locale(), { dateStyle: 'medium', timeStyle: 'short' }).format(
		new Date(value)
	);
}

/**
 * The name a person is shown under: the display name the identity provider
 * sent, else the local part of their email, since a 220px rail cannot fit a
 * full address.
 */
export function displayName(person: { name?: string; email?: string } | null | undefined): string {
	return person?.name || person?.email?.split('@')[0] || '';
}
