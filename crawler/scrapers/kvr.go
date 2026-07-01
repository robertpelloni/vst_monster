package scrapers

import (
	"github.com/PuerkitoBio/goquery"
	"log"
	"strings"
	"time"

	"github.com/gocolly/colly/v2"
)

func ScrapeKVR(onPluginFound func(PluginData), proxyFunc colly.ProxyFunc) {
	c := colly.NewCollector(
		colly.Async(true),
		colly.UserAgent("VST-Monster-Bot/1.0 (+https://vstmonster.com)"),
		colly.AllowedDomains("www.kvraudio.com", "kvraudio.com"),
	)

	c.SetRequestTimeout(60 * time.Second)

	if proxyFunc != nil {
		c.SetProxyFunc(proxyFunc)
	}

	// Limit concurrency to be polite to the site
	c.Limit(&colly.LimitRule{
		DomainGlob:  "*kvraudio.*",
		Parallelism: 2,
		Delay:       2 * time.Second,
	})

	c.OnRequest(func(r *colly.Request) {
		log.Println("KVR Scraper Visiting", r.URL.String())
	})

	c.OnError(func(_ *colly.Response, err error) {
		log.Println("KVR Scraper Error:", err)
	})

	// The list of items typically contains a link with class or direct structure
	// Look at typical KVR listing: <div class="product-info">...
	// We'll scrape links to product pages first, then scrape the product pages
	c.OnHTML("a[href^='/product/']", func(e *colly.HTMLElement) {
		link := e.Attr("href")
		// Visit the product page
		e.Request.Visit(e.Request.AbsoluteURL(link))
	})

	// Scrape product details on product page
	c.OnHTML("h1.product-name", func(e *colly.HTMLElement) {
		name := strings.TrimSpace(e.Text)

		// Find developer
		developer := ""
		e.DOM.Parent().Find(".developer-name").Each(func(i int, s *goquery.Selection) {
			developer = strings.TrimSpace(s.Text())
		})

		// Fallback developer extraction
		if developer == "" {
			e.DOM.Closest(".product-header").Find("a[href^='/developer/']").Each(func(i int, s *goquery.Selection) {
				developer = strings.TrimSpace(s.Text())
			})
		}

		if name != "" && developer != "" {
			onPluginFound(PluginData{
				Name:      name,
				Developer: developer,
			})
		}
	})

	// Start at the newest free plugins page
	err := c.Visit("https://www.kvraudio.com/plugins/free/newest")
	if err != nil {
		log.Printf("KVR Scraper Visit Error: %v\n", err)
	}

	c.Wait()
}
