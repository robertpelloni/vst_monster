package scraper

import (
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/gocolly/colly/v2"
	"github.com/robertpelloni/vst_monster/crawler/parser"
)

func TestBuildKVRScraper(t *testing.T) {
	c := colly.NewCollector()
	ch := make(chan parser.StandardPlugin, 10)
	BuildKVRScraper(c, ch)
	if c == nil {
		t.Fatal("Expected non-nil collector")
	}
}

func TestBuildKVRScraper_Extraction(t *testing.T) {
	htmlPayload := `
	<!DOCTYPE html>
	<html>
	<body>
		<div class="plugin_list_item">
			<div class="plugin_name"><a href="/product/super-synth-by-audio-lab">Super Synth</a></div>
			<div class="developer_name"><a href="/developer/audio-lab">Audio Lab</a></div>
			<div class="plugin_formats">VST3, CLAP, AU 64-bit</div>
			<div class="plugin_category">Virtual Synthesizer Instrument</div>
		</div>
		<div class="plugin_list_item">
			<div class="plugin_name"><a href="/product/tape-compressor">Tape Compressor</a></div>
			<div class="developer_name"><a href="/developer/analog-dsp">Analog DSP</a></div>
			<div class="plugin_formats">VST3, AU</div>
			<div class="plugin_category">Dynamics Bus Compressor</div>
		</div>
	</body>
	</html>
	`

	ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "text/html")
		_, _ = w.Write([]byte(htmlPayload))
	}))
	defer ts.Close()

	c := colly.NewCollector()
	results := make(chan parser.StandardPlugin, 5)

	BuildKVRScraper(c, results)

	err := c.Visit(ts.URL)
	if err != nil {
		t.Fatalf("Failed to visit mock server: %v", err)
	}

	var extracted []parser.StandardPlugin
	timeout := time.After(2 * time.Second)

loop:
	for len(extracted) < 2 {
		select {
		case p := <-results:
			extracted = append(extracted, p)
		case <-timeout:
			break loop
		}
	}

	if len(extracted) != 2 {
		t.Fatalf("Expected 2 extracted plugins, got %d", len(extracted))
	}

	synth := extracted[0]
	if synth.Name != "Super Synth" {
		t.Errorf("Expected 'Super Synth', got '%s'", synth.Name)
	}
	if synth.Developer != "Audio Lab" {
		t.Errorf("Expected 'Audio Lab', got '%s'", synth.Developer)
	}
	if synth.License != "free" {
		t.Errorf("Expected license 'free', got '%s'", synth.License)
	}
	if synth.Metadata["category"] != "synth" {
		t.Errorf("Expected category 'synth', got '%v'", synth.Metadata["category"])
	}

	comp := extracted[1]
	if comp.Name != "Tape Compressor" {
		t.Errorf("Expected 'Tape Compressor', got '%s'", comp.Name)
	}
	if comp.Metadata["category"] != "dynamics" {
		t.Errorf("Expected category 'dynamics', got '%v'", comp.Metadata["category"])
	}
}
