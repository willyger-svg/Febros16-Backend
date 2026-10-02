package config

import (
	"database/sql"
	"log"
	"os"

	"github.com/joho/godotenv"
	_ "github.com/lib/pq" // Driver ya PostgreSQL
)

// DB ni variable itakayotumika na mfumo mzima (sasa ni standard sql.DB)
var DB *sql.DB

func ConnectDB() {
	err := godotenv.Load()
	if err != nil {
		log.Println("Taarifa: Faili la .env halijaonekana, tunatumia env vars.")
	}

	dbURL := os.Getenv("DATABASE_URL")
	if dbURL == "" {
		log.Fatal("KOSA: DATABASE_URL haijapatikana!")
	}

	// Kuunganisha na Database kupitia standard database/sql
	database, err := sql.Open("postgres", dbURL)
	if err != nil {
		log.Fatal("Kosa kufungua Database:", err)
	}

	// Pima kama connection ipo hai
	if err := database.Ping(); err != nil {
		log.Fatal("Kosa kuunganisha na Database (Ping failed):", err)
	}

	DB = database
	log.Println("✅ Database imeunganishwa kikamilifu (Raw SQL)!")
	
	// Kumbuka: Migrations zitaendeshwa kwa kutumia golang-migrate CLI au script, 
	// tumeacha AutoMigrate ya GORM kama maelekezo ya Lead Architect yalivyotaka.
}
