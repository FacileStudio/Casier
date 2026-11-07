import type { Row } from '../row';

/** The members and API tokens tabs of a project. */
export const members = {
	'members.head': [
		'Members — {project} — Casier',
		'Membres — {project} — Casier',
		'Miembros — {project} — Casier',
		'Mitglieder — {project} — Casier'
	],
	'members.title': ['Members', 'Membres', 'Miembros', 'Mitglieder'],
	'members.description': [
		'Everyone who can reach this project’s secrets. A role is set when the member is added and applies to every environment.',
		'Toutes les personnes qui peuvent accéder aux secrets de ce projet. Un rôle est défini à l’ajout du membre et s’applique à tous les environnements.',
		'Todas las personas que pueden acceder a los secretos de este proyecto. El rol se define al añadir al miembro y se aplica a todos los entornos.',
		'Alle, die auf die Secrets dieses Projekts zugreifen können. Die Rolle wird beim Hinzufügen festgelegt und gilt für alle Umgebungen.'
	],
	'members.add': ['Add member', 'Ajouter un membre', 'Añadir miembro', 'Mitglied hinzufügen'],
	'members.empty.title': ['No members yet', 'Aucun membre pour l’instant', 'Aún no hay miembros', 'Noch keine Mitglieder'],
	'members.empty.description': [
		'Add a teammate to share this project’s secrets with them.',
		'Ajoutez un coéquipier pour partager les secrets de ce projet avec lui.',
		'Añade a un compañero para compartir con él los secretos de este proyecto.',
		'Füge Teammitglieder hinzu, um die Secrets dieses Projekts mit ihnen zu teilen.'
	],
	'members.col.person': ['Person', 'Personne', 'Persona', 'Person'],
	'members.col.role': ['Role', 'Rôle', 'Rol', 'Rolle'],
	'members.you': ['This is you', 'C’est vous', 'Eres tú', 'Das bist du'],
	'members.role.owner': [
		'Everything, including deleting the project',
		'Tout, y compris supprimer le projet',
		'Todo, incluido eliminar el proyecto',
		'Alles, einschließlich das Projekt zu löschen'
	],
	'members.role.admin': [
		'Members, tokens, environments and secrets',
		'Membres, jetons, environnements et secrets',
		'Miembros, tokens, entornos y secretos',
		'Mitglieder, Tokens, Umgebungen und Secrets'
	],
	'members.role.member': [
		'Reads and writes secrets',
		'Lit et écrit les secrets',
		'Lee y escribe secretos',
		'Liest und schreibt Secrets'
	],
	'members.role.viewer': [
		'Reads secrets, changes nothing',
		'Lit les secrets, ne modifie rien',
		'Lee secretos, no cambia nada',
		'Liest Secrets, ändert nichts'
	],
	'members.remove.aria': ['Remove {email}', 'Retirer {email}', 'Quitar a {email}', '{email} entfernen'],
	'members.emailHelper': [
		'Must already be a Casier account on this instance.',
		'Doit déjà être un compte Casier sur cette instance.',
		'Debe ser ya una cuenta de Casier en esta instancia.',
		'Muss bereits ein Casier-Konto auf dieser Instanz sein.'
	],
	'members.roleHelper': [
		'A role can only be changed by removing and re-adding the member.',
		'Un rôle ne peut être changé qu’en retirant puis en rajoutant le membre.',
		'Un rol solo se cambia quitando y volviendo a añadir al miembro.',
		'Eine Rolle lässt sich nur ändern, indem das Mitglied entfernt und erneut hinzugefügt wird.'
	],
	'members.option.admin': [
		'Admin — manage members, tokens and environments',
		'Administrateur : gère les membres, les jetons et les environnements',
		'Administrador: gestiona miembros, tokens y entornos',
		'Admin — verwaltet Mitglieder, Tokens und Umgebungen'
	],
	'members.option.member': [
		'Member — read and write secrets',
		'Membre : lit et écrit les secrets',
		'Miembro: lee y escribe secretos',
		'Mitglied — liest und schreibt Secrets'
	],
	'members.option.viewer': [
		'Viewer — read-only',
		'Lecteur : lecture seule',
		'Lector: solo lectura',
		'Leser — nur lesen'
	],
	'members.added': ['{email} added as {role}', '{email} ajouté comme {role}', '{email} añadido como {role}', '{email} als {role} hinzugefügt'],
	'members.removed': ['{email} removed', '{email} retiré', '{email} quitado', '{email} entfernt'],
	'members.remove.title': ['Remove member', 'Retirer le membre', 'Quitar miembro', 'Mitglied entfernen'],
	'members.remove.message': [
		'This removes {email} from the project. They immediately lose access to every secret in every environment, and any session they have open stops working.',
		'Cela retire {email} du projet. La personne perd immédiatement l’accès à tous les secrets de tous les environnements, et ses sessions ouvertes cessent de fonctionner.',
		'Esto quita a {email} del proyecto. Pierde de inmediato el acceso a todos los secretos de todos los entornos y sus sesiones abiertas dejan de funcionar.',
		'{email} wird aus dem Projekt entfernt. Der Zugriff auf alle Secrets in allen Umgebungen endet sofort, und offene Sitzungen funktionieren nicht mehr.'
	],
	'members.error.load': [
		'Failed to load members',
		'Échec du chargement des membres',
		'No se pudieron cargar los miembros',
		'Mitglieder konnten nicht geladen werden'
	],
	'members.error.add': [
		'Failed to add member',
		'Échec de l’ajout du membre',
		'No se pudo añadir al miembro',
		'Mitglied konnte nicht hinzugefügt werden'
	],
	'members.error.remove': [
		'Failed to remove member',
		'Échec du retrait du membre',
		'No se pudo quitar al miembro',
		'Mitglied konnte nicht entfernt werden'
	],
	'tokens.head': [
		'API tokens — {project} — Casier',
		'Jetons d’API — {project} — Casier',
		'Tokens de API — {project} — Casier',
		'API-Tokens — {project} — Casier'
	],
	'tokens.description': [
		'Machine credentials for CI and the casier CLI. A token is bound to this project, to a permission set and to the environments it names — it can never read past them, and revoking one never touches the others.',
		'Identifiants machine pour la CI et la CLI casier. Un jeton est lié à ce projet, à un ensemble de permissions et aux environnements qu’il nomme : il ne peut jamais lire au-delà, et en révoquer un ne touche pas les autres.',
		'Credenciales de máquina para CI y la CLI casier. Un token queda ligado a este proyecto, a un conjunto de permisos y a los entornos que nombra: nunca puede leer más allá, y revocar uno no afecta a los demás.',
		'Maschinenzugänge für CI und die casier-CLI. Ein Token ist an dieses Projekt, einen Berechtigungssatz und die genannten Umgebungen gebunden — es kann nie darüber hinaus lesen, und der Widerruf eines Tokens berührt die anderen nicht.'
	],
	'tokens.create': ['Create a token', 'Créer un jeton', 'Crear un token', 'Token erstellen'],
	'tokens.col.environments': ['Environments', 'Environnements', 'Entornos', 'Umgebungen'],
	'tokens.status.expired': ['Expired', 'Expiré', 'Caducado', 'Abgelaufen'],
	'tokens.revoke': ['Revoke', 'Révoquer', 'Revocar', 'Widerrufen'],
	'tokens.permissions.read': [
		'Read — fetch secrets',
		'Lecture : récupérer les secrets',
		'Lectura: obtener secretos',
		'Lesen — Secrets abrufen'
	],
	'tokens.permissions.write': [
		'Read and write — fetch and update secrets',
		'Lecture et écriture : récupérer et modifier les secrets',
		'Lectura y escritura: obtener y actualizar secretos',
		'Lesen und Schreiben — Secrets abrufen und ändern'
	],
	'tokens.environments.helperAll': [
		'Comma-separated environment slugs, or * for every environment.',
		'Slugs d’environnement séparés par des virgules, ou * pour tous les environnements.',
		'Slugs de entorno separados por comas, o * para todos los entornos.',
		'Kommagetrennte Umgebungs-Slugs oder * für alle Umgebungen.'
	],
	'tokens.environments.helperList': [
		'Comma-separated environment slugs, or * for every environment. This project has: {slugs}.',
		'Slugs d’environnement séparés par des virgules, ou * pour tous les environnements. Ce projet a : {slugs}.',
		'Slugs de entorno separados por comas, o * para todos los entornos. Este proyecto tiene: {slugs}.',
		'Kommagetrennte Umgebungs-Slugs oder * für alle Umgebungen. Dieses Projekt hat: {slugs}.'
	],
	'tokens.created': ['Token "{name}" created', 'Jeton « {name} » créé', 'Token «{name}» creado', 'Token „{name}“ erstellt'],
	'tokens.createdNoValue': [
		'Token "{name}" was created, but the server did not return its value. Revoke it and issue another.',
		'Le jeton « {name} » a été créé, mais le serveur n’a pas renvoyé sa valeur. Révoquez-le et émettez-en un autre.',
		'El token «{name}» se creó, pero el servidor no devolvió su valor. Revócalo y emite otro.',
		'Das Token „{name}“ wurde erstellt, aber der Server hat den Wert nicht zurückgegeben. Widerrufe es und stelle ein neues aus.'
	],
	'tokens.revoked': ['Token "{name}" revoked', 'Jeton « {name} » révoqué', 'Token «{name}» revocado', 'Token „{name}“ widerrufen'],
	'tokens.error.load': [
		'Failed to load tokens',
		'Échec du chargement des jetons',
		'No se pudieron cargar los tokens',
		'Tokens konnten nicht geladen werden'
	],
	'tokens.error.create': [
		'Failed to create token',
		'Échec de la création du jeton',
		'No se pudo crear el token',
		'Token konnte nicht erstellt werden'
	],
	'tokens.error.revoke': [
		'Failed to revoke token',
		'Échec de la révocation du jeton',
		'No se pudo revocar el token',
		'Token konnte nicht widerrufen werden'
	]
} as const satisfies Record<string, Row>;
