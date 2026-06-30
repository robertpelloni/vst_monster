package main

import (
	"log"
	"os"
	"strings"

	"github.com/gocolly/colly/v2"
	"github.com/gocolly/colly/v2/proxy"
)

// GetProxySwitcher reads a comma-separated PROXY_LIST from the environment
// and returns a proxy.RoundRobinSwitcher which is compatible with colly's SetProxyFunc.
// If PROXY_LIST is empty, it returns nil.
func GetProxySwitcher() colly.ProxyFunc {
	proxyStr := os.Getenv("PROXY_LIST")
	if proxyStr == "" {
		return nil
	}

	proxies := strings.Split(proxyStr, ",")
	var validProxies []string

	for _, p := range proxies {
		trimmed := strings.TrimSpace(p)
		if trimmed != "" {
			validProxies = append(validProxies, trimmed)
		}
	}

	if len(validProxies) == 0 {
		return nil
	}

	proxySwitcher, err := proxy.RoundRobinProxySwitcher(validProxies...)
	if err != nil {
		log.Printf("Error creating proxy switcher: %v\n", err)
		return nil
	}

	log.Printf("Configured %d proxies for rotation.\n", len(validProxies))
	return proxySwitcher
}
