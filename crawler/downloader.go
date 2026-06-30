package main

import (
	"crypto/sha256"
	"encoding/hex"
	"fmt"
	"io"
	"log"
	"net/http"
	"os"
	"time"
)

var httpClient = &http.Client{
	Timeout: 30 * time.Second,
}

// CalculateSHA256 downloads the file from the given URL to a temporary file,
// calculates its SHA256 hash, and then cleans up the temporary file.
func CalculateSHA256(url string) (string, error) {
	if url == "" {
		return "", fmt.Errorf("empty URL provided")
	}

	req, err := http.NewRequest("GET", url, nil)
	if err != nil {
		return "", fmt.Errorf("error creating request: %w", err)
	}

	req.Header.Set("User-Agent", "VST-Monster-Bot/1.0 (+https://vstmonster.com)")

	resp, err := httpClient.Do(req)
	if err != nil {
		return "", fmt.Errorf("error downloading file: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return "", fmt.Errorf("bad status code: %s", resp.Status)
	}

	// Create a temporary file
	tmpFile, err := os.CreateTemp("", "vst-plugin-*.tmp")
	if err != nil {
		return "", fmt.Errorf("error creating temporary file: %w", err)
	}
	defer os.Remove(tmpFile.Name()) // Clean up
	defer tmpFile.Close()

	// Write the body to the temporary file
	_, err = io.Copy(tmpFile, resp.Body)
	if err != nil {
		return "", fmt.Errorf("error writing to temporary file: %w", err)
	}

	// Rewind the file to calculate hash
	_, err = tmpFile.Seek(0, 0)
	if err != nil {
		return "", fmt.Errorf("error seeking temporary file: %w", err)
	}

	// Calculate hash
	hasher := sha256.New()
	if _, err := io.Copy(hasher, tmpFile); err != nil {
		return "", fmt.Errorf("error calculating hash: %w", err)
	}

	hashStr := hex.EncodeToString(hasher.Sum(nil))
	log.Printf("Calculated hash %s for %s\n", hashStr, url)
	return hashStr, nil
}
