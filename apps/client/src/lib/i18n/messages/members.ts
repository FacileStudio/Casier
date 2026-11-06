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
	'tokens.new': ['New token', 'Nouveau jeton', 'Nuevo token', 'Neues Token'],
	'tokens.empty.title': ['No API tokens', 'Aucun jeton d’API', 'Sin tokens de API', 'Keine API-Tokens'],
	'tokens.empty.description': [
		'Create one to give a pipeline or a machine scoped access to this project.',
		'Créez-en un pour donner à un pipeline ou à une machine un accès limité à ce projet.',
		'Crea uno para dar a un pipeline o a una máquina acceso acotado a este proyecto.',
		'Erstelle eines, um einer Pipeline oder Maschine eingeschränkten Zugriff auf dieses Projekt zu geben.'
	],
	'tokens.create': ['Create a token', 'Créer un jeton', 'Crear un token', 'Token erstellen'],
	'tokens.col.name': ['Name', 'Nom', 'Nombre', 'Name'],
	'tokens.col.owner': ['Owner', 'Propriétaire', 'Propietario', 'Eigentümer'],
	'tokens.col.status': ['Status', 'Statut', 'Estado', 'Status'],
	'tokens.col.permissions': ['Permissions', 'Permissions', 'Permisos', 'Berechtigungen'],
	'tokens.col.environments': ['Environments', 'Environnements', 'Entornos', 'Umgebungen'],
	'tokens.col.lastUsed': ['Last used', 'Dernière utilisation', 'Último uso', 'Zuletzt verwendet'],
	'tokens.col.expires': ['Expires', 'Expire', 'Caduca', 'Läuft ab'],
	'tokens.status.revoked': ['Revoked {date}', 'Révoqué le {date}', 'Revocado el {date}', 'Widerrufen am {date}'],
	'tokens.status.expired': ['Expired', 'Expiré', 'Caducado', 'Abgelaufen'],
	'tokens.status.active': ['Active', 'Actif', 'Activo', 'Aktiv'],
	'tokens.revoke': ['Revoke', 'Révoquer', 'Revocar', 'Widerrufen'],
	'tokens.drawer.title': ['New API token', 'Nouveau jeton d’API', 'Nuevo token de API', 'Neues API-Token'],
	'tokens.copy.title': ['Copy it now', 'Copiez-le maintenant', 'Cópialo ahora', 'Jetzt kopieren'],
	'tokens.copy.body': [
		'This is the only time the token is shown. Casier stores a hash of it, so it cannot be shown again — losing it means issuing another.',
		'C’est la seule fois où le jeton est affiché. Casier n’en conserve qu’un hachage : il ne peut plus être montré, et le perdre oblige à en émettre un autre.',
		'Esta es la única vez que se muestra el token. Casier guarda un hash, así que no se puede volver a mostrar: perderlo obliga a emitir otro.',
		'Das Token wird nur dieses eine Mal angezeigt. Casier speichert nur einen Hash und kann es nicht erneut anzeigen — geht es verloren, musst du ein neues ausstellen.'
	],
	'tokens.copy.label': ['Token', 'Jeton', 'Token', 'Token'],
	'tokens.copy.helper': [
		'Store it in your CI secret store, or hand it to the CLI as CASIER_TOKEN.',
		'Conservez-le dans le coffre à secrets de votre CI, ou donnez-le à la CLI sous CASIER_TOKEN.',
		'Guárdalo en el almacén de secretos de tu CI, o pásalo a la CLI como CASIER_TOKEN.',
		'Lege es im Secret-Speicher deiner CI ab oder übergib es der CLI als CASIER_TOKEN.'
	],
	'tokens.name.helper': [
		'Where the token will live — a machine, a pipeline, a script.',
		'Où le jeton vivra : une machine, un pipeline, un script.',
		'Dónde vivirá el token: una máquina, un pipeline, un script.',
		'Wo das Token liegen wird — eine Maschine, eine Pipeline, ein Skript.'
	],
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
	'tokens.expires.label': ['Expires in', 'Expire dans', 'Caduca en', 'Läuft ab in'],
	'tokens.expires.helper': [
		'A machine credential that never lapses is the one nobody remembers to rotate.',
		'Un identifiant machine qui n’expire jamais est celui que personne ne pense à renouveler.',
		'Una credencial de máquina que nunca caduca es la que nadie recuerda rotar.',
		'Ein Maschinenzugang, der nie abläuft, ist der, den niemand rotiert.'
	],
	'tokens.expires.30': ['30 days', '30 jours', '30 días', '30 Tage'],
	'tokens.expires.90': ['90 days', '90 jours', '90 días', '90 Tage'],
	'tokens.expires.365': ['1 year', '1 an', '1 año', '1 Jahr'],
	'tokens.expires.730': ['2 years', '2 ans', '2 años', '2 Jahre'],
	'tokens.createSubmit': ['Create token', 'Créer le jeton', 'Crear token', 'Token erstellen'],
	'tokens.created': ['Token "{name}" created', 'Jeton « {name} » créé', 'Token «{name}» creado', 'Token „{name}“ erstellt'],
	'tokens.createdNoValue': [
		'Token "{name}" was created, but the server did not return its value. Revoke it and issue another.',
		'Le jeton « {name} » a été créé, mais le serveur n’a pas renvoyé sa valeur. Révoquez-le et émettez-en un autre.',
		'El token «{name}» se creó, pero el servidor no devolvió su valor. Revócalo y emite otro.',
		'Das Token „{name}“ wurde erstellt, aber der Server hat den Wert nicht zurückgegeben. Widerrufe es und stelle ein neues aus.'
	],
	'tokens.revoke.title': ['Revoke token', 'Révoquer le jeton', 'Revocar token', 'Token widerrufen'],
	'tokens.revoke.message': [
		'"{name}" stops working immediately, and any CI pipeline or CLI still using it starts failing. It cannot be un-revoked. The row stays listed so the audit log still names what it changed.',
		'« {name} » cesse de fonctionner immédiatement, et tout pipeline CI ou CLI qui l’utilise encore échoue. Une révocation est irréversible. La ligne reste listée pour que le journal d’audit nomme toujours ce qu’il a modifié.',
		'«{name}» deja de funcionar de inmediato, y cualquier pipeline de CI o CLI que lo siga usando empezará a fallar. No se puede deshacer la revocación. La fila sigue listada para que el registro de auditoría nombre lo que cambió.',
		'„{name}“ funktioniert sofort nicht mehr, und jede CI-Pipeline oder CLI, die es noch nutzt, schlägt fehl. Der Widerruf lässt sich nicht rückgängig machen. Die Zeile bleibt gelistet, damit das Audit-Log weiter nennt, was es geändert hat.'
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
