package config

import (
	"log"
	"os"

	"github.com/joho/godotenv"
	"gorm.io/driver/postgres"
	"gorm.io/gorm"
	
	"febros16-backend/internal/models"
)

// DB ni variable itakayotumika na mfumo mzima
var DB *gorm.DB

func ConnectDB() {
	err := godotenv.Load()
	if err != nil {
		log.Println("Taarifa: Faili la .env halijaonekana, tunatumia mfumo wa mazingira (env vars) wa kawaida.")
	}

	dbURL := os.Getenv("DATABASE_URL")
	if dbURL == "" {
		log.Fatal("KOSA: DATABASE_URL haijapatikana!")
	}

	// Kuunganisha na Database kupitia GORM
	database, err := gorm.Open(postgres.Open(dbURL), &gorm.Config{})
	if err != nil {
		log.Fatal("Kosa kufungua Database:", err)
	}

	// Tengeneza au update majedwali (Auto-Migrate)
	log.Println("Inafanya Auto-Migrate ya database schemas...")
	err = database.AutoMigrate(
		&models.User{},
		&models.Profile{},
	)
	if err != nil {
		log.Fatal("Kosa kutengeneza majedwali:", err)
	}

	DB = database
	log.Println("✅ Database imeunganishwa na Majedwali yapo tayari!")
}
