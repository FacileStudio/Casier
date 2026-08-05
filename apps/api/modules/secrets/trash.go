package secrets

import (
	"time"

	"github.com/FacileStudio/Casier/apps/api/schemas"
)

// trashRetention is how long a soft-deleted secret stays restorable. It is read
// from the schema so the purge loop and the deadline the API advertises can
// never drift apart.
const trashRetention = time.Duration(schemas.TrashRetentionDays) * 24 * time.Hour

// trashPurgeInterval is how often the background loop looks for expired
// secrets. Retention is measured in days, so an hourly sweep is precise enough
// and cheap.
const trashPurgeInterval = time.Hour

// purgeAt is the moment a trashed secret stops being recoverable. It is what
// the trash listing advertises, so a caller can see the deadline rather than
// having to compute it from a retention constant it does not know.
func purgeAt(deletedAt time.Time) time.Time {
	return deletedAt.Add(trashRetention)
}

// trashCutoff is the deletion timestamp at or before which a trashed secret is
// destroyed for good. The purge loop selects on this rather than on a computed
// per-row deadline so the sweep stays one indexed comparison.
func trashCutoff(now time.Time) time.Time {
	return now.Add(-trashRetention)
}

// purgeDue answers whether a secret deleted at deletedAt has outlived the
// recovery window. The boundary is inclusive: a secret is destroyed the moment
// its deadline is reached, not a tick later.
func purgeDue(deletedAt, now time.Time) bool {
	if deletedAt.IsZero() {
		return false
	}
	return !purgeAt(deletedAt).After(now)
}
