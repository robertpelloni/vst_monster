package scraper

import (
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"os"
	"strings"
	"time"

	"github.com/gocolly/colly/v2"
	"github.com/robertpelloni/vst_monster/crawler/parser"
)

// GitHubRelease represents the subset of fields we need from the API response
type GitHubRelease struct {
	Assets []struct {
		BrowserDownloadURL string `json:"browser_download_url"`
	} `json:"assets"`
}

func fetchGitHubReleaseDownloadURL(owner, repo string) string {
	url := fmt.Sprintf("https://api.github.com/repos/%s/%s/releases/latest", owner, repo)

	client := &http.Client{Timeout: 10 * time.Second}
	req, err := http.NewRequest("GET", url, nil)
	if err != nil {
		return ""
	}

	req.Header.Set("User-Agent", "VST-Monster-Bot/1.0 (+https://vstmonster.com)")

	if token := os.Getenv("GITHUB_TOKEN"); token != "" {
		req.Header.Set("Authorization", "token "+token)
	}

	resp, err := client.Do(req)
	if err != nil || resp.StatusCode != http.StatusOK {
		if resp != nil {
			resp.Body.Close()
		}
		return ""
	}
	defer resp.Body.Close()

	var release GitHubRelease
	if err := json.NewDecoder(resp.Body).Decode(&release); err != nil {
		return ""
	}

	for _, asset := range release.Assets {
		lowerURL := strings.ToLower(asset.BrowserDownloadURL)
		if strings.HasSuffix(lowerURL, ".zip") || strings.HasSuffix(lowerURL, ".tar.gz") {
			return asset.BrowserDownloadURL
		}
	}

	return ""
}

func BuildGithubScraper(c *colly.Collector, results chan<- parser.StandardPlugin) {
	c.Limit(&colly.LimitRule{
		DomainGlob:  "*github.*",
		Parallelism: 2,
		Delay:       3 * time.Second,
	})

	// Example targeted scrape of github topics
	c.OnHTML("article.border", func(e *colly.HTMLElement) {
		repo := e.ChildText("h3 a.text-bold")
		desc := e.ChildText("p.color-text-secondary")

		if repo != "" {
			parts := strings.Split(repo, "/")
			dev := "Unknown"
			name := repo
			if len(parts) == 2 {
				dev = parts[0]
				name = parts[1]
			}

			downloadURL := fetchGitHubReleaseDownloadURL(dev, name)

			results <- parser.StandardPlugin{
				Name:        name,
				Developer:   dev,
				Version:     "Latest",
				Formats:     parser.ExtractFormats(desc),
				DownloadURL: downloadURL,
				License:     "opensource",
				Source:      "GitHub",
				Metadata:    map[string]interface{}{"source": "github"},
			}
			log.Printf("Extracted from GitHub: %s/%s", dev, name)
		}
	})

	// Add error handling
	c.OnError(func(r *colly.Response, err error) {
		log.Printf("GitHub Scraper Error: %v on %s", err, r.Request.URL)
	})
}
