package config

import (
	"database/sql"
	"log"
	"os"

	"github.com/joho/godotenv"
	_ "github.com/lib/pq"

	"febros16-backend/migrations"
)

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

	if err := database.Ping(); err != nil {
		log.Fatal("Kosa kuunganisha na Database (Ping failed):", err)
	}

	// === AUTO-MIGRATION LOGIC (Raw SQL) ===
	runMigrations(database)

	DB = database
	log.Println("✅ Database imeunganishwa kikamilifu (Raw SQL)!")
}

func runMigrations(db *sql.DB) {
	log.Println("🛠 Inasuka majedwali (Running Migrations) kwenye database...")
	
	// Soma faili la .up.sql kutoka kwenye memory (embedded)
	content, err := migrations.FS.ReadFile("000001_create_users_table.up.sql")
	if err != nil {
		log.Fatalf("KOSA: Imeshindwa kusoma faili la migration: %v", err)
	}

	// Run SQL code
	_, err = db.Exec(string(content))
	if err != nil {
		log.Fatalf("KOSA LAKATILI: Imeshindwa kutengeneza table ya 'users': %v", err)
	}
	
	log.Println("✅ Majedwali yapo tayari (Migrations applied successfully)!")
}
