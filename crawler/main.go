package main

import (
	"log"

	"github.com/gocolly/colly/v2"
	"github.com/gocolly/colly/v2/proxy"

	"github.com/robertpelloni/vst_monster/crawler/config"
	"github.com/robertpelloni/vst_monster/crawler/queue"
	"github.com/robertpelloni/vst_monster/crawler/scraper"
)

func main() {
	log.Println("Starting VST Monster Crawler Pipeline...")
	cfg := config.LoadConfig()

	pipeline := queue.NewPipeline()
	pipeline.StartConsumers()

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
	"encoding/json"
	"fmt"
	"log"
	"sync"

	"github.com/robertpelloni/vst_monster/crawler/scrapers"
)

func main() {
	log.Println("VST Monster Crawler starting...")

	db := InitDB()
	if db != nil {
		defer db.Close()
	}

	plugins := NewPluginCollection()

	// Create a callback that scrapers use to send data back
	onPluginFound := func(data scrapers.PluginData) {
		plugin := ScrapedPlugin{
			Name:      data.Name,
			Developer: data.Developer,
		}

		plugins.Add(plugin)
		log.Printf("Found plugin: %s by %s\n", plugin.Name, plugin.Developer)

		// Attempt upsert immediately if db is available
		if db != nil {
			err := UpsertPlugin(db, plugin, "")
			if err != nil {
				log.Printf("Error upserting %s: %v\n", plugin.Name, err)
			}
		}
	}

	proxyFunc := GetProxySwitcher()

	// Run scrapers concurrently
	var wg sync.WaitGroup

	log.Println("Starting BedroomProducersBlog Scraper...")
	wg.Add(1)
	go func() {
		defer wg.Done()
		scrapers.ScrapeBPB(onPluginFound, proxyFunc)
	}()

	log.Println("Starting KVR Scraper...")
	wg.Add(1)
	go func() {
		defer wg.Done()
		scrapers.ScrapeKVR(onPluginFound, proxyFunc)
	}()

	log.Println("Starting GitHub Scraper...")
	wg.Add(1)
	go func() {
		defer wg.Done()
		scrapers.ScrapeGitHub(onPluginFound, proxyFunc)
	}()

	// Wait for all scrapers to finish
	wg.Wait()

	finalPlugins := plugins.GetAll()
	log.Printf("Finished all scraping tasks. Found %d plugins total.\n", len(finalPlugins))

	if len(finalPlugins) > 0 {
		jsonData, _ := json.MarshalIndent(finalPlugins, "", "  ")
		fmt.Println(string(jsonData))
	}
}
