import type { Row } from '../row';

/** The project list, the project layout, the environments tab and the audit log. */
export const projects = {
	'projects.head': ['Projects — Casier', 'Projets — Casier', 'Proyectos — Casier', 'Projekte — Casier'],
	'projects.description': [
		'Every project has its own encryption key and its own environments.',
		'Chaque projet a sa propre clé de chiffrement et ses propres environnements.',
		'Cada proyecto tiene su propia clave de cifrado y sus propios entornos.',
		'Jedes Projekt hat einen eigenen Verschlüsselungsschlüssel und eigene Umgebungen.'
	],
	'projects.new': ['New project', 'Nouveau projet', 'Nuevo proyecto', 'Neues Projekt'],
	'projects.empty.title': ['No projects yet', 'Aucun projet pour l’instant', 'Aún no hay proyectos', 'Noch keine Projekte'],
	'projects.empty.description': [
		'A project groups the secrets for one app or service, across all of its environments.',
		'Un projet regroupe les secrets d’une application ou d’un service, dans tous ses environnements.',
		'Un proyecto agrupa los secretos de una aplicación o servicio en todos sus entornos.',
		'Ein Projekt fasst die Secrets einer App oder eines Dienstes über alle Umgebungen zusammen.'
	],
	'projects.noDescription': ['No description', 'Aucune description', 'Sin descripción', 'Keine Beschreibung'],
	'projects.updated': ['Updated {date}', 'Mis à jour le {date}', 'Actualizado el {date}', 'Aktualisiert am {date}'],
	'projects.nameTooShort': [
		'Name must be at least 2 characters',
		'Le nom doit contenir au moins 2 caractères',
		'El nombre debe tener al menos 2 caracteres',
		'Der Name muss mindestens 2 Zeichen lang sein'
	],
	'projects.description.label': ['Description', 'Description', 'Descripción', 'Beschreibung'],
	'projects.description.placeholder': [
		'What lives in this project?',
		'Que contient ce projet ?',
		'¿Qué hay en este proyecto?',
		'Was liegt in diesem Projekt?'
	],
	'projects.created': ['Project "{name}" created', 'Projet « {name} » créé', 'Proyecto «{name}» creado', 'Projekt „{name}“ erstellt'],
	'projects.error.load': [
		'Failed to load projects',
		'Échec du chargement des projets',
		'No se pudieron cargar los proyectos',
		'Projekte konnten nicht geladen werden'
	],
	'projects.error.create': [
		'Failed to create project',
		'Échec de la création du projet',
		'No se pudo crear el proyecto',
		'Projekt konnte nicht erstellt werden'
	],

	'project.fallbackName': ['Project', 'Projet', 'Proyecto', 'Projekt'],
	'project.error.load': [
		'Failed to load project',
		'Échec du chargement du projet',
		'No se pudo cargar el proyecto',
		'Projekt konnte nicht geladen werden'
	],
	'project.sections': ['Project sections', 'Sections du projet', 'Secciones del proyecto', 'Projektbereiche'],
	'project.tab.secrets': ['Secrets', 'Secrets', 'Secretos', 'Secrets'],
	'project.tab.members': ['Members', 'Membres', 'Miembros', 'Mitglieder'],
	'project.tab.tokens': ['API tokens', 'Jetons d’API', 'Tokens de API', 'API-Tokens'],
	'project.tab.audit': ['Audit log', 'Journal d’audit', 'Registro de auditoría', 'Audit-Log'],
	'project.tab.settings': ['Settings', 'Paramètres', 'Ajustes', 'Einstellungen'],

	'env.empty.title': ['No environments', 'Aucun environnement', 'Sin entornos', 'Keine Umgebungen'],
	'env.empty.description': [
		'A project needs at least one environment before it can hold secrets. Environments keep production, staging and local values apart.',
		'Un projet a besoin d’au moins un environnement pour contenir des secrets. Les environnements séparent les valeurs de production, de staging et locales.',
		'Un proyecto necesita al menos un entorno para guardar secretos. Los entornos separan los valores de producción, staging y locales.',
		'Ein Projekt braucht mindestens eine Umgebung, um Secrets aufzunehmen. Umgebungen trennen Produktions-, Staging- und lokale Werte.'
	],
	'env.create': ['Create an environment', 'Créer un environnement', 'Crear un entorno', 'Umgebung erstellen'],
	'env.manage': ['Manage environments', 'Gérer les environnements', 'Gestionar entornos', 'Umgebungen verwalten'],
	'env.tabs': ['Environments', 'Environnements', 'Entornos', 'Umgebungen'],

	'audit.head': [
		'Audit log — {project} — Casier',
		'Journal d’audit — {project} — Casier',
		'Registro de auditoría — {project} — Casier',
		'Audit-Log — {project} — Casier'
	],
	'audit.description': [
		'Every change to a secret, a member, a token or a setting, newest first. Revealing a value is recorded too — reading a secret is an event, not a free action.',
		'Chaque modification d’un secret, d’un membre, d’un jeton ou d’un réglage, du plus récent au plus ancien. Révéler une valeur est aussi enregistré : lire un secret est un événement, pas un geste gratuit.',
		'Cada cambio en un secreto, un miembro, un token o un ajuste, del más reciente al más antiguo. Revelar un valor también se registra: leer un secreto es un evento, no una acción gratuita.',
		'Jede Änderung an einem Secret, Mitglied, Token oder einer Einstellung, neueste zuerst. Auch das Anzeigen eines Werts wird erfasst — ein Secret zu lesen ist ein Ereignis, keine Nebensache.'
	],
	'audit.empty.title': ['No activity yet', 'Aucune activité pour l’instant', 'Aún no hay actividad', 'Noch keine Aktivität'],
	'audit.empty.description': [
		'Changes to secrets, members and settings will show up here.',
		'Les modifications des secrets, des membres et des réglages apparaîtront ici.',
		'Los cambios en secretos, miembros y ajustes aparecerán aquí.',
		'Änderungen an Secrets, Mitgliedern und Einstellungen erscheinen hier.'
	],
	'audit.col.when': ['When', 'Quand', 'Cuándo', 'Wann'],
	'audit.col.who': ['Who', 'Qui', 'Quién', 'Wer'],
	'audit.col.action': ['Action', 'Action', 'Acción', 'Aktion'],
	'audit.col.resource': ['Resource', 'Ressource', 'Recurso', 'Ressource'],
	'audit.col.detail': ['Detail', 'Détail', 'Detalle', 'Detail'],
	'audit.limit': [
		'Showing the 50 most recent events.',
		'Affichage des 50 événements les plus récents.',
		'Se muestran los 50 eventos más recientes.',
		'Die 50 neuesten Ereignisse werden angezeigt.'
	],
	'audit.error.load': [
		'Failed to load audit log',
		'Échec du chargement du journal d’audit',
		'No se pudo cargar el registro de auditoría',
		'Audit-Log konnte nicht geladen werden'
	]
} as const satisfies Record<string, Row>;
