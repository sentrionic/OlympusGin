package database

import (
	"github.com/go-gormigrate/gormigrate/v2"
	"github.com/sentrionic/OlympusGin/models"
	"gorm.io/gorm"
)

func synchronize(db *gorm.DB) error {
	m := gormigrate.New(db, gormigrate.DefaultOptions, []*gormigrate.Migration{
		{
			ID: "Add Comments",
			Migrate: func(tx *gorm.DB) error {
				if err := tx.AutoMigrate(
					&models.User{},
					&models.Article{},
					&models.Tag{},
					&models.Comment{},
				); err != nil {
					return err
				}
				return nil
			},
			Rollback: func(tx *gorm.DB) error {
				return tx.Migrator().DropTable(
					&models.User{},
					&models.Article{},
					&models.Tag{},
					&models.Comment{},
				)
			},
		},
	})

	return m.Migrate()
}
