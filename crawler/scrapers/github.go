package scrapers

import (
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"os"
	"strings"
	"time"

	"github.com/gocolly/colly/v2"
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

func ScrapeGitHub(onPluginFound func(PluginData), proxyFunc colly.ProxyFunc) {
	c := colly.NewCollector(
		colly.Async(true),
		colly.UserAgent("VST-Monster-Bot/1.0 (+https://vstmonster.com)"),
		colly.AllowedDomains("github.com"),
	)

	c.SetRequestTimeout(60 * time.Second)

	if proxyFunc != nil {
		c.SetProxyFunc(proxyFunc)
	}

	// Limit concurrency
	c.Limit(&colly.LimitRule{
		DomainGlob:  "*github.*",
		Parallelism: 2,
		Delay:       3 * time.Second,
	})

	c.OnRequest(func(r *colly.Request) {
		log.Println("GitHub Scraper Visiting", r.URL.String())
	})

	c.OnError(func(_ *colly.Response, err error) {
		log.Println("GitHub Scraper Error:", err)
	})

	// Extract from typical repo layout on a topic page
	c.OnHTML("h3.f3", func(e *colly.HTMLElement) {
		// Inside the h3 there are generally two links: owner and repo name
		links := e.DOM.Find("a")
		if links.Length() >= 2 {
			owner := strings.TrimSpace(links.Eq(0).Text())
			repo := strings.TrimSpace(links.Eq(1).Text())

			// Clean up newlines and spaces that github wraps text in
			owner = strings.ReplaceAll(owner, "\n", "")
			owner = strings.TrimSpace(owner)
			repo = strings.ReplaceAll(repo, "\n", "")
			repo = strings.TrimSpace(repo)

			if owner != "" && repo != "" {
				downloadURL := fetchGitHubReleaseDownloadURL(owner, repo)
				onPluginFound(PluginData{
					Name:        repo,
					Developer:   owner,
					DownloadURL: downloadURL,
				})
			}
		}
	})

	// Follow pagination
	c.OnHTML("a.next_page", func(e *colly.HTMLElement) {
		nextPage := e.Attr("href")
		e.Request.Visit(e.Request.AbsoluteURL(nextPage))
	})

	// Target GitHub topics for VST plugins
	err := c.Visit("https://github.com/topics/vst-plugin")
	if err != nil {
		log.Printf("GitHub Scraper Visit Error: %v\n", err)
	}

	c.Wait()
}
