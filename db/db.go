package db

import (
	"goapi-template/models"

	"github.com/glebarez/sqlite"
	"gorm.io/driver/postgres"
	"gorm.io/gorm"
)

func Init(provider string, connectionString string, migrate bool) (*gorm.DB, error) {
	var dialector gorm.Dialector

	switch provider {
	case "postgres":
		dialector = postgres.Open(connectionString)
	case "sqlite":
		dialector = sqlite.Open(connectionString)
	}

	db, err := gorm.Open(dialector, &gorm.Config{TranslateError: true})

	if err != nil {
		return nil, err
	}

	if migrate {
		if err := db.AutoMigrate(&models.Person{}); err != nil {
			return nil, err
		}
	}

	return db, nil
}
