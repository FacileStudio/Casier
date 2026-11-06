import type { Catalogs, Message } from '@facile/muse';
import { members } from './members';
import { projects } from './projects';
import { psettings } from './psettings';
import { secrets } from './secrets';
import { settings } from './settings';
import { shell } from './shell';
import type { Row } from '../row';

const rows = { ...shell, ...settings, ...projects, ...members, ...psettings, ...secrets };

export type AppKey = keyof typeof rows;
type AppCatalog = Record<AppKey, Message>;

function column(index: number): AppCatalog {
	const entries = Object.entries(rows as Record<string, Row>).map(([key, row]) => [key, row[index]]);
	return Object.fromEntries(entries) as AppCatalog;
}

/** The app's catalogs: English is the typed base, and every row carries all four locales. */
export const catalogs: Catalogs<AppCatalog> = {
	en: column(0),
	fr: column(1),
	es: column(2),
	de: column(3)
};
