import { rememberLocale, resolveClientLocale } from '@facile/client/locale';
import { createTranslator, LOCALES, resolveLocale, type Locale } from '@facile/muse';
import { auth } from '$lib/auth.svelte';
import { catalogs } from './messages';

export { rememberLocale };

let active: () => Locale = () => 'en';

/** The locale source for `provideLocale`: the URL, the session, the picked cookie, then the browser. Null when nothing matched. */
export function localeSource(): Locale | null {
	return resolveClientLocale({ supported: LOCALES, profile: auth.locale }).hint;
}

/** Points the translator and the Intl helpers at the locale `provideLocale` resolved. */
export function bindLocale(source: () => Locale) {
	active = source;
}

/** The locale currently in force. */
export function locale(): Locale {
	return resolveLocale(active());
}

/** Translates an app message in the current locale. */
export const t = createTranslator(catalogs, () => active());
