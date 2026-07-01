package main

import (
	"encoding/json"
	"fmt"
	"log"
	"strings"
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
			Name:        data.Name,
			Developer:   data.Developer,
			DownloadURL: data.DownloadURL,
		}

		plugins.Add(plugin)
		log.Printf("Found plugin: %s by %s\n", plugin.Name, plugin.Developer)

		// Compute hash if there's a download URL that points directly to a file
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

		// Attempt upsert immediately if db is available
		if db != nil {
			err := UpsertPlugin(db, plugin, hash)
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
