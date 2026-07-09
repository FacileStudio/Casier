package schemas

import "gorm.io/gorm"

func Migrate(db *gorm.DB) error {
	return db.AutoMigrate(
		&User{},
		&Session{},
		&Space{},
		&SpaceMember{},
		&Environment{},
		&Secret{},
	)
}
