package scraper

import (
	"log"

	"github.com/gocolly/colly/v2"
	"github.com/robertpelloni/vst_monster/crawler/parser"
)

func BuildKVRScraper(c *colly.Collector, results chan<- parser.StandardPlugin) {
	// Example targeted scrape of the free instruments page
	c.OnHTML(".plugin_list_item", func(e *colly.HTMLElement) {
		name := e.ChildText(".plugin_name a")
		dev := e.ChildText(".developer_name a")
		formatDesc := e.ChildText(".plugin_formats")
		url := e.ChildAttr(".plugin_name a", "href")

		if name != "" {
			results <- parser.StandardPlugin{
				Name:        name,
				Developer:   dev,
				Version:     "1.0.0", // Extracted from deeper detail pages usually
				Formats:     parser.ExtractFormats(formatDesc),
				DownloadURL: "https://www.kvraudio.com" + url,
				License:     "free",
				Source:      "KVR Audio",
			}
			log.Printf("Extracted from KVR: %s by %s", name, dev)
		}
	})

	// Add error handling
	c.OnError(func(r *colly.Response, err error) {
		log.Printf("KVR Scraper Error: %v on %s", err, r.Request.URL)
	})
}
