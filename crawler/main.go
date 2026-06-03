package main

import (
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"io"
	"log"
	"net/http"
	"strings"
	"time"

	"github.com/gocolly/colly/v2"
)

type ScrapedPlugin struct {
	Name        string   `json:"name"`
	Developer   string   `json:"developer"`
	DownloadURL string   `json:"download_url"`
	Version     string   `json:"version"`
	Platform    string   `json:"platform"` // windows, macos, linux
}

// InitCrawler initializes the colly collector with standard settings.
func InitCrawler() *colly.Collector {
	c := colly.NewCollector(
		colly.Async(true),
		colly.UserAgent("VST-Monster-Bot/1.0 (+https://vstmonster.com)"),
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

// CalculateSHA256 downloads the file from the given URL and calculates its SHA256 hash.
func CalculateSHA256(url string) (string, error) {
	resp, err := http.Get(url)
	if err != nil {
		return "", err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return "", fmt.Errorf("bad status: %s", resp.Status)
	}

	hasher := sha256.New()
	if _, err := io.Copy(hasher, resp.Body); err != nil {
		return "", err
	}
	return hex.EncodeToString(hasher.Sum(nil)), nil
}

func main() {
	log.Println("VST Monster Crawler starting...")

	c := InitCrawler()

	plugins := make([]ScrapedPlugin, 0)

	c.OnHTML("h1, h2, h3", func(e *colly.HTMLElement) {
		text := e.Text
		if strings.Contains(text, " by ") {
			parts := strings.Split(text, " by ")
			if len(parts) >= 2 {
				nameAndNum := strings.TrimSpace(parts[0])
				developer := strings.TrimSpace(parts[1])

				name := nameAndNum
				dotIndex := strings.Index(nameAndNum, ". ")
				if dotIndex != -1 && dotIndex < 5 {
					name = strings.TrimSpace(nameAndNum[dotIndex+2:])
				}

				plugin := ScrapedPlugin{
					Name:      name,
					Developer: developer,
				}

				plugins = append(plugins, plugin)
				log.Printf("Found plugin: %s by %s\n", name, developer)
			}
		}
	})

	c.OnScraped(func(r *colly.Response) {
		log.Printf("Finished scraping %s. Found %d plugins.\n", r.Request.URL, len(plugins))
		if len(plugins) > 0 {
			jsonData, _ := json.MarshalIndent(plugins, "", "  ")
			fmt.Println(string(jsonData))
		}
	})

	err := c.Visit("https://bedroomproducersblog.com/free-vst-plugins/")
	if err != nil {
		log.Fatal(err)
	}

	c.Wait()
}
