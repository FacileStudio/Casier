const GO_ZERO_TIME = '0001-01-01T00:00:00Z';

export function isZeroTime(value: string): boolean {
	return !value || value === GO_ZERO_TIME;
}

export function formatDate(value: string): string {
	if (isZeroTime(value)) return '—';
	return new Intl.DateTimeFormat(undefined, { dateStyle: 'medium' }).format(new Date(value));
}

export function formatDateTime(value: string): string {
	if (isZeroTime(value)) return '—';
	return new Intl.DateTimeFormat(undefined, { dateStyle: 'medium', timeStyle: 'short' }).format(
		new Date(value)
	);
}
