package config

import (
	"os"
	"time"
)

type CrawlerConfig struct {
	UserAgent   string
	Timeout     time.Duration
	ProxyURL    string
	Concurrency int
}

func LoadConfig() CrawlerConfig {
	return CrawlerConfig{
		UserAgent:   "VSTMonsterBot/1.0",
		Timeout:     30 * time.Second,
		ProxyURL:    os.Getenv("CRAWLER_PROXY_URL"), // Fallback to none if empty
		Concurrency: 5,
	}
}
