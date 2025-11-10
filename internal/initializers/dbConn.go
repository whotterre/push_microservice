package initializers

import (
	"log"

	"gorm.io/driver/postgres"
	"gorm.io/gorm"
)

func ConnectToDB(connString string) (*gorm.DB, error) {
	db, err := gorm.Open(postgres.Open(connString))
	if err != nil {
		log.Printf("Failed to connect to the PostgreSQL database because %s", err)
		return nil, err
	}
	log.Println("Successfully connected to the PostgreSQL database")
	return db, nil
}

func PerformMigrations(db *gorm.DB) error {
	err := db.AutoMigrate()
	if err != nil {
		log.Printf("Failed to perform migrations because: %s", err.Error())
	}
	log.Println("Successfully performed migrations")
	return nil
}