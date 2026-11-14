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
	'landing.description': [
		'Encrypted secrets for your team, injected straight into your app.',
		'Des secrets chiffrés pour votre équipe, injectés directement dans votre application.',
		'Secretos cifrados para tu equipo, inyectados directamente en tu aplicación.',
		'Verschlüsselte Secrets für dein Team, direkt an deine App übergeben.'
	],
	'landing.heroTitle': [
		'Your secrets.|Your server.',
		'Vos secrets.|Votre serveur.',
		'Tus secretos.|Tu servidor.',
		'Deine Secrets.|Dein Server.'
	],
	'landing.demo.label': [
		'A secret is added to the prod environment, then casier run starts an app with it. Every value stays masked.',
		'Un secret est ajouté à l’environnement prod, puis casier run lance une application avec. Chaque valeur reste masquée.',
		'Se añade un secreto al entorno prod y casier run arranca una aplicación con él. Todos los valores siguen ocultos.',
		'Ein Secret kommt in die Umgebung prod, dann startet casier run eine App damit. Jeder Wert bleibt maskiert.'
	],
	'landing.nav.features': ['Features', 'Fonctionnalités', 'Funciones', 'Funktionen'],
	'landing.nav.download': ['CLI', 'CLI', 'CLI', 'CLI'],
	'landing.nav.pricing': ['Pricing', 'Tarifs', 'Precios', 'Preise'],
	'landing.feature.encrypted.title': [
		'Encrypted at rest',
		'Chiffré au repos',
		'Cifrado en reposo',
		'Verschlüsselt gespeichert'
	],
	'landing.feature.encrypted.body': [
		'Each project has its own AES-256-GCM key.',
		'Chaque projet a sa propre clé AES-256-GCM.',
		'Cada proyecto tiene su propia clave AES-256-GCM.',
		'Jedes Projekt hat einen eigenen AES-256-GCM-Schlüssel.'
	],
	'landing.feature.environments.title': [
		'An environment per deploy',
		'Un environnement par déploiement',
		'Un entorno por despliegue',
		'Eine Umgebung pro Deployment'
	],
	'landing.feature.environments.body': [
		'dev, staging and prod, and every secret keeps its history.',
		'dev, staging et prod, et chaque secret garde son historique.',
		'dev, staging y prod, y cada secreto guarda su historial.',
		'dev, staging und prod, und jedes Secret behält seinen Verlauf.'
	],
	'landing.feature.ci.title': ['Made for CI', 'Pensé pour la CI', 'Pensado para CI', 'Für CI gemacht'],
	'landing.feature.ci.body': [
		'Scoped API tokens give a machine only what it needs.',
		'Des jetons d’API limités donnent à une machine le strict nécessaire.',
		'Los tokens de API acotados dan a una máquina solo lo que necesita.',
		'Eingeschränkte API-Tokens geben einer Maschine nur, was sie braucht.'
	],
	'landing.feature.audit.title': [
		'Every reveal is logged',
		'Chaque lecture est tracée',
		'Cada lectura queda registrada',
		'Jedes Aufdecken wird protokolliert'
	],
	'landing.feature.audit.body': [
		'The audit log shows who read which secrets, and when.',
		'Le journal d’audit montre qui a lu quels secrets, et quand.',
		'El registro de auditoría muestra quién leyó qué secretos y cuándo.',
		'Das Audit-Log zeigt, wer welche Secrets wann gelesen hat.'
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
		'casier run legt Secrets in die Umgebung deiner App, nicht in eine .env-Datei.'
	],
	'downloads.cta': ['Install the CLI', 'Installer la CLI', 'Instalar la CLI', 'CLI installieren'],
	'downloads.install.script': ['Install script', 'Script d’installation', 'Script de instalación', 'Installationsskript'],
	'downloads.install.facile': ['With facile', 'Avec facile', 'Con facile', 'Mit facile'],
	'downloads.install.brew': ['Homebrew', 'Homebrew', 'Homebrew', 'Homebrew'],
	'landing.ctaTitle': ['Ready to sign in?', 'Prêt à vous connecter ?', '¿Listo para iniciar sesión?', 'Bereit zur Anmeldung?'],
	'landing.ctaText': [
		'Use your Facile account.',
		'Utilisez votre compte Facile.',
		'Usa tu cuenta de Facile.',
		'Melde dich mit deinem Facile-Konto an.'
	],
	'landing.pricing': ['Pricing', 'Tarifs', 'Precios', 'Preise']
} as const satisfies Record<string, Row>;
