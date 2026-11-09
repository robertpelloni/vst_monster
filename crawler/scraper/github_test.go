package scraper

import (
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
