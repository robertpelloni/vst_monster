package scraper

import (
	"log"
	"strings"

	"github.com/gocolly/colly/v2"
	"github.com/robertpelloni/vst_monster/crawler/parser"
)

func BuildGithubScraper(c *colly.Collector, results chan<- parser.StandardPlugin) {
	// Example targeted scrape of github topics
	c.OnHTML("article.border", func(e *colly.HTMLElement) {
		repo := e.ChildText("h3 a.text-bold")
		desc := e.ChildText("p.color-text-secondary")
		url := "https://github.com" + e.ChildAttr("h3 a.text-bold", "href")

		if repo != "" {
			parts := strings.Split(repo, "/")
			dev := "Unknown"
			name := repo
			if len(parts) == 2 {
				dev = parts[0]
				name = parts[1]
			}

			results <- parser.StandardPlugin{
				Name:        name,
				Developer:   dev,
				Version:     "Latest",
				Formats:     parser.ExtractFormats(desc),
				DownloadURL: url,
				License:     "opensource",
				Source:      "GitHub",
			}
			log.Printf("Extracted from GitHub: %s/%s", dev, name)
		}
	})

	// Add error handling
	c.OnError(func(r *colly.Response, err error) {
		log.Printf("GitHub Scraper Error: %v on %s", err, r.Request.URL)
	})
}
