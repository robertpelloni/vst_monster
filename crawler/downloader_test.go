package main

import (
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
)

func TestCalculateSHA256(t *testing.T) {
	// Create a mock HTTP server
	mockData := "test data for hashing"
	// The sha256 for "test data for hashing" is:
	// f7eb7961d8a233e6256d3a6257548bbb9293c3a08fb3574c88c7d6b429dbb9f5
	expectedHash := "f7eb7961d8a233e6256d3a6257548bbb9293c3a08fb3574c88c7d6b429dbb9f5"

	ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
		w.Write([]byte(mockData))
	}))
	defer ts.Close()

	hash, err := CalculateSHA256(ts.URL)
	if err != nil {
		t.Fatalf("CalculateSHA256 returned error: %v", err)
	}

	if hash != expectedHash {
		t.Errorf("Expected hash %s, got %s", expectedHash, hash)
	}
}

func TestCalculateSHA256_EmptyURL(t *testing.T) {
	_, err := CalculateSHA256("")
	if err == nil {
		t.Error("Expected error for empty URL, got nil")
	}
	if !strings.Contains(err.Error(), "empty URL") {
		t.Errorf("Expected 'empty URL' error, got: %v", err)
	}
}

func TestCalculateSHA256_NotFound(t *testing.T) {
	ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusNotFound)
	}))
	defer ts.Close()

	_, err := CalculateSHA256(ts.URL)
	if err == nil {
		t.Error("Expected error for 404 response, got nil")
	}
	if !strings.Contains(err.Error(), "bad status code") {
		t.Errorf("Expected 'bad status code' error, got: %v", err)
	}
}