// Package people is the one shape a user takes in an API response, so every
// surface that names a person carries the same fields for the client's Avatar:
// a picture when there is one, and a name and email to draw initials from when
// there is not.
package people

import (
	"context"

	"gorm.io/gorm"
)

// Columns selects a person from the users table joined as `u`. Scanning it into
// a response with Email, Name and AvatarURL fields fills them by name, which is
// how the flat rows (the caller, members, ACL grantees) carry the same fields
// as a nested Person.
const Columns = "u.email, u.name, u.avatar_url"

// Person is a user as the client renders them. AvatarURL is the identity
// provider's stable picture URL, or "" for an account that has none.
type Person struct {
	Email     string `json:"email"`
	Name      string `json:"name"`
	AvatarURL string `json:"avatar_url"`
}

// Resolve looks up the people behind user ids stored as text, the way audit
// rows, secret versions and the trash record who acted. An id that names no
// user, such as one whose account was deleted, is absent from the map, so the
// caller renders it as unknown rather than as somebody else.
func Resolve(ctx context.Context, orm *gorm.DB, ids []string) (map[string]Person, error) {
	found := make(map[string]Person, len(ids))
	if len(ids) == 0 {
		return found, nil
	}
	var rows []struct {
		ID string
		Person
	}
	err := orm.WithContext(ctx).
		Table("users u").
		Select("u.id::text AS id, "+Columns).
		Where("u.id::text IN ?", ids).
		Scan(&rows).Error
	if err != nil {
		return nil, err
	}
	for _, row := range rows {
		found[row.ID] = row.Person
	}
	return found, nil
}

// Lookup returns the person behind id from a Resolve result, or nil when the
// id resolved to nobody, so the JSON carries null instead of an empty person.
func Lookup(found map[string]Person, id string) *Person {
	person, ok := found[id]
	if !ok {
		return nil
	}
	return &person
}
