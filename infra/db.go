package infra

import (
	"log"
	"os"
	"path/filepath"

	"gorm.io/driver/sqlite"
	"gorm.io/gorm"
	"openflp.com/model"
)

func NewSQLiteDB() (*gorm.DB, error) {
	dbDir := "./data"
	if _, err := os.Stat(dbDir); os.IsNotExist(err) {
		os.Mkdir(dbDir, 0755)
	}

	dbPath := filepath.Join(dbDir, "openflp.db")
	db, err := gorm.Open(sqlite.Open(dbPath), &gorm.Config{})
	if err != nil {
		return nil, err
	}

	// Auto Migrate
	err = db.AutoMigrate(&model.Project{}, &model.Tag{})
	if err != nil {
		log.Printf("Failed to auto migrate: %v", err)
	}

	return db, nil
}
