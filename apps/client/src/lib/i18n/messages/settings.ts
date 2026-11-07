import type { Row } from '../row';

/** The account settings pages: profile, appearance, API and advanced. */
export const settings = {
	'settings.subtitle': [
		'Your account, this browser, and how machines reach this instance.',
		'Votre compte, ce navigateur et la façon dont les machines accèdent à cette instance.',
		'Tu cuenta, este navegador y cómo acceden las máquinas a esta instancia.',
		'Dein Konto, dieser Browser und der Zugriff von Maschinen auf diese Instanz.'
	],
	'settings.tab.profile': ['Profile', 'Profil', 'Perfil', 'Profil'],
	'settings.tab.preferences': ['Preferences', 'Préférences', 'Preferencias', 'Präferenzen'],
	'settings.tab.tokens': ['API tokens', 'Jetons d’API', 'Tokens de API', 'API-Tokens'],
	'settings.tab.advanced': ['Advanced', 'Avancé', 'Avanzado', 'Erweitert'],
	'settings.head.profile': ['Settings — Casier', 'Paramètres — Casier', 'Ajustes — Casier', 'Einstellungen — Casier'],
	'settings.head.preferences': ['Preferences — Casier', 'Préférences — Casier', 'Preferencias — Casier', 'Präferenzen — Casier'],
	'settings.head.tokens': ['API tokens — Casier', 'Jetons d’API — Casier', 'Tokens de API — Casier', 'API-Tokens — Casier'],
	'settings.head.advanced': ['Advanced — Casier', 'Avancé — Casier', 'Avanzado — Casier', 'Erweitert — Casier'],
	'profile.interface': ['Interface', 'Interface', 'Interfaz', 'Oberfläche'],
	'profile.auth.title': ['Authentication', 'Authentification', 'Autenticación', 'Authentifizierung'],
	'profile.auth.description': [
		'How you get into Casier. It is set by the server, not by you.',
		'Comment vous accédez à Casier. C’est le serveur qui le définit, pas vous.',
		'Cómo accedes a Casier. Lo define el servidor, no tú.',
		'Wie du in Casier gelangst. Das legt der Server fest, nicht du.'
	],
	'profile.auth.label': ['Facile sign-in', 'Connexion Facile', 'Inicio de sesión con Facile', 'Facile-Anmeldung'],
	'profile.auth.rowDescription': [
		'Casier has no password of its own. Your Facile account owns your credentials — change them there.',
		'Casier n’a pas de mot de passe propre. Votre compte Facile détient vos identifiants : modifiez-les là-bas.',
		'Casier no tiene contraseña propia. Tu cuenta de Facile gestiona tus credenciales: cámbialas allí.',
		'Casier hat kein eigenes Passwort. Dein Facile-Konto verwaltet deine Zugangsdaten — ändere sie dort.'
	],
	'profile.auth.badge': ['Managed by Facile', 'Géré par Facile', 'Gestionado por Facile', 'Von Facile verwaltet'],
	'profile.logout.warning': [
		'The server did not confirm the sign-out, ending the session locally anyway.',
		'Le serveur n’a pas confirmé la déconnexion, la session est quand même terminée localement.',
		'El servidor no confirmó el cierre de sesión; se termina la sesión localmente de todos modos.',
		'Der Server hat die Abmeldung nicht bestätigt; die Sitzung wird trotzdem lokal beendet.'
	],
	'api.endpoint.title': ['Endpoint', 'Point d’accès', 'Endpoint', 'Endpunkt'],
	'api.endpoint.description': [
		'Point the CLI or any HTTP client here. The URL is not a secret — the token is.',
		'Pointez la CLI ou tout client HTTP ici. L’URL n’est pas un secret, le jeton l’est.',
		'Apunta aquí la CLI o cualquier cliente HTTP. La URL no es un secreto; el token sí.',
		'Richte die CLI oder jeden HTTP-Client hierhin. Die URL ist kein Geheimnis — das Token schon.'
	],
	'api.baseUrl.label': ['Base URL', 'URL de base', 'URL base', 'Basis-URL'],
	'api.baseUrl.description': [
		'Behind a reverse proxy the API answers under /api; in development it answers at the root.',
		'Derrière un reverse proxy, l’API répond sous /api ; en développement, elle répond à la racine.',
		'Detrás de un proxy inverso la API responde bajo /api; en desarrollo responde en la raíz.',
		'Hinter einem Reverse-Proxy antwortet die API unter /api; in der Entwicklung an der Wurzel.'
	],
	'api.cli.title': ['Command line', 'Ligne de commande', 'Línea de comandos', 'Kommandozeile'],
	'api.cli.description': [
		'The casier CLI pulls the same secrets into your shell, your scripts and your CI.',
		'La CLI casier charge les mêmes secrets dans votre shell, vos scripts et votre CI.',
		'La CLI casier trae los mismos secretos a tu shell, tus scripts y tu CI.',
		'Die casier-CLI lädt dieselben Secrets in deine Shell, deine Skripte und deine CI.'
	],
	'api.cli.login.label': ['Sign in', 'Se connecter', 'Iniciar sesión', 'Anmelden'],
	'api.cli.login.description': [
		'Discovers this instance through /auth/config and stores a session token.',
		'Découvre cette instance via /auth/config et enregistre un jeton de session.',
		'Descubre esta instancia mediante /auth/config y guarda un token de sesión.',
		'Erkennt diese Instanz über /auth/config und speichert ein Sitzungstoken.'
	],
	'api.cli.init.label': ['Link a directory', 'Lier un dossier', 'Vincular un directorio', 'Verzeichnis verknüpfen'],
	'api.cli.init.description': [
		'Writes the project and environment this directory reads from.',
		'Écrit le projet et l’environnement que ce dossier lit.',
		'Escribe el proyecto y el entorno de los que lee este directorio.',
		'Schreibt das Projekt und die Umgebung, aus denen dieses Verzeichnis liest.'
	],
	'api.cli.run.label': ['Run with secrets', 'Exécuter avec les secrets', 'Ejecutar con secretos', 'Mit Secrets ausführen'],
	'api.cli.run.description': [
		'Injects the environment for one process. Nothing is written to disk.',
		'Injecte l’environnement pour un seul processus. Rien n’est écrit sur le disque.',
		'Inyecta el entorno para un solo proceso. No se escribe nada en disco.',
		'Übergibt die Umgebung an einen Prozess. Es wird nichts auf die Festplatte geschrieben.'
	],
	'api.tokens.title': ['Tokens', 'Jetons', 'Tokens', 'Tokens'],
	'api.tokens.description': [
		'Casier scopes a machine credential to one project, so tokens are issued there rather than here.',
		'Casier limite un identifiant machine à un projet : les jetons sont donc émis là-bas plutôt qu’ici.',
		'Casier acota una credencial de máquina a un proyecto, así que los tokens se emiten allí y no aquí.',
		'Casier beschränkt Maschinenzugänge auf ein Projekt, daher werden Tokens dort ausgestellt, nicht hier.'
	],
	'api.tokens.projects': ['Projects', 'Projets', 'Proyectos', 'Projekte'],
	'api.tokens.where.label': ['Where tokens live', 'Où vivent les jetons', 'Dónde están los tokens', 'Wo Tokens liegen'],
	'api.tokens.where.description': [
		'Open a project, then API tokens. A token is bound to that project, to a permission set and to the environments you name — it can never read past them.',
		'Ouvrez un projet, puis « Jetons d’API ». Un jeton est lié à ce projet, à un ensemble de permissions et aux environnements que vous nommez : il ne peut jamais lire au-delà.',
		'Abre un proyecto y luego «Tokens de API». Un token queda ligado a ese proyecto, a un conjunto de permisos y a los entornos que indiques: nunca puede leer más allá.',
		'Öffne ein Projekt und dann „API-Tokens“. Ein Token ist an dieses Projekt, einen Berechtigungssatz und die genannten Umgebungen gebunden — es kann nie darüber hinaus lesen.'
	],
	'api.tokens.issue': ['Issue a token', 'Émettre un jeton', 'Emitir un token', 'Token ausstellen'],
	'api.tokens.cli.label': ['How the CLI uses one', 'Comment la CLI en utilise un', 'Cómo lo usa la CLI', 'Wie die CLI es nutzt'],
	'api.tokens.cli.description': [
		'Non-interactive callers send it as CASIER_TOKEN. Tokens expire — 90 days by default, 730 at most — and are revoked rather than deleted so the audit log still names them.',
		'Les appelants non interactifs l’envoient sous CASIER_TOKEN. Les jetons expirent (90 jours par défaut, 730 au maximum) et sont révoqués plutôt que supprimés, pour que le journal d’audit les nomme toujours.',
		'Los llamadores no interactivos lo envían como CASIER_TOKEN. Los tokens caducan (90 días por defecto, 730 como máximo) y se revocan en lugar de borrarse, para que el registro de auditoría los siga nombrando.',
		'Nicht interaktive Aufrufer senden es als CASIER_TOKEN. Tokens laufen ab — standardmäßig nach 90 Tagen, höchstens nach 730 — und werden widerrufen statt gelöscht, damit das Audit-Log sie weiter nennt.'
	],
	'advanced.instance.title': ['Instance', 'Instance', 'Instancia', 'Instanz'],
	'advanced.instance.description': [
		'Worth quoting when you file a bug against a self-hosted install.',
		'À citer lorsque vous signalez un bug sur une installation auto-hébergée.',
		'Conviene citarlo al informar de un fallo en una instalación autoalojada.',
		'Gut zu nennen, wenn du einen Fehler zu einer selbst gehosteten Installation meldest.'
	],
	'advanced.interface.description': [
		'This app is the browser client for the Casier API.',
		'Cette application est le client navigateur de l’API Casier.',
		'Esta aplicación es el cliente de navegador de la API de Casier.',
		'Diese App ist der Browser-Client für die Casier-API.'
	],
	'advanced.sso.label': ['Facile sign-in', 'Connexion Facile', 'Inicio de sesión con Facile', 'Facile-Anmeldung'],
	'advanced.sso.description': [
		'OIDC federation against the instance’s identity provider. Enabled by OIDC_ISSUER on the server.',
		'Fédération OIDC auprès du fournisseur d’identité de l’instance. Activée par OIDC_ISSUER sur le serveur.',
		'Federación OIDC con el proveedor de identidad de la instancia. Se activa con OIDC_ISSUER en el servidor.',
		'OIDC-Föderation mit dem Identitätsanbieter der Instanz. Aktiviert durch OIDC_ISSUER auf dem Server.'
	],
	'advanced.enabled': ['Enabled', 'Activé', 'Activado', 'Aktiviert'],
	'advanced.notConfigured': ['Not configured', 'Non configuré', 'No configurado', 'Nicht konfiguriert'],
	'advanced.checking': ['Checking…', 'Vérification…', 'Comprobando…', 'Wird geprüft…'],
	'advanced.storage.label': ['Secret storage', 'Stockage des secrets', 'Almacenamiento de secretos', 'Secret-Speicher'],
	'advanced.storage.description': [
		'Values are encrypted with a per-project data key and a fresh nonce per secret. The server holds a hash of every session and API token, never the token itself.',
		'Les valeurs sont chiffrées avec une clé de données par projet et un nonce neuf par secret. Le serveur conserve un hachage de chaque session et jeton d’API, jamais le jeton lui-même.',
		'Los valores se cifran con una clave de datos por proyecto y un nonce nuevo por secreto. El servidor guarda un hash de cada sesión y token de API, nunca el token.',
		'Werte werden mit einem Datenschlüssel pro Projekt und einem frischen Nonce pro Secret verschlüsselt. Der Server speichert einen Hash jeder Sitzung und jedes API-Tokens, nie das Token selbst.'
	],
	'advanced.storage.status': [
		'AES-256-GCM at rest',
		'AES-256-GCM au repos',
		'AES-256-GCM en reposo',
		'AES-256-GCM im Ruhezustand'
	],
	'advanced.about.title': ['About', 'À propos', 'Acerca de', 'Über'],
	'advanced.about.description': [
		'Casier is the secrets and environment manager of the Facile Suite.',
		'Casier est le gestionnaire de secrets et d’environnements de la Facile Suite.',
		'Casier es el gestor de secretos y entornos de la Facile Suite.',
		'Casier ist der Manager für Secrets und Umgebungen der Facile Suite.'
	],
	'advanced.about.what.label': ['What it does', 'Ce qu’il fait', 'Qué hace', 'Was es tut'],
	'advanced.about.what.description': [
		'Groups secrets into environments inside a project, encrypts every value at rest, versions each change and writes an audit row for it. The CLI reads the same secrets into a shell, a script or CI.',
		'Regroupe les secrets en environnements au sein d’un projet, chiffre chaque valeur au repos, versionne chaque modification et écrit une ligne d’audit. La CLI lit les mêmes secrets dans un shell, un script ou la CI.',
		'Agrupa los secretos en entornos dentro de un proyecto, cifra cada valor en reposo, versiona cada cambio y escribe una fila de auditoría. La CLI lee los mismos secretos en un shell, un script o CI.',
		'Gruppiert Secrets in Umgebungen innerhalb eines Projekts, verschlüsselt jeden Wert, versioniert jede Änderung und schreibt dazu einen Audit-Eintrag. Die CLI liest dieselben Secrets in eine Shell, ein Skript oder die CI.'
	],
	'advanced.about.open': ['Open projects', 'Ouvrir les projets', 'Abrir proyectos', 'Projekte öffnen']
} as const satisfies Record<string, Row>;
