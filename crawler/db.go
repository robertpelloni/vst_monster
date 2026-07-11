package main

import (
	"database/sql"
	"encoding/json"
	"fmt"
	"log"
	"os"

	_ "github.com/lib/pq"
	"github.com/joho/godotenv"
	"github.com/robertpelloni/vst_monster/crawler/parser"
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

	return db
}

func UpsertPlugin(db *sql.DB, p parser.StandardPlugin, hash string) error {
	if db == nil {
		return nil // Graceful degradation if DB is not connected
	}

	metadataBytes, _ := json.Marshal(p.Metadata)
	if metadataBytes == nil || string(metadataBytes) == "null" {
		metadataBytes = []byte("{}")
	}

	licenseModel := p.License
	if licenseModel == "" {
		licenseModel = "free"
	}

	// 1. Upsert Plugin
	var pluginID string
	err := db.QueryRow(`
		INSERT INTO plugins (name, developer, license_model, updated_at)
		VALUES ($1, $2, $3, NOW())
		ON CONFLICT (name, developer) DO UPDATE SET
			license_model = EXCLUDED.license_model,
			updated_at = NOW()
		RETURNING id;
	`, p.Name, p.Developer, licenseModel).Scan(&pluginID)

	if err != nil {
		err = db.QueryRow(`SELECT id FROM plugins WHERE name = $1 AND developer = $2`, p.Name, p.Developer).Scan(&pluginID)
		if err != nil {
			return fmt.Errorf("error resolving plugin id: %w", err)
		}
	}

	version := p.Version
	if version == "" {
		version = "1.0.0"
	}

	// 2. Upsert Release
	var releaseID string
	err = db.QueryRow(`
		INSERT INTO plugin_releases (plugin_id, version, release_date)
		VALUES ($1, $2, NOW())
		ON CONFLICT (plugin_id, version) DO UPDATE SET
			release_date = EXCLUDED.release_date
		RETURNING id;
	`, pluginID, version).Scan(&releaseID)

	if err != nil {
		err = db.QueryRow(`SELECT id FROM plugin_releases WHERE plugin_id = $1 AND version = $2`, pluginID, version).Scan(&releaseID)
		if err != nil {
			return fmt.Errorf("error resolving release id: %w", err)
		}
	}

	// Determine generic platform (mock for now, assume universal or windows)
	platform := "windows"

	// Determine strategy
	strategy := "extract_binaries"

	// 3. Upsert Distribution
	// We lack a robust unique constraint on distributions, but checking manually prevents naive duplicates
	var distID string
	err = db.QueryRow(`SELECT id FROM plugin_distributions WHERE release_id = $1 AND platform = $2 AND architecture = $3`, releaseID, platform, "x86_64").Scan(&distID)
	if err == sql.ErrNoRows {
		_, err = db.Exec(`
			INSERT INTO plugin_distributions (release_id, platform, architecture, download_url, sha256_hash, strategy, extraction_rules)
			VALUES ($1, $2, $3, $4, $5, $6, $7)
		`, releaseID, platform, "x86_64", p.DownloadURL, hash, strategy, metadataBytes)
	} else if err == nil {
		_, err = db.Exec(`
			UPDATE plugin_distributions SET download_url = $1, sha256_hash = $2, extraction_rules = $3 WHERE id = $4
		`, p.DownloadURL, hash, metadataBytes, distID)
	}

	if err != nil {
		log.Printf("Distribution insertion note: %v", err)
	}

	return nil
}
