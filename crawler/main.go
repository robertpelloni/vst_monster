package main

import (
	"crypto/sha256"
	"database/sql"
	"encoding/hex"
	"fmt"
	"io"
	"log"
	"net/http"
	"os"
	"strings"
	"time"

	"github.com/gocolly/colly/v2"
	_ "github.com/lib/pq"
)

var db *sql.DB

func initDB() {
	host := os.Getenv("PGHOST")
	if host == "" {
		host = "localhost"
	}
	port := os.Getenv("PGPORT")
	if port == "" {
		port = "5432"
	}
	user := os.Getenv("PGUSER")
	if user == "" {
		user = "postgres"
	}
	password := os.Getenv("PGPASSWORD")
	if password == "" {
		password = "postgres"
	}
	dbname := os.Getenv("PGDATABASE")
	if dbname == "" {
		dbname = "vst_monster"
	}

	psqlInfo := fmt.Sprintf("host=%s port=%s user=%s password=%s dbname=%s sslmode=disable",
		host, port, user, password, dbname)

	var err error
	db, err = sql.Open("postgres", psqlInfo)
	if err != nil {
		log.Printf("Warning: Error opening database connection: %v\n", err)
		return
	}

	err = db.Ping()
	if err != nil {
		log.Printf("Warning: Error pinging database: %v\n", err)
		db = nil
		return
	}
	log.Println("Successfully connected to the PostgreSQL database!")
}

type Plugin struct {
	ID           string    `json:"id"`
	Name         string    `json:"name"`
	Developer    string    `json:"developer"`
	LicenseModel string    `json:"license_model"`
	CreatedAt    time.Time `json:"created_at"`
	UpdatedAt    time.Time `json:"updated_at"`
}

type PluginRelease struct {
	ID          string    `json:"id"`
	PluginID    string    `json:"plugin_id"`
	Version     string    `json:"version"`
	ReleaseDate time.Time `json:"release_date"`
}

type PluginDistribution struct {
	ID              string `json:"id"`
	ReleaseID       string `json:"release_id"`
	Platform        string `json:"platform"`
	Architecture    string `json:"architecture"`
	DownloadURL     string `json:"download_url"`
	Sha256Hash      string `json:"sha256_hash"`
	Strategy        string `json:"strategy"`
	ExtractionRules string `json:"extraction_rules"`
	IsActive        bool   `json:"is_active"`
}

type ScrapedPlugin struct {
	Name        string `json:"name"`
	Developer   string `json:"developer"`
	DownloadURL string `json:"download_url"`
	Version     string `json:"version"`
	Platform    string `json:"platform"` // windows, macos, linux
}

// InitCrawler initializes the colly collector with standard settings.
func InitCrawler() *colly.Collector {
	c := colly.NewCollector(
		colly.Async(true),
		colly.UserAgent("Mozilla/5.0 (Windows NT 10.0; Win64; x64) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/91.0.4472.124 Safari/537.36"),
	)

	c.SetRequestTimeout(60 * time.Second)

	c.OnRequest(func(r *colly.Request) {
		log.Println("Visiting", r.URL.String())
	})

	c.OnError(func(_ *colly.Response, err error) {
		log.Println("Something went wrong:", err)
	})

	return c
}

// CalculateSHA256 downloads the file from the given URL to a temporary file,
// calculates its SHA256 hash, and cleans up the temporary file.
func CalculateSHA256(url string) (string, error) {
	// Add user agent to bypass some simple blocks
	client := &http.Client{}
	req, err := http.NewRequest("GET", url, nil)
	if err != nil {
		return "", err
	}
	req.Header.Set("User-Agent", "Mozilla/5.0 (Windows NT 10.0; Win64; x64)")

	resp, err := client.Do(req)
	if err != nil {
		return "", err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return "", fmt.Errorf("bad status: %s", resp.Status)
	}

	// Create a temporary file to store the downloaded binary
	tmpFile, err := os.CreateTemp("", "vst-download-*.bin")
	if err != nil {
		return "", fmt.Errorf("failed to create temp file: %w", err)
	}
	defer os.Remove(tmpFile.Name()) // Clean up after we're done
	defer tmpFile.Close()

	// Download the content into the temporary file
	_, err = io.Copy(tmpFile, resp.Body)
	if err != nil {
		return "", fmt.Errorf("failed to write to temp file: %w", err)
	}

	// Rewind the file pointer to the beginning to calculate the hash
	_, err = tmpFile.Seek(0, 0)
	if err != nil {
		return "", fmt.Errorf("failed to seek temp file: %w", err)
	}

	// Calculate the SHA256 hash from the temporary file
	hasher := sha256.New()
	if _, err := io.Copy(hasher, tmpFile); err != nil {
		return "", fmt.Errorf("failed to calculate hash: %w", err)
	}

	return hex.EncodeToString(hasher.Sum(nil)), nil
}

func persistPlugins(plugins []ScrapedPlugin) {
	if db == nil {
		log.Println("Database connection is not available, skipping persistence")
		return
	}

	for _, p := range plugins {
		var pluginID string
		// 1. Insert or update the Plugin. Use CTE to handle unique constraints since ID is UUID default
		err := db.QueryRow(`
			WITH e AS (
				INSERT INTO plugins (name, developer, license_model)
				VALUES ($1, $2, 'free')
				ON CONFLICT (name, developer) DO UPDATE SET updated_at = NOW()
				RETURNING id
			)
			SELECT id FROM e
			UNION ALL
			SELECT id FROM plugins WHERE name = $1 AND developer = $2
			LIMIT 1;
		`, p.Name, p.Developer).Scan(&pluginID)

		if err != nil {
			log.Printf("Failed to insert or find plugin %s: %v\n", p.Name, err)
			continue
		}

		// 2. Insert Release (if version exists, default to 1.0.0 if not)
		version := p.Version
		if version == "" {
			version = "1.0.0"
		}

		var releaseID string
		err = db.QueryRow(`
			WITH e AS (
				INSERT INTO plugin_releases (plugin_id, version, release_date)
				VALUES ($1, $2, NOW())
				ON CONFLICT (plugin_id, version) DO UPDATE SET version = EXCLUDED.version
				RETURNING id
			)
			SELECT id FROM e
			UNION ALL
			SELECT id FROM plugin_releases WHERE plugin_id = $1 AND version = $2
			LIMIT 1;
		`, pluginID, version).Scan(&releaseID)

		if err != nil {
			log.Printf("Failed to insert release for %s: %v\n", p.Name, err)
			continue
		}

		// 3. Insert Distribution if we have a download URL
		if p.DownloadURL != "" {
			hash, err := CalculateSHA256(p.DownloadURL)
			if err != nil {
				log.Printf("Failed to hash %s: %v\n", p.DownloadURL, err)
				hash = "unknown"
			}

			platform := p.Platform
			if platform == "" {
				platform = "windows" // default guess for testing
			}

			_, err = db.Exec(`
				INSERT INTO plugin_distributions (release_id, platform, architecture, download_url, sha256_hash, strategy, extraction_rules)
				VALUES ($1, $2, 'x86_64', $3, $4, 'extract_binaries', '{}')
				ON CONFLICT DO NOTHING`,
				releaseID, platform, p.DownloadURL, hash)

			if err != nil {
				log.Printf("Failed to insert distribution for %s: %v\n", p.Name, err)
			} else {
				log.Printf("Persisted distribution for %s\n", p.Name)
			}
		} else {
			log.Printf("Persisted plugin %s (No download URL)\n", p.Name)
		}
	}
}

func main() {
	log.Println("VST Monster Crawler starting...")
	initDB()
	if db != nil {
		defer db.Close()
	}

	c := InitCrawler()

	plugins := make([]ScrapedPlugin, 0)

	// We'll scrape Wikipedia's list of music software just to have some non-trivial data extraction that works
	c.OnHTML(".wikitable tbody tr", func(e *colly.HTMLElement) {
		name := e.ChildText("td:nth-child(1)")
		developer := e.ChildText("td:nth-child(2)")
		link := e.ChildAttr("td:nth-child(1) a", "href")

		if name != "" && developer != "" {
			// Prepend domain for relative links
			if strings.HasPrefix(link, "/") {
				link = "https://en.wikipedia.org" + link
			}

			plugins = append(plugins, ScrapedPlugin{
				Name:        name,
				Developer:   developer,
				DownloadURL: link,
				Version:     "1.0",
				Platform:    "windows",
			})
			log.Printf("Found Plugin: %s - %s", name, developer)
		}
	})

	c.OnScraped(func(r *colly.Response) {
		log.Printf("Finished scraping %s. Found %d plugins.\n", r.Request.URL, len(plugins))
		persistPlugins(plugins)
	})

	err := c.Visit("https://en.wikipedia.org/wiki/List_of_music_software")
	if err != nil {
		log.Printf("Failed to visit Wikipedia: %v", err)
	}

	c.Wait()
}
