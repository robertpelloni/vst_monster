package parser

import (
	"encoding/json"
	"strings"
)

// StandardPlugin maps to the expected schema in the Registry
type StandardPlugin struct {
	Name        string   `json:"name"`
	Developer   string   `json:"developer"`
	Version     string   `json:"version"`
	Formats     []string `json:"formats"` // VST, VST3, AU, CLAP
	DownloadURL string   `json:"download_url"`
	License     string   `json:"license"` // free, commercial, opensource
	Source      string   `json:"source"`
}

func ExtractFormats(description string) []string {
	var formats []string
	descLower := strings.ToLower(description)
	if strings.Contains(descLower, "vst3") {
		formats = append(formats, "VST3")
	} else if strings.Contains(descLower, "vst") {
		formats = append(formats, "VST")
	}
	if strings.Contains(descLower, "au") || strings.Contains(descLower, "audio unit") {
		formats = append(formats, "AU")
	}
	if strings.Contains(descLower, "clap") {
		formats = append(formats, "CLAP")
	}
	return formats
}

func ToJSON(plugin StandardPlugin) (string, error) {
	bytes, err := json.Marshal(plugin)
	return string(bytes), err
}
