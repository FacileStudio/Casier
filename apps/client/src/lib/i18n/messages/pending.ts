import type { Row } from '../row';

/** The page where the person confirms what an AI client asked Casier to do. */
export const pending = {
	'pending.title': ['Pending actions · Casier', 'Actions en attente · Casier', 'Acciones pendientes · Casier', 'Ausstehende Aktionen · Casier'],
	'pending.heading': ['Pending actions', 'Actions en attente', 'Acciones pendientes', 'Ausstehende Aktionen'],
	'pending.intro': [
		'An AI client asked Casier to do these. Nothing happens until you confirm.',
		'Un client IA a demandé ces actions à Casier. Rien ne se passe tant que vous ne confirmez pas.',
		'Un cliente de IA pidió estas acciones a Casier. No ocurre nada hasta que confirmes.',
		'Ein KI-Client hat Casier um diese Aktionen gebeten. Nichts passiert, bevor Sie bestätigen.'
	],
	'pending.none': ['Nothing is waiting', 'Rien en attente', 'Nada pendiente', 'Nichts ausstehend'],
	'pending.noneBody': [
		'Actions an AI client asks for appear here until you confirm or dismiss them.',
		'Les actions demandées par un client IA apparaissent ici jusqu’à ce que vous les confirmiez ou les ignoriez.',
		'Las acciones que pide un cliente de IA aparecen aquí hasta que las confirmes o descartes.',
		'Aktionen, um die ein KI-Client bittet, erscheinen hier, bis Sie sie bestätigen oder verwerfen.'
	],
	'pending.kind.reveal': ['Reveal values', 'Révéler des valeurs', 'Revelar valores', 'Werte freigeben'],
	'pending.kind.delete': ['Delete secret', 'Supprimer le secret', 'Eliminar secreto', 'Secret löschen'],
	'pending.askedBy': ['Asked by {client}', 'Demandé par {client}', 'Pedido por {client}', 'Angefragt von {client}'],
	'pending.version': ['v{version}', 'v{version}', 'v{version}', 'v{version}'],
	'pending.revealNote': [
		'The AI client that asked can then read these values once, within 15 minutes. They are not shown here.',
		'Le client IA qui a demandé pourra alors lire ces valeurs une fois, dans les 15 minutes. Elles ne sont pas affichées ici.',
		'El cliente de IA que lo pidió podrá leer estos valores una vez, en los próximos 15 minutos. No se muestran aquí.',
		'Der anfragende KI-Client kann diese Werte dann einmal innerhalb von 15 Minuten lesen. Sie werden hier nicht angezeigt.'
	],
	'pending.deleteNote': [
		'The secret moves to the trash and can be restored for 30 days.',
		'Le secret part dans la corbeille et peut être restauré pendant 30 jours.',
		'El secreto pasa a la papelera y se puede restaurar durante 30 días.',
		'Das Secret wandert in den Papierkorb und lässt sich 30 Tage lang wiederherstellen.'
	],
	'pending.confirmReveal': ['Allow reading', 'Autoriser la lecture', 'Permitir la lectura', 'Lesen erlauben'],
	'pending.confirmDelete': ['Delete the secret', 'Supprimer le secret', 'Eliminar el secreto', 'Secret löschen'],
	'pending.dismiss': ['Dismiss', 'Ignorer', 'Descartar', 'Verwerfen'],
	'pending.changed': [
		'A secret changed after the request. Ask the AI client again.',
		'Un secret a changé après la demande. Redemandez au client IA.',
		'Un secreto cambió después de la solicitud. Vuelve a pedírselo al cliente de IA.',
		'Ein Secret wurde nach der Anfrage geändert. Fragen Sie den KI-Client erneut.'
	],
	'pending.missing': [
		'A secret this names no longer exists or is out of your reach.',
		'Un secret visé n’existe plus ou n’est plus à votre portée.',
		'Un secreto indicado ya no existe o está fuera de tu alcance.',
		'Ein genanntes Secret existiert nicht mehr oder ist für Sie nicht erreichbar.'
	],
	'pending.approved': [
		'Approved. The AI client can read the values now.',
		'Autorisé. Le client IA peut maintenant lire les valeurs.',
		'Autorizado. El cliente de IA ya puede leer los valores.',
		'Erlaubt. Der KI-Client kann die Werte jetzt lesen.'
	],
	'pending.trashed': ['Secret moved to the trash', 'Secret placé dans la corbeille', 'Secreto movido a la papelera', 'Secret in den Papierkorb verschoben'],
	'pending.dismissed': ['Request dismissed', 'Demande ignorée', 'Solicitud descartada', 'Anfrage verworfen'],
	'pending.loadFailed': [
		'Could not load the pending actions.',
		'Impossible de charger les actions en attente.',
		'No se pudieron cargar las acciones pendientes.',
		'Die ausstehenden Aktionen konnten nicht geladen werden.'
	],
	'pending.signInAgain.title': [
		'Sign in again to confirm',
		'Reconnectez-vous pour confirmer',
		'Vuelve a iniciar sesión para confirmar',
		'Melden Sie sich erneut an, um zu bestätigen'
	],
	'pending.signInAgain.body': [
		'Only a fresh sign-in in your browser can confirm what an AI client asked. Sign in again here, then come back to this page. The mobile and desktop apps cannot confirm.',
		'Seule une connexion récente dans votre navigateur peut confirmer ce qu’un client IA a demandé. Reconnectez-vous ici, puis revenez sur cette page. Les applications mobile et de bureau ne peuvent pas confirmer.',
		'Solo un inicio de sesión reciente en tu navegador puede confirmar lo que pidió un cliente de IA. Vuelve a iniciar sesión aquí y regresa a esta página. Las aplicaciones móvil y de escritorio no pueden confirmar.',
		'Nur eine frische Anmeldung in Ihrem Browser kann bestätigen, worum ein KI-Client gebeten hat. Melden Sie sich hier erneut an und kehren Sie dann zu dieser Seite zurück. Die Mobil- und Desktop-Apps können nicht bestätigen.'
	],
	'pending.signInAgain.action': ['Sign in again', 'Se reconnecter', 'Iniciar sesión de nuevo', 'Erneut anmelden'],
	'pending.failed': [
		'That did not work. Try again.',
		'Cela n’a pas fonctionné. Réessayez.',
		'No funcionó. Inténtalo de nuevo.',
		'Das hat nicht geklappt. Versuchen Sie es erneut.'
	]
} as const satisfies Record<string, Row>;
