package main

import (
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
