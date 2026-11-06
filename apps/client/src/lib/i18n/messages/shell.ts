import type { Row } from '../row';

/** Strings shared by the whole app: buttons, the shell, the sign-in and landing pages. */
export const shell = {
	'common.cancel': ['Cancel', 'Annuler', 'Cancelar', 'Abbrechen'],
	'common.save': ['Save', 'Enregistrer', 'Guardar', 'Speichern'],
	'common.saving': ['Saving…', 'Enregistrement…', 'Guardando…', 'Wird gespeichert…'],
	'common.create': ['Create', 'Créer', 'Crear', 'Erstellen'],
	'common.creating': ['Creating…', 'Création…', 'Creando…', 'Wird erstellt…'],
	'common.add': ['Add', 'Ajouter', 'Añadir', 'Hinzufügen'],
	'common.adding': ['Adding…', 'Ajout…', 'Añadiendo…', 'Wird hinzugefügt…'],
	'common.done': ['Done', 'Terminé', 'Hecho', 'Fertig'],
	'common.refresh': ['Refresh', 'Actualiser', 'Actualizar', 'Aktualisieren'],
	'common.loading': ['Loading', 'Chargement', 'Cargando', 'Wird geladen'],
	'common.loadingDots': ['Loading…', 'Chargement…', 'Cargando…', 'Wird geladen…'],
	'common.tryAgain': ['Try again', 'Réessayer', 'Reintentar', 'Erneut versuchen'],
	'common.genericError': [
		'Something went wrong.',
		'Une erreur est survenue.',
		'Se ha producido un error.',
		'Es ist ein Fehler aufgetreten.'
	],
	'common.clipboardUnavailable': [
		'Clipboard unavailable',
		'Presse-papiers indisponible',
		'Portapapeles no disponible',
		'Zwischenablage nicht verfügbar'
	],
	'common.optional': ['Optional', 'Facultatif', 'Opcional', 'Optional'],
	'common.never': ['Never', 'Jamais', 'Nunca', 'Nie'],
	'common.all': ['all', 'tous', 'todos', 'alle'],
	'common.name': ['Name', 'Nom', 'Nombre', 'Name'],
	'common.email': ['Email', 'E-mail', 'Correo electrónico', 'E-Mail'],
	'common.actions': ['Actions', 'Actions', 'Acciones', 'Aktionen'],
	'common.unknown': ['unknown', 'inconnu', 'desconocido', 'unbekannt'],
	'common.deletedUser': ['Deleted user', 'Utilisateur supprimé', 'Usuario eliminado', 'Gelöschter Nutzer'],
	'common.account': ['Account', 'Compte', 'Cuenta', 'Konto'],
	'role.owner': ['owner', 'propriétaire', 'propietario', 'Eigentümer'],
	'role.admin': ['admin', 'administrateur', 'administrador', 'Administrator'],
	'role.member': ['member', 'membre', 'miembro', 'Mitglied'],
	'role.viewer': ['viewer', 'lecteur', 'lector', 'Leser'],
	'nav.projects': ['Projects', 'Projets', 'Proyectos', 'Projekte'],
	'nav.allProjects': ['All projects', 'Tous les projets', 'Todos los proyectos', 'Alle Projekte'],
	'login.title': ['Sign in — Casier', 'Connexion — Casier', 'Iniciar sesión — Casier', 'Anmelden — Casier'],
	'login.tagline': [
		'Self-hosted secrets and environment management for teams.',
		'Gestion de secrets et d’environnements auto-hébergée pour les équipes.',
		'Gestión de secretos y entornos autoalojada para equipos.',
		'Selbst gehostete Verwaltung von Secrets und Umgebungen für Teams.'
	],
	'login.failed': [
		'Sign-in did not complete. Try again.',
		'La connexion n’a pas abouti. Réessayez.',
		'No se pudo completar el inicio de sesión. Inténtalo de nuevo.',
		'Die Anmeldung wurde nicht abgeschlossen. Versuche es erneut.'
	],
	'landing.title': [
		'Casier — Self-hosted secrets manager',
		'Casier — Gestionnaire de secrets auto-hébergé',
		'Casier — Gestor de secretos autoalojado',
		'Casier — Selbst gehosteter Secrets-Manager'
	],
	'landing.description': [
		'Casier is a self-hosted secrets and environment manager for teams. Encrypted at rest, versioned, audited — and one command away from your app.',
		'Casier est un gestionnaire de secrets et d’environnements auto-hébergé pour les équipes. Chiffré au repos, versionné, audité, et à une commande de votre application.',
		'Casier es un gestor de secretos y entornos autoalojado para equipos. Cifrado en reposo, versionado, auditado y a un comando de tu aplicación.',
		'Casier ist ein selbst gehosteter Manager für Secrets und Umgebungen für Teams. Verschlüsselt gespeichert, versioniert, protokolliert und nur einen Befehl von deiner App entfernt.'
	],
	'landing.heroTitle': [
		'Your secrets.|Your server.',
		'Vos secrets.|Votre serveur.',
		'Tus secretos.|Tu servidor.',
		'Deine Secrets.|Dein Server.'
	],
	'landing.feature.encrypted.title': [
		'Encrypted at rest',
		'Chiffré au repos',
		'Cifrado en reposo',
		'Verschlüsselt gespeichert'
	],
	'landing.feature.encrypted.body': [
		'Every project gets its own AES-256-GCM data key. Values are encrypted before they ever reach the database.',
		'Chaque projet a sa propre clé de données AES-256-GCM. Les valeurs sont chiffrées avant d’atteindre la base de données.',
		'Cada proyecto tiene su propia clave de datos AES-256-GCM. Los valores se cifran antes de llegar a la base de datos.',
		'Jedes Projekt hat einen eigenen AES-256-GCM-Datenschlüssel. Werte werden verschlüsselt, bevor sie die Datenbank erreichen.'
	],
	'landing.feature.environments.title': [
		'An environment per deploy',
		'Un environnement par déploiement',
		'Un entorno por despliegue',
		'Eine Umgebung pro Deployment'
	],
	'landing.feature.environments.body': [
		'dev, staging and prod out of the box. Diff them, import a .env, roll any secret back to an earlier version.',
		'dev, staging et prod dès le départ. Comparez-les, importez un .env, restaurez n’importe quel secret à une version antérieure.',
		'dev, staging y prod de serie. Compáralos, importa un .env y devuelve cualquier secreto a una versión anterior.',
		'dev, staging und prod von Anfang an. Vergleiche sie, importiere eine .env und setze jedes Secret auf eine frühere Version zurück.'
	],
	'landing.feature.ci.title': ['Made for CI', 'Pensé pour la CI', 'Pensado para CI', 'Für CI gemacht'],
	'landing.feature.ci.body': [
		'casier run injects secrets as env vars. Scoped API tokens for machines, a full audit log for humans.',
		'casier run injecte les secrets comme variables d’environnement. Des jetons d’API limités pour les machines, un journal d’audit complet pour les humains.',
		'casier run inyecta los secretos como variables de entorno. Tokens de API acotados para máquinas y un registro de auditoría completo para personas.',
		'casier run übergibt Secrets als Umgebungsvariablen. Eingeschränkte API-Tokens für Maschinen, ein vollständiges Audit-Log für Menschen.'
	],
	'landing.commandsTitle': [
		'Three commands and you’re done.',
		'Trois commandes et c’est terminé.',
		'Tres comandos y listo.',
		'Drei Befehle und fertig.'
	],
	'landing.commandsNote': [
		'Secrets never touch your shell history or a .env file on disk.',
		'Les secrets ne passent jamais par l’historique du shell ni par un fichier .env sur le disque.',
		'Los secretos nunca pasan por el historial del shell ni por un archivo .env en disco.',
		'Secrets landen weder in der Shell-Historie noch in einer .env-Datei auf der Festplatte.'
	],
	'landing.ctaTitle': ['Ready to sign in?', 'Prêt à vous connecter ?', '¿Listo para iniciar sesión?', 'Bereit zur Anmeldung?'],
	'landing.ctaText': [
		'Use your Facile account to access Casier.',
		'Utilisez votre compte Facile pour accéder à Casier.',
		'Usa tu cuenta de Facile para acceder a Casier.',
		'Melde dich mit deinem Facile-Konto bei Casier an.'
	],
	'landing.pricing': ['Pricing', 'Tarifs', 'Precios', 'Preise']
} as const satisfies Record<string, Row>;
