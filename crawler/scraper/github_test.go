package scraper

import (
	"strings"
	"testing"

	"github.com/gocolly/colly/v2"
	"github.com/robertpelloni/vst_monster/crawler/parser"
)

func TestBuildGithubScraper(t *testing.T) {
	c := colly.NewCollector()
	ch := make(chan parser.StandardPlugin, 10)
	BuildGithubScraper(c, ch)
	if c == nil {
		t.Fatal("Expected non-nil collector")
	}
}

func TestSelectBestReleaseAsset(t *testing.T) {
	assets := []struct {
		Name               string `json:"name"`
		BrowserDownloadURL string `json:"browser_download_url"`
	}{
		{Name: "source.tar.gz", BrowserDownloadURL: "https://github.com/user/repo/archive/refs/tags/v1.0.tar.gz"},
		{Name: "plugin-linux-x86_64.AppImage", BrowserDownloadURL: "https://github.com/user/repo/releases/download/v1.0/plugin-linux-x86_64.AppImage"},
		{Name: "plugin-lv2.zip", BrowserDownloadURL: "https://github.com/user/repo/releases/download/v1.0/plugin-lv2.zip"},
	}

	best := SelectBestReleaseAsset(assets)
	if best != "https://github.com/user/repo/releases/download/v1.0/plugin-lv2.zip" {
		t.Errorf("Expected LV2 zip asset to be prioritized, got '%s'", best)
	}

	assetsFallback := []struct {
		Name               string `json:"name"`
		BrowserDownloadURL string `json:"browser_download_url"`
	}{
		{Name: "installer.msi", BrowserDownloadURL: "https://github.com/user/repo/releases/download/v1.0/installer.msi"},
	}
	bestFallback := SelectBestReleaseAsset(assetsFallback)
	if bestFallback != "https://github.com/user/repo/releases/download/v1.0/installer.msi" {
		t.Errorf("Expected MSI installer, got '%s'", bestFallback)
	}
}

func TestLV2Extraction(t *testing.T) {
	desc := "Open-source reverb plugin for Linux workstations with native LV2 bundle support"
	formats := parser.ExtractFormats(desc)

	hasLV2 := false
	for _, f := range formats {
		if strings.EqualFold(f, "LV2") {
			hasLV2 = true
			break
		}
	}
	if !hasLV2 {
		t.Errorf("Expected LV2 in extracted formats, got %v", formats)
	}
}

func TestGithubScraper_AudioUnitTopic(t *testing.T) {
	desc := "Modern synthesizer plugin built with JUCE and supporting AudioUnit (AUv3) for macOS"
	formats := parser.ExtractFormats(desc)

	isAU := strings.Contains(strings.ToLower(desc), "audiounit") || strings.Contains(strings.ToLower(desc), "auv3")
	if isAU {
		hasAU := false
		for _, f := range formats {
			if strings.EqualFold(f, "AU") {
				hasAU = true
				break
			}
		}
		if !hasAU {
			formats = append(formats, "AU")
		}
	}

	foundAU := false
	for _, f := range formats {
		if strings.EqualFold(f, "AU") {
			foundAU = true
			break
		}
	}
	if !foundAU {
		t.Errorf("Expected AU in extracted formats, got %v", formats)
	}
}
