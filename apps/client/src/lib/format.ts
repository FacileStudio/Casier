const GO_ZERO_TIME = '0001-01-01T00:00:00Z';

export function isZeroTime(value: string): boolean {
	return !value || value === GO_ZERO_TIME;
}

export function formatDate(value: string): string {
	if (isZeroTime(value)) return '—';
	return new Intl.DateTimeFormat(undefined, { dateStyle: 'medium' }).format(new Date(value));
}

export function daysSince(value: string): number | null {
	if (isZeroTime(value)) return null;
	const elapsed = Date.now() - new Date(value).getTime();
	if (Number.isNaN(elapsed)) return null;
	return Math.max(0, Math.floor(elapsed / 86_400_000));
}

export function formatDateTime(value: string): string {
	if (isZeroTime(value)) return '—';
	return new Intl.DateTimeFormat(undefined, { dateStyle: 'medium', timeStyle: 'short' }).format(
		new Date(value)
	);
}
