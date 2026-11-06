import type { Row } from '../row';

/** A project's settings tab. */
export const psettings = {
	'psettings.head': [
		'Settings — {project} — Casier',
		'Paramètres — {project} — Casier',
		'Ajustes — {project} — Casier',
		'Einstellungen — {project} — Casier'
	],
	'psettings.general.title': ['General', 'Général', 'General', 'Allgemein'],
	'psettings.general.description': [
		'How this project is named and how long a secret may sit unchanged before Casier flags it.',
		'Comment ce projet est nommé et combien de temps un secret peut rester inchangé avant que Casier le signale.',
		'Cómo se llama este proyecto y cuánto tiempo puede un secreto permanecer sin cambios antes de que Casier lo marque.',
		'Wie dieses Projekt heißt und wie lange ein Secret unverändert bleiben darf, bevor Casier es markiert.'
	],
	'psettings.name.description': [
		'Renaming a project rewrites its slug, and every URL that names it moves with it.',
		'Renommer un projet réécrit son slug, et toutes les URL qui le nomment changent avec lui.',
		'Renombrar un proyecto reescribe su slug, y todas las URL que lo nombran cambian con él.',
		'Beim Umbenennen eines Projekts ändert sich sein Slug, und alle URLs, die es nennen, ziehen mit.'
	],
	'psettings.description.description': [
		'What lives in this project. Shown on the project list.',
		'Ce que contient ce projet. Affiché dans la liste des projets.',
		'Qué hay en este proyecto. Se muestra en la lista de proyectos.',
		'Was in diesem Projekt liegt. Wird in der Projektliste angezeigt.'
	],
	'psettings.rotation.label': ['Rotation reminder', 'Rappel de rotation', 'Recordatorio de rotación', 'Rotationserinnerung'],
	'psettings.rotation.days': ['days', 'jours', 'días', 'Tage'],
	'psettings.rotation.invalid': [
		'Use a whole number of days between 0 and 3650',
		'Utilisez un nombre entier de jours entre 0 et 3650',
		'Usa un número entero de días entre 0 y 3650',
		'Verwende eine ganze Zahl von Tagen zwischen 0 und 3650'
	],
	'psettings.rotation.off': [
		'Rotation reminders are off. Secrets are never flagged as stale.',
		'Les rappels de rotation sont désactivés. Aucun secret n’est jamais signalé comme périmé.',
		'Los recordatorios de rotación están desactivados. Los secretos nunca se marcan como obsoletos.',
		'Rotationserinnerungen sind aus. Secrets werden nie als veraltet markiert.'
	],
	'psettings.rotation.on': [
		'Secrets whose value has not changed in {days} days are flagged for rotation in every environment of this project.',
		'Les secrets dont la valeur n’a pas changé depuis {days} jours sont signalés pour rotation dans chaque environnement de ce projet.',
		'Los secretos cuyo valor no ha cambiado en {days} días se marcan para rotación en cada entorno de este proyecto.',
		'Secrets, deren Wert sich seit {days} Tagen nicht geändert hat, werden in jeder Umgebung dieses Projekts zur Rotation markiert.'
	],
	'psettings.save.dirty': [
		'You have unsaved changes.',
		'Vous avez des modifications non enregistrées.',
		'Tienes cambios sin guardar.',
		'Du hast ungespeicherte Änderungen.'
	],
	'psettings.save.clean': [
		'Nothing has changed yet.',
		'Rien n’a encore changé.',
		'Aún no ha cambiado nada.',
		'Noch nichts geändert.'
	],
	'psettings.save.button': ['Save changes', 'Enregistrer les modifications', 'Guardar cambios', 'Änderungen speichern'],
	'psettings.updated': ['Project updated', 'Projet mis à jour', 'Proyecto actualizado', 'Projekt aktualisiert'],
	'psettings.envs.title': ['Environments', 'Environnements', 'Entornos', 'Umgebungen'],
	'psettings.envs.description': [
		'Each environment holds its own copy of every secret. A protected one takes owner or admin to write to.',
		'Chaque environnement contient sa propre copie de chaque secret. Écrire dans un environnement protégé exige d’être propriétaire ou administrateur.',
		'Cada entorno guarda su propia copia de cada secreto. Escribir en uno protegido requiere ser propietario o administrador.',
		'Jede Umgebung enthält ihre eigene Kopie jedes Secrets. In eine geschützte Umgebung schreiben nur Eigentümer oder Admins.'
	],
	'psettings.env.protected': ['Protected', 'Protégé', 'Protegido', 'Geschützt'],
	'psettings.env.unprotectAria': [
		'Remove protection from {name}',
		'Retirer la protection de {name}',
		'Quitar la protección de {name}',
		'Schutz von {name} aufheben'
	],
	'psettings.env.protectAria': ['Protect {name}', 'Protéger {name}', 'Proteger {name}', '{name} schützen'],
	'psettings.env.unprotectTitle': [
		'Allow any member to write here',
		'Autoriser tout membre à écrire ici',
		'Permitir que cualquier miembro escriba aquí',
		'Jedem Mitglied das Schreiben erlauben'
	],
	'psettings.env.protectTitle': [
		'Require owner or admin to write here',
		'Exiger propriétaire ou administrateur pour écrire ici',
		'Exigir propietario o administrador para escribir aquí',
		'Schreiben nur für Eigentümer oder Admins'
	],
	'psettings.env.deleteAria': [
		'Delete environment {name}',
		'Supprimer l’environnement {name}',
		'Eliminar el entorno {name}',
		'Umgebung {name} löschen'
	],
	'psettings.env.deleteTitle': ['Delete {name}', 'Supprimer {name}', 'Eliminar {name}', '{name} löschen'],
	'psettings.env.keepOne': [
		'A project keeps at least one environment',
		'Un projet garde au moins un environnement',
		'Un proyecto conserva al menos un entorno',
		'Ein Projekt behält mindestens eine Umgebung'
	],
	'psettings.env.add.label': ['Add an environment', 'Ajouter un environnement', 'Añadir un entorno', 'Umgebung hinzufügen'],
	'psettings.env.add.description': [
		'The slug is derived from the name and is what the CLI and every token scope refer to.',
		'Le slug est dérivé du nom ; c’est ce que la CLI et chaque périmètre de jeton utilisent.',
		'El slug se deriva del nombre y es lo que usan la CLI y cada ámbito de token.',
		'Der Slug wird aus dem Namen abgeleitet und ist das, worauf CLI und jeder Token-Geltungsbereich verweisen.'
	],
	'psettings.env.created': [
		'Environment "{name}" created',
		'Environnement « {name} » créé',
		'Entorno «{name}» creado',
		'Umgebung „{name}“ erstellt'
	],
	'psettings.env.unprotected': [
		'"{name}" is no longer protected',
		'« {name} » n’est plus protégé',
		'«{name}» ya no está protegido',
		'„{name}“ ist nicht mehr geschützt'
	],
	'psettings.env.nowProtected': [
		'"{name}" now requires owner or admin to write',
		'« {name} » exige désormais propriétaire ou administrateur pour écrire',
		'«{name}» ahora exige propietario o administrador para escribir',
		'„{name}“ erfordert jetzt Eigentümer oder Admin zum Schreiben'
	],
	'psettings.env.deleted': [
		'Environment "{name}" deleted',
		'Environnement « {name} » supprimé',
		'Entorno «{name}» eliminado',
		'Umgebung „{name}“ gelöscht'
	],
	'psettings.env.deleteDialog': [
		'Delete environment',
		'Supprimer l’environnement',
		'Eliminar entorno',
		'Umgebung löschen'
	],
	'psettings.env.deleteMessage': [
		'Deleting "{name}" destroys every secret stored in it, including all past versions, and every API token scope that named it. Other environments in this project are untouched. This cannot be undone.',
		'Supprimer « {name} » détruit tous les secrets qu’il contient, avec toutes leurs versions passées, et tout périmètre de jeton d’API qui le nommait. Les autres environnements du projet ne sont pas touchés. Cette action est irréversible.',
		'Eliminar «{name}» destruye todos los secretos que contiene, con todas sus versiones anteriores, y todo ámbito de token de API que lo nombraba. Los demás entornos del proyecto no se tocan. No se puede deshacer.',
		'Das Löschen von „{name}“ zerstört alle darin gespeicherten Secrets samt früheren Versionen und jeden API-Token-Geltungsbereich, der sie nannte. Andere Umgebungen im Projekt bleiben unberührt. Das lässt sich nicht rückgängig machen.'
	],
	'psettings.danger.title': ['Danger zone', 'Zone dangereuse', 'Zona de peligro', 'Gefahrenbereich'],
	'psettings.danger.description': [
		'Irreversible, and nobody can undo it for you.',
		'Irréversible, et personne ne peut l’annuler pour vous.',
		'Irreversible, y nadie puede deshacerlo por ti.',
		'Unumkehrbar, und niemand kann es für dich rückgängig machen.'
	],
	'psettings.delete.label': ['Delete this project', 'Supprimer ce projet', 'Eliminar este proyecto', 'Dieses Projekt löschen'],
	'psettings.delete.description': [
		'Erases every environment, every secret value and version, every member and every API token {name} holds. Anything reading from it breaks immediately. There is no backup and no undo.',
		'Efface chaque environnement, chaque valeur et version de secret, chaque membre et chaque jeton d’API que contient {name}. Tout ce qui y lit cesse immédiatement de fonctionner. Il n’y a ni sauvegarde ni annulation.',
		'Borra cada entorno, cada valor y versión de secreto, cada miembro y cada token de API que contiene {name}. Todo lo que lea de él deja de funcionar de inmediato. No hay copia de seguridad ni deshacer.',
		'Löscht jede Umgebung, jeden Secret-Wert samt Version, jedes Mitglied und jedes API-Token von {name}. Alles, was daraus liest, bricht sofort. Es gibt kein Backup und kein Rückgängigmachen.'
	],
	'psettings.delete.button': ['Delete project', 'Supprimer le projet', 'Eliminar proyecto', 'Projekt löschen'],
	'psettings.delete.message': [
		'Deleting "{name}" permanently destroys all of its environments, secrets, members and API tokens. Anything reading these secrets will stop working. This cannot be undone.',
		'Supprimer « {name} » détruit définitivement tous ses environnements, secrets, membres et jetons d’API. Tout ce qui lit ces secrets cessera de fonctionner. Cette action est irréversible.',
		'Eliminar «{name}» destruye de forma permanente todos sus entornos, secretos, miembros y tokens de API. Todo lo que lea estos secretos dejará de funcionar. No se puede deshacer.',
		'Das Löschen von „{name}“ zerstört alle Umgebungen, Secrets, Mitglieder und API-Tokens endgültig. Alles, was diese Secrets liest, funktioniert nicht mehr. Das lässt sich nicht rückgängig machen.'
	],
	'psettings.delete.deleted': [
		'Project "{name}" deleted',
		'Projet « {name} » supprimé',
		'Proyecto «{name}» eliminado',
		'Projekt „{name}“ gelöscht'
	],
	'psettings.error.update': [
		'Failed to update project',
		'Échec de la mise à jour du projet',
		'No se pudo actualizar el proyecto',
		'Projekt konnte nicht aktualisiert werden'
	],
	'psettings.error.createEnv': [
		'Failed to create environment',
		'Échec de la création de l’environnement',
		'No se pudo crear el entorno',
		'Umgebung konnte nicht erstellt werden'
	],
	'psettings.error.protect': [
		'Failed to change protection',
		'Échec du changement de protection',
		'No se pudo cambiar la protección',
		'Schutz konnte nicht geändert werden'
	],
	'psettings.error.deleteEnv': [
		'Failed to delete environment',
		'Échec de la suppression de l’environnement',
		'No se pudo eliminar el entorno',
		'Umgebung konnte nicht gelöscht werden'
	],
	'psettings.error.deleteProject': [
		'Failed to delete project',
		'Échec de la suppression du projet',
		'No se pudo eliminar el proyecto',
		'Projekt konnte nicht gelöscht werden'
	]
} as const satisfies Record<string, Row>;
