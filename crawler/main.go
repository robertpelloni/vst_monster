package main

import (
	"fmt"
	"log"
	"strings"
	"sync"

	"github.com/gocolly/colly/v2"
	"github.com/gocolly/colly/v2/proxy"

	"github.com/robertpelloni/vst_monster/crawler/config"
	"github.com/robertpelloni/vst_monster/crawler/parser"
	"github.com/robertpelloni/vst_monster/crawler/queue"
	"github.com/robertpelloni/vst_monster/crawler/scraper"
	"github.com/robertpelloni/vst_monster/crawler/scrapers"
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

	var proxyFunc colly.ProxyFunc
	if cfg.ProxyURL != "" {
		rp, err := proxy.RoundRobinProxySwitcher(cfg.ProxyURL)
		if err != nil {
			log.Fatalf("Failed to configure proxy: %v", err)
		}
		proxyFunc = rp
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

	log.Println("Crawling GitHub LV2 topics...")
	githubCollector.Visit("https://github.com/topics/lv2")
	githubCollector.Visit("https://github.com/topics/lv2-plugin")

	var wg sync.WaitGroup

	// Run Bedroom Producers Blog scraper with WaitGroup tracking
	log.Println("Crawling Bedroom Producers Blog...")
	wg.Add(1)
	go func() {
		defer wg.Done()
		defer func() {
			if r := recover(); r != nil {
				log.Printf("Recovered from panic in BPB scraper: %v\n", r)
			}
		}()
		scrapers.ScrapeBPB(func(data scrapers.PluginData) {
			pipeline.Results <- parser.StandardPlugin{
				Name:        data.Name,
				Developer:   data.Developer,
				Version:     "1.0.0",
				DownloadURL: data.DownloadURL,
				License:     "free",
				Source:      "Bedroom Producers Blog",
				Metadata:    map[string]interface{}{"source": "bpb", "category": data.Category, "description": fmt.Sprintf("%s by %s via Bedroom Producers Blog", data.Name, data.Developer)},
			}
		}, proxyFunc)
	}()

	// Wait for all collectors and goroutines to finish
	kvrCollector.Wait()
	githubCollector.Wait()
	wg.Wait()

	// Seed Ingestion: Ensure popular open-source and free plugins exist in database
	seedStaplePlugins(pipeline)

	pipeline.Close()
	log.Println("Crawling session complete.")
}

func seedStaplePlugins(pipeline *queue.Pipeline) {
	staples := []parser.StandardPlugin{
		{
			Name:        "Vital Audio Synthesizer",
			Developer:   "Matt Tytel",
			Version:     "1.5.5",
			Formats:     []string{"VST3", "CLAP", "AU"},
			DownloadURL: "https://vital-audio.s3.dualstack.us-east-1.amazonaws.com/VitalInstaller.zip",
			License:     "opensource",
			Source:      "Seed Manifest",
			Metadata:    map[string]interface{}{"description": "Spectral warping wavetable synthesizer with modern modulation workflow.", "source": "official"},
		},
		{
			Name:        "Surge XT",
			Developer:   "Surge Synth Team",
			Version:     "1.3.4",
			Formats:     []string{"VST3", "CLAP", "AU"},
			DownloadURL: "https://github.com/surge-synthesizer/releases-xt/releases/download/1.3.4/surge-xt-linux-x64-1.3.4.deb",
			License:     "opensource",
			Source:      "Seed Manifest",
			Metadata:    map[string]interface{}{"description": "Free and open-source hybrid synthesizer with deep routing and MPE support.", "source": "github"},
		},
		{
			Name:        "Dexed FM Plugin Synth",
			Developer:   "Digital Suburban",
			Version:     "0.9.7",
			Formats:     []string{"VST3", "AU", "CLAP"},
			DownloadURL: "https://github.com/asb2m10/dexed/releases/download/v0.9.7/Dexed_Linux.zip",
			License:     "opensource",
			Source:      "Seed Manifest",
			Metadata:    map[string]interface{}{"description": "Multi-platform, multi-format plugin synth modeled on Yamaha DX7 sound engine.", "source": "github"},
		},
		{
			Name:        "Odin 2 Synthesizer",
			Developer:   "The Wave Warden",
			Version:     "2.3.4",
			Formats:     []string{"VST3", "CLAP", "AU"},
			DownloadURL: "https://thewavewarden.com/assets/Odin2Linux.zip",
			License:     "free",
			Source:      "Seed Manifest",
			Metadata:    map[string]interface{}{"description": "24-voice polyphonic powerhouse with analog-modeled filters and distortion modules.", "source": "official", "category": "synth"},
		},
		{
			Name:        "Cardinal Modular Synthesizer",
			Developer:   "DISTRHO",
			Version:     "24.04",
			Formats:     []string{"VST3", "CLAP", "LV2"},
			DownloadURL: "https://github.com/DISTRHO/Cardinal/releases/download/24.04/Cardinal-linux-x86_64-24.04.tar.gz",
			License:     "opensource",
			Source:      "Seed Manifest",
			Metadata:    map[string]interface{}{"description": "Virtual modular synthesizer plugin based on VCV Rack with multi-format LV2 and CLAP support.", "source": "github", "category": "synth"},
		},
		{
			Name:        "Chow Tape Model",
			Developer:   "Chowdhury DSP",
			Version:     "2.11.4",
			Formats:     []string{"VST3", "CLAP", "AU", "LV2"},
			DownloadURL: "https://github.com/jatinchowdhury18/AnalogTapeModel/releases/download/v2.11.4/ChowTapeModel-Linux-x64-2.11.4.deb",
			License:     "opensource",
			Source:      "Seed Manifest",
			Metadata:    map[string]interface{}{"description": "Physical modeling analog reel-to-reel tape machine saturation and hysteresis effect.", "source": "github", "category": "effect"},
		},
		{
			Name:        "Dragonfly Early Reflections Reverb",
			Developer:   "Michael Willis",
			Version:     "3.2.10",
			Formats:     []string{"LV2", "VST3"},
			DownloadURL: "https://github.com/michaelwillis/dragonfly-reverb/releases/download/3.2.10/dragonfly-reverb-linux-x86_64-3.2.10.tar.gz",
			License:     "opensource",
			Source:      "Seed Manifest",
			Metadata:    map[string]interface{}{"description": "Concert hall and studio room acoustic algorithmic reverb suite in native LV2 bundle format.", "source": "github", "category": "effect"},
		},
		{
			Name:        "Helm Polyphonic Synthesizer",
			Developer:   "Matt Tytel",
			Version:     "0.9.0",
			Formats:     []string{"LV2", "VST3", "AU"},
			DownloadURL: "https://github.com/mtytel/helm/releases/download/v0.9.0/helm-0.9.0-x86_64.tar.gz",
			License:     "opensource",
			Source:      "Seed Manifest",
			Metadata:    map[string]interface{}{"description": "Free polyphonic synthesizer running as standalone or LV2/VST3 plugin with visual modulation.", "source": "github", "category": "synth"},
		},
	}

	for _, p := range staples {
		pipeline.Results <- p
	}
}
