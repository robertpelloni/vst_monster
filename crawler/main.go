package main

import (
	"log"
	"strings"

	"github.com/gocolly/colly/v2"
	"github.com/gocolly/colly/v2/proxy"

	"github.com/robertpelloni/vst_monster/crawler/config"
	"github.com/robertpelloni/vst_monster/crawler/parser"
	"github.com/robertpelloni/vst_monster/crawler/queue"
	"github.com/robertpelloni/vst_monster/crawler/scraper"
)

func main() {
	log.Println("Starting VST Monster Crawler Pipeline...")
	cfg := config.LoadConfig()

	db := InitDB()
	if db != nil {
		defer db.Close()
	}

	pipeline := queue.NewPipeline(db)

	pipeline.SetCallback(func(plugin parser.StandardPlugin) {
		log.Printf("Found plugin: %s by %s\n", plugin.Name, plugin.Developer)

		hash := ""
		if plugin.DownloadURL != "" {
			// Basic check to avoid hashing HTML pages
			isLikelyArchive := false
			lowerURL := strings.ToLower(plugin.DownloadURL)
			if len(lowerURL) > 4 {
				if lowerURL[len(lowerURL)-4:] == ".zip" || lowerURL[len(lowerURL)-4:] == ".exe" || lowerURL[len(lowerURL)-4:] == ".dmg" || lowerURL[len(lowerURL)-4:] == ".pkg" || lowerURL[len(lowerURL)-4:] == ".msi" || (len(lowerURL) > 7 && lowerURL[len(lowerURL)-7:] == ".tar.gz") {
					isLikelyArchive = true
				}
			}

			if isLikelyArchive {
				computed, err := CalculateSHA256(plugin.DownloadURL)
				if err != nil {
					log.Printf("Failed to compute hash for %s: %v", plugin.DownloadURL, err)
				} else {
					hash = computed
				}
			}
		}

		if db != nil {
			err := UpsertPlugin(db, plugin, hash)
			if err != nil {
				log.Printf("Error upserting %s: %v\n", plugin.Name, err)
			}
		}
	})

	pipeline.StartConsumers(10) // 10 concurrent database/hashing workers

	c := colly.NewCollector(
		colly.Async(true),
		colly.UserAgent(cfg.UserAgent),
	)

	c.SetRequestTimeout(cfg.Timeout)
	c.Limit(&colly.LimitRule{
		DomainGlob:  "*",
		Parallelism: cfg.Concurrency,
	})

	if cfg.ProxyURL != "" {
		rp, err := proxy.RoundRobinProxySwitcher(cfg.ProxyURL)
		if err != nil {
			log.Fatalf("Failed to configure proxy: %v", err)
		}
		c.SetProxyFunc(rp)
		log.Println("Proxy rotation configured.")
	}

	// Create targeted scrapers
	kvrCollector := c.Clone()
	scraper.BuildKVRScraper(kvrCollector, pipeline.Results)

	githubCollector := c.Clone()
	scraper.BuildGithubScraper(githubCollector, pipeline.Results)

	// Start crawls
	log.Println("Crawling KVR Audio...")
	kvrCollector.Visit("https://www.kvraudio.com/plugins/free/newest")

	log.Println("Crawling GitHub VST topics...")
	githubCollector.Visit("https://github.com/topics/vst-plugin")

	// Wait for completion
	kvrCollector.Wait()
	githubCollector.Wait()

	pipeline.Close()
	log.Println("Crawling session complete.")
}
