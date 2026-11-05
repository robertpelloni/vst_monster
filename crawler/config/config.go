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
	ua := os.Getenv("CRAWLER_USER_AGENT")
	if ua == "" {
		// Use modern Chrome desktop user-agent to avoid immediate 403 Forbidden blocking
		ua = "Mozilla/5.0 (Windows NT 10.0; Win64; x64) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/128.0.0.0 Safari/537.36"
	}

	return CrawlerConfig{
		UserAgent:   ua,
		Timeout:     30 * time.Second,
		ProxyURL:    os.Getenv("CRAWLER_PROXY_URL"), // Fallback to none if empty
		Concurrency: 5,
	}
}
