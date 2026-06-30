package main

import (
	"os"
	"testing"
)

func TestGetProxySwitcher(t *testing.T) {
	// Test empty PROXY_LIST
	os.Setenv("PROXY_LIST", "")
	proxyFunc := GetProxySwitcher()
	if proxyFunc != nil {
		t.Error("Expected nil proxyFunc for empty PROXY_LIST")
	}

	// Test single proxy
	os.Setenv("PROXY_LIST", "http://proxy1.com")
	proxyFunc = GetProxySwitcher()
	if proxyFunc == nil {
		t.Error("Expected valid proxyFunc for single proxy")
	}

	// Test multiple proxies with whitespace
	os.Setenv("PROXY_LIST", "http://proxy1.com, http://proxy2.com ,  http://proxy3.com  ")
	proxyFunc = GetProxySwitcher()
	if proxyFunc == nil {
		t.Error("Expected valid proxyFunc for multiple proxies")
	}

	// Test invalid proxy string format that are just empty spaces
	os.Setenv("PROXY_LIST", " ,  ,   ")
	proxyFunc = GetProxySwitcher()
	if proxyFunc != nil {
		t.Error("Expected nil proxyFunc for empty/whitespace only proxies")
	}

	os.Unsetenv("PROXY_LIST")
}