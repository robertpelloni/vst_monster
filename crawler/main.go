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
}
