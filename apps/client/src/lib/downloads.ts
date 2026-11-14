import type { CliCommand } from '@facile/muse';
import { t } from '$lib/i18n/index.svelte';

export const RELEASES_URL = 'https://github.com/FacileStudio/casier-cli/releases';
export const CHECKSUMS_URL = `${RELEASES_URL}/latest/download/checksums.txt`;

export const INSTALL_METHODS: CliCommand[] = [
	{
		get label() {
			return t('downloads.install.script');
		},
		command: 'curl -fsSL https://raw.githubusercontent.com/FacileStudio/casier-cli/main/install.sh | bash'
	},
	{
		get label() {
			return t('downloads.install.facile');
		},
		command: 'facile install casier'
	},
	{
		get label() {
			return t('downloads.install.brew');
		},
		command: 'brew install FacileStudio/tap/casier'
	}
];
