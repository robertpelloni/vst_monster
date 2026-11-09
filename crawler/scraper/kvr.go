package scraper

import (
	"log"
	"strings"
	"time"

	"github.com/gocolly/colly/v2"
	"github.com/robertpelloni/vst_monster/crawler/parser"
)

func containsAny(s string, keywords ...string) bool {
	lower := strings.ToLower(s)
	for _, kw := range keywords {
		if strings.Contains(lower, kw) {
			return true
		}
	}
	return false
}

func BuildKVRScraper(c *colly.Collector, results chan<- parser.StandardPlugin) {
	c.Limit(&colly.LimitRule{
		DomainGlob:  "*kvraudio.*",
		Parallelism: 2,
		Delay:       2 * time.Second,
	})

	// Example targeted scrape of the free instruments page
	c.OnHTML(".plugin_list_item", func(e *colly.HTMLElement) {
		name := e.ChildText(".plugin_name a")
		dev := e.ChildText(".developer_name a")
		formatDesc := e.ChildText(".plugin_formats")
		url := e.ChildAttr(".plugin_name a", "href")

		if name != "" {
			metadata := make(map[string]interface{})
			metadata["raw_formats"] = formatDesc
			metadata["source"] = "kvr_audio"

			category := "effect"
			catLower := e.ChildText(".plugin_category")
			descLower := name + " " + formatDesc + " " + catLower
			if containsAny(descLower, "synth", "instrument", "piano", "organ", "drum", "sampler", "generator") {
				category = "synth"
			} else if containsAny(descLower, "compressor", "limiter", "gate", "maximizer", "dynamics") {
				category = "dynamics"
			} else if containsAny(descLower, "reverb", "delay", "chorus", "flanger", "phaser", "distortion", "filter", "eq", "equalizer") {
				category = "effect"
			}
			metadata["category"] = category

			results <- parser.StandardPlugin{
				Name:        name,
				Developer:   dev,
				Version:     "1.0.0",
				Formats:     parser.ExtractFormats(formatDesc),
				DownloadURL: "https://www.kvraudio.com" + url,
				License:     "free",
				Source:      "KVR Audio",
				Metadata:    metadata,
			}
			log.Printf("Extracted from KVR: %s by %s (Category: %s)", name, dev, category)
		}
	})

	// Add error handling
	c.OnError(func(r *colly.Response, err error) {
		log.Printf("KVR Scraper Error: %v on %s", err, r.Request.URL)
	})
}
