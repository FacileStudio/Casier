import type { Row } from '@facile/muse';

/** Strings shared by the whole app: buttons, the shell, the sign-in and download pages. */
export const shell = {
	'common.cancel': ['Cancel', 'Annuler', 'Cancelar', 'Abbrechen'],
	'common.save': ['Save', 'Enregistrer', 'Guardar', 'Speichern'],
	'common.saving': ['Saving…', 'Enregistrement…', 'Guardando…', 'Wird gespeichert…'],
	'common.create': ['Create', 'Créer', 'Crear', 'Erstellen'],
	'common.creating': ['Creating…', 'Création…', 'Creando…', 'Wird erstellt…'],
	'common.add': ['Add', 'Ajouter', 'Añadir', 'Hinzufügen'],
	'common.adding': ['Adding…', 'Ajout…', 'Añadiendo…', 'Wird hinzugefügt…'],
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
	'login.failed': [
		'Sign-in did not complete. Try again.',
		'La connexion n’a pas abouti. Réessayez.',
		'No se pudo completar el inicio de sesión. Inténtalo de nuevo.',
		'Die Anmeldung wurde nicht abgeschlossen. Versuchen Sie es erneut.'
	],
	'downloads.title': [
		'Install the Casier CLI',
		'Installer la CLI Casier',
		'Instalar la CLI de Casier',
		'Die Casier-CLI installieren'
	],
	'downloads.heading': ['Install the CLI', 'Installer la CLI', 'Instala la CLI', 'Die CLI installieren'],
	'downloads.intro': [
		'casier run puts secrets in your app’s environment, not in a .env file.',
		'casier run place les secrets dans l’environnement de votre application, pas dans un fichier .env.',
		'casier run pone los secretos en el entorno de tu aplicación, no en un archivo .env.',
		'casier run legt Secrets in die Umgebung Ihrer App, nicht in eine .env-Datei.'
	],
	'downloads.cta': ['Install the CLI', 'Installer la CLI', 'Instalar la CLI', 'CLI installieren'],
	'downloads.install.script': ['Install script', 'Script d’installation', 'Script de instalación', 'Installationsskript'],
	'downloads.install.facile': ['With facile', 'Avec facile', 'Con facile', 'Mit facile'],
	'downloads.install.brew': ['Homebrew', 'Homebrew', 'Homebrew', 'Homebrew']
} as const satisfies Record<string, Row>;
