package main

import (
	"database/sql"
	"fmt"
	"log"
	"os"

	_ "github.com/lib/pq"
	"github.com/joho/godotenv"
)

func InitDB() *sql.DB {
	_ = godotenv.Load() // Load .env file if it exists

	host := os.Getenv("PGHOST")
	port := os.Getenv("PGPORT")
	user := os.Getenv("PGUSER")
	password := os.Getenv("PGPASSWORD")
	dbname := os.Getenv("PGDATABASE")

	if host == "" || port == "" || user == "" || password == "" || dbname == "" {
		log.Println("Database environment variables not fully set. Skipping DB connection.")
		return nil
	}

	psqlInfo := fmt.Sprintf("host=%s port=%s user=%s password=%s dbname=%s sslmode=disable",
		host, port, user, password, dbname)

	db, err := sql.Open("postgres", psqlInfo)
	if err != nil {
		log.Printf("Error opening database: %v\n", err)
		return nil
	}

	err = db.Ping()
	if err != nil {
		log.Printf("Error connecting to database: %v\n", err)
		return nil
	}

	log.Println("Successfully connected to database")

	// Ensure the table exists and has the unique constraint
	createTableSQL := `
	CREATE TABLE IF NOT EXISTS plugins (
		id SERIAL PRIMARY KEY,
		name VARCHAR(255) NOT NULL,
		developer VARCHAR(255) NOT NULL,
		download_url TEXT,
		version VARCHAR(50),
		platform VARCHAR(50),
		hash VARCHAR(64),
		metadata JSONB,
		created_at TIMESTAMP WITH TIME ZONE DEFAULT CURRENT_TIMESTAMP,
		updated_at TIMESTAMP WITH TIME ZONE DEFAULT CURRENT_TIMESTAMP,
		UNIQUE(name, developer)
	);`

	_, err = db.Exec(createTableSQL)
	if err != nil {
		log.Printf("Error creating plugins table: %v\n", err)
	}

	return db
}

func UpsertPlugin(db *sql.DB, p ScrapedPlugin, hash string) error {
	if db == nil {
		return nil // Graceful degradation if DB is not connected
	}

	query := `
		INSERT INTO plugins (name, developer, download_url, version, platform, hash, metadata, updated_at)
		VALUES ($1, $2, $3, $4, $5, $6, $7, CURRENT_TIMESTAMP)
		ON CONFLICT (name, developer) DO UPDATE SET
			download_url = EXCLUDED.download_url,
			version = EXCLUDED.version,
			platform = EXCLUDED.platform,
			hash = EXCLUDED.hash,
			metadata = EXCLUDED.metadata,
			updated_at = CURRENT_TIMESTAMP;
	`

	// If Metadata is nil, we insert an empty JSON object to satisfy JSONB expectations
	metadata := p.Metadata
	if metadata == nil {
		metadata = []byte("{}")
	}

	_, err := db.Exec(query, p.Name, p.Developer, p.DownloadURL, p.Version, p.Platform, hash, metadata)
	if err != nil {
		return fmt.Errorf("error upserting plugin %s by %s: %w", p.Name, p.Developer, err)
	}

	return nil
}
