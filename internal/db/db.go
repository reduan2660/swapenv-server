package db

import (
	"github.com/reduan2660/swapenv-server/internal/models"
	"gorm.io/driver/postgres"
	"gorm.io/gorm"
)

func Connect(databaseURL string) (*gorm.DB, error) {

	db, err := gorm.Open(postgres.Open(databaseURL), &gorm.Config{})
	if err != nil {
		return nil, err
	}

	return db, nil

}

func AutoMigrate(db *gorm.DB) error {
	return db.AutoMigrate(&models.Organization{}, &models.User{})
}
