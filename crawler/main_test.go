package main

import (
	"testing"
)

func TestPluginCollection(t *testing.T) {
	pc := &PluginCollection{
		plugins: make([]ScrapedPlugin, 0),
	}

	p1 := ScrapedPlugin{Name: "Test 1"}
	p2 := ScrapedPlugin{Name: "Test 2"}

	pc.Add(p1)
	pc.Add(p2)

	plugins := pc.Get()
	if len(plugins) != 2 {
		t.Errorf("Expected 2 plugins, got %d", len(plugins))
	}

	if plugins[0].Name != "Test 1" || plugins[1].Name != "Test 2" {
		t.Errorf("Plugins not added in correct order or missing")
	}
}
