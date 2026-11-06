import type { Row } from '../row';

/** The account settings pages: profile, appearance, API and advanced. */
export const settings = {
	'settings.title': ['Settings', 'Paramètres', 'Ajustes', 'Einstellungen'],
	'settings.subtitle': [
		'Your account, this browser, and how machines reach this instance.',
		'Votre compte, ce navigateur et la façon dont les machines accèdent à cette instance.',
		'Tu cuenta, este navegador y cómo acceden las máquinas a esta instancia.',
		'Dein Konto, dieser Browser und der Zugriff von Maschinen auf diese Instanz.'
	],
	'settings.sections': ['Settings sections', 'Sections des paramètres', 'Secciones de ajustes', 'Einstellungsbereiche'],
	'settings.tab.profile': ['Profile', 'Profil', 'Perfil', 'Profil'],
	'settings.tab.appearance': ['Appearance', 'Apparence', 'Apariencia', 'Darstellung'],
	'settings.tab.api': ['API', 'API', 'API', 'API'],
	'settings.tab.advanced': ['Advanced', 'Avancé', 'Avanzado', 'Erweitert'],
	'settings.head.profile': ['Settings — Casier', 'Paramètres — Casier', 'Ajustes — Casier', 'Einstellungen — Casier'],
	'settings.head.appearance': ['Appearance — Casier', 'Apparence — Casier', 'Apariencia — Casier', 'Darstellung — Casier'],
	'settings.head.api': ['API — Casier', 'API — Casier', 'API — Casier', 'API — Casier'],
	'settings.head.advanced': ['Advanced — Casier', 'Avancé — Casier', 'Avanzado — Casier', 'Erweitert — Casier'],

	'profile.signedInWith': ['Signed in with', 'Connecté avec', 'Sesión iniciada con', 'Angemeldet mit'],
	'profile.interface': ['Interface', 'Interface', 'Interfaz', 'Oberfläche'],
	'profile.account.title': ['Account', 'Compte', 'Cuenta', 'Konto'],
	'profile.account.description': [
		'An account in Casier is an email address. Your name and picture come from your Facile account.',
		'Un compte Casier est une adresse e-mail. Votre nom et votre photo viennent de votre compte Facile.',
		'Una cuenta de Casier es una dirección de correo. Tu nombre y tu imagen proceden de tu cuenta de Facile.',
		'Ein Casier-Konto ist eine E-Mail-Adresse. Name und Bild stammen aus deinem Facile-Konto.'
	],
	'profile.email.description': [
		'Your identity across every project you belong to. Members are invited by this address.',
		'Votre identité dans tous les projets dont vous faites partie. Les membres sont invités avec cette adresse.',
		'Tu identidad en todos los proyectos a los que perteneces. Los miembros se invitan con esta dirección.',
		'Deine Identität in allen Projekten, denen du angehörst. Mitglieder werden über diese Adresse eingeladen.'
	],
	'profile.picture.label': ['Name and picture', 'Nom et photo', 'Nombre e imagen', 'Name und Bild'],
	'profile.picture.description': [
		'Set them in your Facile account. A change shows here on the next page load.',
		'Définissez-les dans votre compte Facile. Une modification apparaît ici au prochain chargement de la page.',
		'Defínelos en tu cuenta de Facile. Un cambio se ve aquí en la próxima carga de la página.',
		'Lege sie in deinem Facile-Konto fest. Eine Änderung erscheint hier beim nächsten Laden der Seite.'
	],
	'profile.picture.edit': [
		'Edit in Facile account',
		'Modifier dans le compte Facile',
		'Editar en la cuenta de Facile',
		'Im Facile-Konto bearbeiten'
	],
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
	'profile.session.title': ['Session', 'Session', 'Sesión', 'Sitzung'],
	'profile.session.description': [
		'Sessions are per browser. Signing out here leaves your other devices alone.',
		'Les sessions sont propres à chaque navigateur. Se déconnecter ici n’affecte pas vos autres appareils.',
		'Las sesiones son por navegador. Cerrar sesión aquí no afecta a tus otros dispositivos.',
		'Sitzungen gelten pro Browser. Die Abmeldung hier lässt deine anderen Geräte unberührt.'
	],
	'profile.logout.label': ['Log out', 'Se déconnecter', 'Cerrar sesión', 'Abmelden'],
	'profile.logout.description': [
		'Ends this session and returns you to the sign-in page. API tokens keep working — revoke those from the project that owns them.',
		'Termine cette session et vous ramène à la page de connexion. Les jetons d’API continuent de fonctionner : révoquez-les depuis le projet qui les détient.',
		'Termina esta sesión y te lleva a la página de inicio de sesión. Los tokens de API siguen funcionando: revócalos desde el proyecto al que pertenecen.',
		'Beendet diese Sitzung und führt dich zur Anmeldeseite. API-Tokens funktionieren weiter — widerrufe sie im Projekt, dem sie gehören.'
	],
	'profile.logout.busy': ['Signing out…', 'Déconnexion…', 'Cerrando sesión…', 'Abmeldung…'],
	'profile.logout.warning': [
		'The server did not confirm the sign-out, ending the session locally anyway.',
		'Le serveur n’a pas confirmé la déconnexion, la session est quand même terminée localement.',
		'El servidor no confirmó el cierre de sesión; se termina la sesión localmente de todos modos.',
		'Der Server hat die Abmeldung nicht bestätigt; die Sitzung wird trotzdem lokal beendet.'
	],

	'appearance.theme.title': ['Theme', 'Thème', 'Tema', 'Design'],
	'appearance.theme.description': [
		'Stored in this browser. Every device you sign in from keeps its own choice.',
		'Enregistré dans ce navigateur. Chaque appareil depuis lequel vous vous connectez garde son propre choix.',
		'Se guarda en este navegador. Cada dispositivo desde el que inicias sesión conserva su propia elección.',
		'In diesem Browser gespeichert. Jedes Gerät, mit dem du dich anmeldest, behält seine eigene Wahl.'
	],
	'appearance.scheme.label': ['Colour scheme', 'Thème de couleur', 'Esquema de color', 'Farbschema'],
	'appearance.scheme.description': [
		'System follows your operating system and keeps following it while the app is open.',
		'« Système » suit votre système d’exploitation et continue de le suivre tant que l’application est ouverte.',
		'«Sistema» sigue tu sistema operativo y lo sigue mientras la aplicación está abierta.',
		'„System“ folgt deinem Betriebssystem und folgt ihm weiter, solange die App geöffnet ist.'
	],
	'appearance.mode.system': ['System', 'Système', 'Sistema', 'System'],
	'appearance.mode.light': ['Light', 'Clair', 'Claro', 'Hell'],
	'appearance.mode.dark': ['Dark', 'Sombre', 'Oscuro', 'Dunkel'],
	'appearance.now.label': ['Showing now', 'Affiché actuellement', 'Mostrando ahora', 'Aktuell angezeigt'],
	'appearance.now.description': [
		'What the two-token palette resolves to right now — chroma-zero either way.',
		'Ce que la palette à deux jetons donne en ce moment, sans chroma dans les deux cas.',
		'A qué resuelve ahora la paleta de dos tokens, sin croma en ambos casos.',
		'Was die Zwei-Token-Palette gerade ergibt — in beiden Fällen ohne Buntheit.'
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
	'advanced.about.open': ['Open projects', 'Ouvrir les projets', 'Abrir proyectos', 'Projekte öffnen'],
	'advanced.danger.title': ['Danger zone', 'Zone dangereuse', 'Zona de peligro', 'Gefahrenbereich'],
	'advanced.danger.description': [
		'Irreversible actions live at the bottom of this tab — never in the tab strip.',
		'Les actions irréversibles se trouvent en bas de cet onglet, jamais dans la barre d’onglets.',
		'Las acciones irreversibles están al final de esta pestaña, nunca en la barra de pestañas.',
		'Unumkehrbare Aktionen stehen am Ende dieses Tabs — nie in der Tab-Leiste.'
	],
	'advanced.danger.note': [
		'Casier has nothing destructive at the account level: an account is only an email, and it owns no data of its own. What is destructive belongs to a project — deleting an environment, purging a secret, deleting the project itself — and lives in that project’s settings, gated on owner or admin.',
		'Casier n’a rien de destructeur au niveau du compte : un compte n’est qu’une adresse e-mail et ne possède aucune donnée propre. Ce qui est destructeur appartient à un projet (supprimer un environnement, purger un secret, supprimer le projet) et se trouve dans les paramètres du projet, réservé au propriétaire ou à l’administrateur.',
		'Casier no tiene nada destructivo a nivel de cuenta: una cuenta es solo un correo y no posee datos propios. Lo destructivo pertenece a un proyecto (borrar un entorno, purgar un secreto, borrar el proyecto) y está en los ajustes del proyecto, reservado al propietario o administrador.',
		'Casier hat auf Kontoebene nichts Destruktives: Ein Konto ist nur eine E-Mail-Adresse und besitzt keine eigenen Daten. Destruktives gehört zu einem Projekt — Umgebung löschen, Secret endgültig entfernen, Projekt löschen — und liegt in den Projekteinstellungen, nur für Eigentümer oder Admins.'
	]
} as const satisfies Record<string, Row>;
