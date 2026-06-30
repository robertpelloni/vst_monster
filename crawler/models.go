package main

import (
	"encoding/json"
	"sync"
)

type ScrapedPlugin struct {
	Name        string          `json:"name"`
	Developer   string          `json:"developer"`
	DownloadURL string          `json:"download_url"`
	Version     string          `json:"version"`
	Platform    string          `json:"platform"` // windows, macos, linux
	Metadata    json.RawMessage `json:"metadata"` // For JSONB storage in Registry API
}

type PluginCollection struct {
	sync.Mutex
	Plugins []ScrapedPlugin
}

func NewPluginCollection() *PluginCollection {
	return &PluginCollection{
		Plugins: make([]ScrapedPlugin, 0),
	}
}

func (pc *PluginCollection) Add(plugin ScrapedPlugin) {
	pc.Lock()
	defer pc.Unlock()
	pc.Plugins = append(pc.Plugins, plugin)
}

func (pc *PluginCollection) GetAll() []ScrapedPlugin {
	pc.Lock()
	defer pc.Unlock()

	// Return a copy to prevent external mutation while holding lock
	copied := make([]ScrapedPlugin, len(pc.Plugins))
	copy(copied, pc.Plugins)
	return copied
}
