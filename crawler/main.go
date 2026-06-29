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
	"sync"
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
	client := &http.Client{
		Timeout: 30 * time.Second, // Configure strict timeout to prevent hung threads
	}
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

// Thread-safe plugin collection
type PluginCollection struct {
	mu      sync.Mutex
	plugins []ScrapedPlugin
}

func (pc *PluginCollection) Add(plugin ScrapedPlugin) {
	pc.mu.Lock()
	defer pc.mu.Unlock()
	pc.plugins = append(pc.plugins, plugin)
}

func (pc *PluginCollection) Get() []ScrapedPlugin {
	pc.mu.Lock()
	defer pc.mu.Unlock()
	return pc.plugins
}

// ScrapeGithubAwesomeList scrapes an awesome list of VSTs on GitHub.
func ScrapeGithubAwesomeList(c *colly.Collector, pc *PluginCollection) {
	c.OnHTML("article.markdown-body ul li", func(e *colly.HTMLElement) {
		text := e.Text
		link := e.ChildAttr("a", "href")

		if strings.Contains(strings.ToLower(text), "vst") || strings.Contains(strings.ToLower(text), "synth") {
			parts := strings.Split(text, "-")
			name := "Unknown"
			desc := "Unknown"

			if len(parts) >= 2 {
				name = strings.TrimSpace(parts[0])
				desc = strings.TrimSpace(parts[1])
			} else {
				name = text
			}

			if link != "" {
				pc.Add(ScrapedPlugin{
					Name:        name,
					Developer:   "Open Source Community", // Generic developer for awesome list
					DownloadURL: link,
					Version:     "1.0",
					Platform:    "windows",
				})
				log.Printf("Found Plugin: %s - %s", name, desc)
			}
		}
	})

	err := c.Visit("https://github.com/lucianoiam/awesome-vst")
	if err != nil {
		log.Printf("Failed to visit GitHub Awesome List: %v", err)
	}
}

// ScrapeKVR scrapes the KVR Audio news feed for plugin releases.
func ScrapeKVR(c *colly.Collector, pc *PluginCollection) {
	c.OnXML("//item", func(e *colly.XMLElement) {
		title := e.ChildText("title")
		link := e.ChildText("link")
		// Very naive parsing just for demonstration purposes
		if strings.Contains(strings.ToLower(title), "vst") || strings.Contains(strings.ToLower(title), "plugin") {
			pc.Add(ScrapedPlugin{
				Name:        title,
				Developer:   "KVR Audio",
				DownloadURL: link,
				Version:     "1.0",
				Platform:    "windows",
			})
			log.Printf("Found KVR Plugin: %s", title)
		}
	})

	err := c.Visit("https://www.kvraudio.com/news.xml")
	if err != nil {
		log.Printf("Failed to visit KVR: %v", err)
	}
}

// ScrapePluginBoutique scrapes the free plugins page on Plugin Boutique.
func ScrapePluginBoutique(c *colly.Collector, pc *PluginCollection) {
	c.OnHTML(".product-tile", func(e *colly.HTMLElement) {
		name := e.ChildText(".product-title")
		developer := e.ChildText(".product-developer")
		link := e.ChildAttr("a.product-image", "href")

		if name != "" && link != "" {
			if strings.HasPrefix(link, "/") {
				link = "https://www.pluginboutique.com" + link
			}
			pc.Add(ScrapedPlugin{
				Name:        name,
				Developer:   developer,
				DownloadURL: link,
				Version:     "1.0",
				Platform:    "windows",
			})
			log.Printf("Found Plugin Boutique Plugin: %s by %s", name, developer)
		}
	})

	err := c.Visit("https://www.pluginboutique.com/categories/free")
	if err != nil {
		log.Printf("Failed to visit Plugin Boutique: %v", err)
	}
}


func main() {
	log.Println("VST Monster Crawler starting...")
	initDB()
	if db != nil {
		defer db.Close()
	}

	c := InitCrawler()
	pc := &PluginCollection{
		plugins: make([]ScrapedPlugin, 0),
	}

	// Since websites like KVR Audio block simple scrapers with 403 Forbidden,
	// we structure the code to support multiple robust targets.
	// To actually extract without being blocked, an advanced proxy rotation and headless browser
	// would eventually be required for sites like KVR.
	// We'll leave the function stubs active for GitHub and KVR as instructed, and fallback
	// to the functional wikipedia extraction if needed.

	// Target 1: GitHub Awesome VST List
	// Target 2: KVR Audio
	// Target 3: Plugin Boutique

	// Create dedicated clones for different strategies to avoid conflicting callbacks
	c_github := c.Clone()
	ScrapeGithubAwesomeList(c_github, pc)

	c_kvr := c.Clone()
	ScrapeKVR(c_kvr, pc)

	c_pb := c.Clone()
	ScrapePluginBoutique(c_pb, pc)


	c_github.Wait()
	c_kvr.Wait()
	c_pb.Wait()

	plugins := pc.Get()
	log.Printf("Finished all scraping targets. Found %d total plugins.\n", len(plugins))
	persistPlugins(plugins)
}
