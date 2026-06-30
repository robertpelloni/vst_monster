package main

import (
	"sync"
	"testing"
)

func TestPluginCollection(t *testing.T) {
	pc := NewPluginCollection()

	// Test Add and GetAll
	plugin := ScrapedPlugin{
		Name: "Test Plugin",
		Developer: "Test Dev",
	}
	pc.Add(plugin)

	plugins := pc.GetAll()
	if len(plugins) != 1 {
		t.Errorf("Expected 1 plugin, got %d", len(plugins))
	}

	if plugins[0].Name != "Test Plugin" {
		t.Errorf("Expected plugin name 'Test Plugin', got %s", plugins[0].Name)
	}

	// Test concurrent access
	var wg sync.WaitGroup
	for i := 0; i < 100; i++ {
		wg.Add(1)
		go func(idx int) {
			defer wg.Done()
			pc.Add(ScrapedPlugin{Name: "Plugin"})
		}(i)
	}
	wg.Wait()

	allPlugins := pc.GetAll()
	if len(allPlugins) != 101 { // 100 concurrent + 1 original
		t.Errorf("Expected 101 plugins, got %d", len(allPlugins))
	}
}