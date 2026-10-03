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
	
	migrationFiles := []string{
		"000001_create_users_table.up.sql",
		"000002_add_bio_and_profile_to_users.up.sql",
		"000003_create_content_tables.up.sql",
		"000004_create_research_projects.up.sql",
		"000005_add_assessment_to_users.up.sql",
		"000006_add_google_oauth_columns.up.sql",
		"000007_add_email_verification.up.sql",
		"000008_create_user_progress_table.up.sql",
	}

	for _, file := range migrationFiles {
		content, err := migrations.FS.ReadFile(file)
		if err != nil {
			log.Fatalf("KOSA: Imeshindwa kusoma faili la migration %s: %v", file, err)
		}

		// Run SQL code
		_, err = db.Exec(string(content))
		if err != nil {
			log.Fatalf("KOSA LAKATILI: Imeshindwa ku-run migration %s: %v", file, err)
		}
		log.Printf("✅ Migration %s imefanikiwa!", file)
	}
	
	log.Println("✅ Majedwali yapo tayari (Migrations applied successfully)!")
}
