import type { Row } from '../row';

/** The secrets table of an environment and the drawers around it. */
export const secrets = {
	'sec.section': ['Secrets', 'Secrets', 'Secretos', 'Secrets'],
	'sec.protected': [
		'{name} is protected — only an owner or an admin can write to it.',
		'{name} est protégé : seul un propriétaire ou un administrateur peut y écrire.',
		'{name} está protegido: solo un propietario o un administrador puede escribir en él.',
		'{name} ist geschützt — nur Eigentümer oder Admins dürfen darin schreiben.'
	],
	'sec.overdue': [
		{
			one: '{count} secret has gone more than {days} days without being rotated.',
			other: '{count} secrets have gone more than {days} days without being rotated.'
		},
		{
			one: '{count} secret n’a pas été renouvelé depuis plus de {days} jours.',
			other: '{count} secrets n’ont pas été renouvelés depuis plus de {days} jours.'
		},
		{
			one: '{count} secreto lleva más de {days} días sin rotarse.',
			other: '{count} secretos llevan más de {days} días sin rotarse.'
		},
		{
			one: '{count} Secret wurde seit mehr als {days} Tagen nicht rotiert.',
			other: '{count} Secrets wurden seit mehr als {days} Tagen nicht rotiert.'
		}
	],
	'sec.dueSoon': [
		{ one: '{count} is due for rotation shortly.', other: '{count} are due for rotation shortly.' },
		{ one: '{count} sera bientôt à renouveler.', other: '{count} seront bientôt à renouveler.' },
		{ one: '{count} debe rotarse pronto.', other: '{count} deben rotarse pronto.' },
		{ one: '{count} muss bald rotiert werden.', other: '{count} müssen bald rotiert werden.' }
	],
	'sec.review': ['Review them', 'Les examiner', 'Revisarlos', 'Prüfen'],
	'sec.showAll': ['Show all secrets', 'Afficher tous les secrets', 'Mostrar todos los secretos', 'Alle Secrets anzeigen'],
	'sec.filter.placeholder': ['Filter by key or tag', 'Filtrer par clé ou étiquette', 'Filtrar por clave o etiqueta', 'Nach Schlüssel oder Tag filtern'],
	'sec.filter.aria': ['Filter secrets', 'Filtrer les secrets', 'Filtrar secretos', 'Secrets filtern'],
	'sec.trash': ['Trash', 'Corbeille', 'Papelera', 'Papierkorb'],
	'sec.export': ['Export', 'Exporter', 'Exportar', 'Exportieren'],
	'sec.exporting': ['Exporting…', 'Exportation…', 'Exportando…', 'Wird exportiert…'],
	'sec.import': ['Import', 'Importer', 'Importar', 'Importieren'],
	'sec.add': ['Add secret', 'Ajouter un secret', 'Añadir secreto', 'Secret hinzufügen'],
	'sec.empty.title': ['No secrets in {name}', 'Aucun secret dans {name}', 'No hay secretos en {name}', 'Keine Secrets in {name}'],
	'sec.empty.description': [
		'Secrets are encrypted at rest and only visible to members of this project.',
		'Les secrets sont chiffrés au repos et visibles uniquement par les membres de ce projet.',
		'Los secretos se cifran en reposo y solo los ven los miembros de este proyecto.',
		'Secrets sind verschlüsselt gespeichert und nur für Mitglieder dieses Projekts sichtbar.'
	],
	'sec.empty.first': ['Add your first secret', 'Ajouter votre premier secret', 'Añade tu primer secreto', 'Erstes Secret hinzufügen'],
	'sec.noMatches': ['No matches', 'Aucun résultat', 'Sin coincidencias', 'Keine Treffer'],
	'sec.noMatches.stale': [
		'Every secret in this environment is within the rotation policy.',
		'Tous les secrets de cet environnement respectent la politique de rotation.',
		'Todos los secretos de este entorno cumplen la política de rotación.',
		'Alle Secrets in dieser Umgebung entsprechen der Rotationsrichtlinie.'
	],
	'sec.noMatches.search': [
		'No secret matches "{search}".',
		'Aucun secret ne correspond à « {search} ».',
		'Ningún secreto coincide con «{search}».',
		'Kein Secret passt zu „{search}“.'
	],
	'sec.clearFilters': ['Clear filters', 'Effacer les filtres', 'Quitar filtros', 'Filter zurücksetzen'],
	'sec.col.key': ['Key', 'Clé', 'Clave', 'Schlüssel'],
	'sec.col.value': ['Value', 'Valeur', 'Valor', 'Wert'],
	'sec.col.version': ['Version', 'Version', 'Versión', 'Version'],
	'sec.col.rotated': ['Last rotated', 'Dernière rotation', 'Última rotación', 'Zuletzt rotiert'],
	'sec.rotatedOn': ['rotated {date}', 'renouvelé le {date}', 'rotado el {date}', 'rotiert am {date}'],
	'sec.badge.rotate': ['Rotate', 'À renouveler', 'Rotar', 'Rotieren'],
	'sec.badge.rotateAge': ['Rotate · {age}d', 'À renouveler · {age} j', 'Rotar · {age} d', 'Rotieren · {age} T'],
	'sec.badge.dueSoon': ['Due soon', 'Bientôt', 'Pronto', 'Bald fällig'],
	'sec.badge.dueSoonAge': ['Due soon · {age}d', 'Bientôt · {age} j', 'Pronto · {age} d', 'Bald fällig · {age} T'],
	'sec.copied': ['Copied {key} to clipboard', '{key} copié dans le presse-papiers', '{key} copiado al portapapeles', '{key} in die Zwischenablage kopiert'],
	'sec.copy.aria': ['Copy {key} to clipboard', 'Copier {key} dans le presse-papiers', 'Copiar {key} al portapapeles', '{key} in die Zwischenablage kopieren'],
	'sec.copy': ['Copy', 'Copier', 'Copiar', 'Kopieren'],
	'sec.hide.aria': ['Hide {key}', 'Masquer {key}', 'Ocultar {key}', '{key} ausblenden'],
	'sec.reveal.aria': ['Reveal {key}', 'Révéler {key}', 'Mostrar {key}', '{key} anzeigen'],
	'sec.hide': ['Hide', 'Masquer', 'Ocultar', 'Ausblenden'],
	'sec.reveal': ['Reveal', 'Révéler', 'Mostrar', 'Anzeigen'],
	'sec.history.aria': ['History of {key}', 'Historique de {key}', 'Historial de {key}', 'Verlauf von {key}'],
	'sec.history': ['History', 'Historique', 'Historial', 'Verlauf'],
	'sec.edit.aria': ['Edit {key}', 'Modifier {key}', 'Editar {key}', '{key} bearbeiten'],
	'sec.edit': ['Edit', 'Modifier', 'Editar', 'Bearbeiten'],
	'sec.delete.aria': ['Delete {key}', 'Supprimer {key}', 'Eliminar {key}', '{key} löschen'],
	'sec.delete': ['Delete', 'Supprimer', 'Eliminar', 'Löschen'],
	'sec.fallback': ['secret', 'secret', 'secreto', 'Secret'],

	'sec.add.title': ['Add secret to {name}', 'Ajouter un secret à {name}', 'Añadir secreto a {name}', 'Secret zu {name} hinzufügen'],
	'sec.add.key': ['Key', 'Clé', 'Clave', 'Schlüssel'],
	'sec.add.value': ['Value', 'Valeur', 'Valor', 'Wert'],
	'sec.add.valuePlaceholder': ['Secret value', 'Valeur du secret', 'Valor del secreto', 'Wert des Secrets'],
	'sec.add.tags': ['Tags', 'Étiquettes', 'Etiquetas', 'Tags'],
	'sec.add.tagsHelper': ['Comma-separated, optional', 'Séparées par des virgules, facultatif', 'Separadas por comas, opcional', 'Kommagetrennt, optional'],
	'sec.key.invalid': [
		'Keys cannot contain whitespace or "="',
		'Les clés ne peuvent contenir ni espace ni « = »',
		'Las claves no pueden contener espacios ni «=»',
		'Schlüssel dürfen weder Leerzeichen noch „=“ enthalten'
	],
	'sec.key.exists': [
		'This key already exists in this environment',
		'Cette clé existe déjà dans cet environnement',
		'Esta clave ya existe en este entorno',
		'Dieser Schlüssel existiert in dieser Umgebung bereits'
	],
	'sec.added': ['Secret {key} added to {name}', 'Secret {key} ajouté à {name}', 'Secreto {key} añadido a {name}', 'Secret {key} zu {name} hinzugefügt'],

	'sec.trash.title': ['Trash — {name}', 'Corbeille — {name}', 'Papelera — {name}', 'Papierkorb — {name}'],
	'sec.trash.empty.title': ['Nothing in the trash', 'La corbeille est vide', 'Nada en la papelera', 'Der Papierkorb ist leer'],
	'sec.trash.empty.description': [
		'Deleted secrets stay here for 30 days before they are destroyed.',
		'Les secrets supprimés restent ici 30 jours avant d’être détruits.',
		'Los secretos eliminados permanecen aquí 30 días antes de destruirse.',
		'Gelöschte Secrets bleiben hier 30 Tage, bevor sie vernichtet werden.'
	],
	'sec.trash.note': [
		'A deleted key stays reserved until it is restored or purged, so it cannot be recreated from scratch while it sits here.',
		'Une clé supprimée reste réservée jusqu’à sa restauration ou sa purge : elle ne peut pas être recréée de zéro tant qu’elle est ici.',
		'Una clave eliminada sigue reservada hasta que se restaura o se purga, así que no se puede recrear desde cero mientras esté aquí.',
		'Ein gelöschter Schlüssel bleibt reserviert, bis er wiederhergestellt oder endgültig entfernt wird, und lässt sich solange nicht neu anlegen.'
	],
	'sec.trash.deleted': [
		'deleted {deleted} · purged {purged}',
		'supprimé le {deleted} · purgé le {purged}',
		'eliminado el {deleted} · purgado el {purged}',
		'gelöscht am {deleted} · entfernt am {purged}'
	],
	'sec.restore': ['Restore', 'Restaurer', 'Restaurar', 'Wiederherstellen'],
	'sec.restored': ['Secret {key} restored', 'Secret {key} restauré', 'Secreto {key} restaurado', 'Secret {key} wiederhergestellt'],

	'sec.edit.title': ['Edit {key}', 'Modifier {key}', 'Editar {key}', '{key} bearbeiten'],
	'sec.edit.helper': ['Saving creates a new version', 'L’enregistrement crée une nouvelle version', 'Guardar crea una nueva versión', 'Speichern erzeugt eine neue Version'],
	'sec.updated': ['Secret {key} updated', 'Secret {key} mis à jour', 'Secreto {key} actualizado', 'Secret {key} aktualisiert'],

	'sec.history.title': ['History of {key}', 'Historique de {key}', 'Historial de {key}', 'Verlauf von {key}'],
	'sec.history.empty': ['No previous versions', 'Aucune version précédente', 'Sin versiones anteriores', 'Keine früheren Versionen'],
	'sec.history.current': ['current', 'actuelle', 'actual', 'aktuell'],
	'sec.rollback': ['Roll back', 'Revenir en arrière', 'Revertir', 'Zurücksetzen'],
	'sec.rollingBack': ['Rolling back…', 'Retour en arrière…', 'Revirtiendo…', 'Wird zurückgesetzt…'],
	'sec.rolledBack': [
		'Rolled {key} back to v{version}',
		'{key} ramené à la v{version}',
		'{key} revertido a la v{version}',
		'{key} auf v{version} zurückgesetzt'
	],

	'sec.import.title': ['Import into {name}', 'Importer dans {name}', 'Importar en {name}', 'In {name} importieren'],
	'sec.import.label': ['.env content', 'Contenu du .env', 'Contenido del .env', '.env-Inhalt'],
	'sec.import.helper': [
		'Paste KEY=value lines; existing keys are updated',
		'Collez des lignes CLÉ=valeur ; les clés existantes sont mises à jour',
		'Pega líneas CLAVE=valor; las claves existentes se actualizan',
		'Füge Zeilen im Format SCHLÜSSEL=Wert ein; vorhandene Schlüssel werden aktualisiert'
	],
	'sec.importing': ['Importing…', 'Importation…', 'Importando…', 'Wird importiert…'],
	'sec.imported': [
		'Imported: {created} created, {updated} updated, {skipped} skipped',
		'Importé : {created} créés, {updated} mis à jour, {skipped} ignorés',
		'Importado: {created} creados, {updated} actualizados, {skipped} omitidos',
		'Importiert: {created} erstellt, {updated} aktualisiert, {skipped} übersprungen'
	],
	'sec.exported': ['Exported {name} as .env', '{name} exporté en .env', '{name} exportado como .env', '{name} als .env exportiert'],

	'sec.delete.title': ['Delete secret', 'Supprimer le secret', 'Eliminar secreto', 'Secret löschen'],
	'sec.delete.message': [
		'This moves {key} out of the {name} environment and into the trash, where it can be restored for 30 days before it is destroyed for good. Anything reading it will break immediately. The key stays reserved until it is restored or purged.',
		'Cela déplace {key} hors de l’environnement {name} vers la corbeille, où il peut être restauré pendant 30 jours avant d’être détruit définitivement. Tout ce qui le lit cessera immédiatement de fonctionner. La clé reste réservée jusqu’à sa restauration ou sa purge.',
		'Esto mueve {key} fuera del entorno {name} a la papelera, donde se puede restaurar durante 30 días antes de destruirse para siempre. Todo lo que lo lea dejará de funcionar de inmediato. La clave sigue reservada hasta que se restaure o se purgue.',
		'{key} wird aus der Umgebung {name} in den Papierkorb verschoben und kann dort 30 Tage wiederhergestellt werden, bevor es endgültig vernichtet wird. Alles, was es liest, bricht sofort. Der Schlüssel bleibt reserviert, bis er wiederhergestellt oder entfernt wird.'
	],
	'sec.delete.confirm': ['Delete secret', 'Supprimer le secret', 'Eliminar secreto', 'Secret löschen'],
	'sec.deleted': ['Secret {key} deleted', 'Secret {key} supprimé', 'Secreto {key} eliminado', 'Secret {key} gelöscht'],

	'sec.error.load': ['Failed to load secrets', 'Échec du chargement des secrets', 'No se pudieron cargar los secretos', 'Secrets konnten nicht geladen werden'],
	'sec.error.reveal': ['Failed to reveal {key}', 'Échec de la révélation de {key}', 'No se pudo mostrar {key}', '{key} konnte nicht angezeigt werden'],
	'sec.error.trash': ['Failed to load the trash', 'Échec du chargement de la corbeille', 'No se pudo cargar la papelera', 'Papierkorb konnte nicht geladen werden'],
	'sec.error.restore': ['Failed to restore {key}', 'Échec de la restauration de {key}', 'No se pudo restaurar {key}', '{key} konnte nicht wiederhergestellt werden'],
	'sec.error.add': ['Failed to add secret', 'Échec de l’ajout du secret', 'No se pudo añadir el secreto', 'Secret konnte nicht hinzugefügt werden'],
	'sec.error.read': ['Failed to read {key}', 'Échec de la lecture de {key}', 'No se pudo leer {key}', '{key} konnte nicht gelesen werden'],
	'sec.error.update': ['Failed to update secret', 'Échec de la mise à jour du secret', 'No se pudo actualizar el secreto', 'Secret konnte nicht aktualisiert werden'],
	'sec.error.delete': ['Failed to delete secret', 'Échec de la suppression du secret', 'No se pudo eliminar el secreto', 'Secret konnte nicht gelöscht werden'],
	'sec.error.versions': ['Failed to load versions', 'Échec du chargement des versions', 'No se pudieron cargar las versiones', 'Versionen konnten nicht geladen werden'],
	'sec.error.rollback': ['Rollback failed', 'Le retour en arrière a échoué', 'Falló la reversión', 'Zurücksetzen fehlgeschlagen'],
	'sec.error.import': ['Import failed', 'L’importation a échoué', 'Falló la importación', 'Import fehlgeschlagen'],
	'sec.error.export': ['Export failed', 'L’exportation a échoué', 'Falló la exportación', 'Export fehlgeschlagen']
} as const satisfies Record<string, Row>;
