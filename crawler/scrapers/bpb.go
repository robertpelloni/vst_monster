package scrapers

import (
	"log"
	"strings"
	"time"

	"github.com/gocolly/colly/v2"
)

// The models and functions need to be passed in or referenced properly.
// We will define interfaces or structs here to avoid circular imports,
// or let the caller pass a callback.

type PluginData struct {
	Name        string
	Developer   string
	DownloadURL string
	Category    string
}

func ScrapeBPB(onPluginFound func(PluginData), proxyFunc colly.ProxyFunc) {
	c := colly.NewCollector(
		colly.Async(true),
		colly.UserAgent("Mozilla/5.0 (Windows NT 10.0; Win64; x64) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/128.0.0.0 Safari/537.36"),
	)

	c.SetRequestTimeout(60 * time.Second)

	if proxyFunc != nil {
		c.SetProxyFunc(proxyFunc)
	}

	c.OnRequest(func(r *colly.Request) {
		log.Println("BPB Scraper Visiting", r.URL.String())
	})

	c.OnError(func(_ *colly.Response, err error) {
		log.Println("BPB Scraper Error:", err)
	})

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

				// Look for immediate download or outbound link in nearest anchor tags
				downloadURL := ""
				category := "synth"
				if strings.Contains(strings.ToLower(text), "reverb") || strings.Contains(strings.ToLower(text), "delay") || strings.Contains(strings.ToLower(text), "chorus") {
					category = "effect"
				} else if strings.Contains(strings.ToLower(text), "compressor") || strings.Contains(strings.ToLower(text), "limiter") {
					category = "dynamics"
				}

				if href := e.ChildAttr("a", "href"); href != "" {
					downloadURL = href
				}

				onPluginFound(PluginData{
					Name:        name,
					Developer:   developer,
					DownloadURL: downloadURL,
					Category:    category,
				})
			}
		}
	})

	err := c.Visit("https://bedroomproducersblog.com/free-vst-plugins/")
	if err != nil {
		log.Printf("BPB Scraper Visit Error: %v\n", err)
	}

	c.Wait()
}
