package scrapers

import (
	"log"
	"strings"
	"time"

	"github.com/gocolly/colly/v2"
)

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
				onPluginFound(PluginData{
					Name:      repo,
					Developer: owner,
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
