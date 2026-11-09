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
		Name               string `json:"name"`
		BrowserDownloadURL string `json:"browser_download_url"`
	} `json:"assets"`
}

func SelectBestReleaseAsset(assets []struct {
	Name               string `json:"name"`
	BrowserDownloadURL string `json:"browser_download_url"`
}) string {
	// 1. Prioritize dedicated audio bundle archives (.lv2, .vst3, .clap)
	for _, asset := range assets {
		lowerURL := strings.ToLower(asset.BrowserDownloadURL)
		if strings.Contains(lowerURL, "lv2") || strings.Contains(lowerURL, "vst3") || strings.Contains(lowerURL, "clap") {
			if strings.HasSuffix(lowerURL, ".zip") || strings.HasSuffix(lowerURL, ".tar.gz") {
				return asset.BrowserDownloadURL
			}
		}
	}

	// 2. Select standard multi-platform installer or archive formats
	for _, asset := range assets {
		lowerURL := strings.ToLower(asset.BrowserDownloadURL)
		if strings.HasSuffix(lowerURL, ".tar.gz") || strings.HasSuffix(lowerURL, ".zip") ||
			strings.HasSuffix(lowerURL, ".deb") || strings.HasSuffix(lowerURL, ".dmg") ||
			strings.HasSuffix(lowerURL, ".msi") || strings.HasSuffix(lowerURL, ".exe") ||
			strings.HasSuffix(lowerURL, ".appimage") {
			return asset.BrowserDownloadURL
		}
	}

	return ""
}

func fetchGitHubReleaseDownloadURL(owner, repo string) string {
	url := fmt.Sprintf("https://api.github.com/repos/%s/%s/releases/latest", owner, repo)

	client := &http.Client{Timeout: 10 * time.Second}
	req, err := http.NewRequest("GET", url, nil)
	if err != nil {
		return ""
	}

	req.Header.Set("User-Agent", "Mozilla/5.0 (Windows NT 10.0; Win64; x64) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/128.0.0.0 Safari/537.36")

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

	return SelectBestReleaseAsset(release.Assets)
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
			formats := parser.ExtractFormats(desc)

			// If URL or repo description implies LV2 format, ensure LV2 is registered
			lowerDesc := strings.ToLower(desc)
			lowerURL := strings.ToLower(downloadURL)
			isLV2 := strings.Contains(lowerDesc, "lv2") || strings.Contains(lowerURL, "lv2") || strings.Contains(lowerURL, ".lv2.")
			if isLV2 {
				hasLV2 := false
				for _, f := range formats {
					if strings.EqualFold(f, "LV2") {
						hasLV2 = true
						break
					}
				}
				if !hasLV2 {
					formats = append(formats, "LV2")
				}
			}

			category := "effect"
			if strings.Contains(lowerDesc, "synth") || strings.Contains(lowerDesc, "instrument") || strings.Contains(lowerDesc, "generator") || strings.Contains(lowerDesc, "modular") {
				category = "synth"
			} else if strings.Contains(lowerDesc, "compressor") || strings.Contains(lowerDesc, "limiter") || strings.Contains(lowerDesc, "gate") || strings.Contains(lowerDesc, "tape") {
				category = "dynamics"
			}

			meta := map[string]interface{}{
				"source":      "github",
				"category":    category,
				"description": desc,
			}
			if isLV2 {
				meta["has_lv2_manifest"] = true
				meta["manifest_format"] = "Turtle (.ttl)"
			}

			results <- parser.StandardPlugin{
				Name:        name,
				Developer:   dev,
				Version:     "Latest",
				Formats:     formats,
				DownloadURL: downloadURL,
				License:     "opensource",
				Source:      "GitHub",
				Metadata:    meta,
			}
			log.Printf("Extracted from GitHub: %s/%s", dev, name)
		}
	})

	// Add error handling
	c.OnError(func(r *colly.Response, err error) {
		log.Printf("GitHub Scraper Error: %v on %s", err, r.Request.URL)
	})
}
