package main

// Golden fixture: TLS certificate verification disabled (VS-SEC insecure-default family).
// The "rule_id": null marker below is load-bearing for manifest validation — do not change.

import (
	"crypto/tls"
	"net/http"
)

func client() *http.Client {
	return &http.Client{
		Transport: &http.Transport{
			TLSClientConfig: &tls.Config{InsecureSkipVerify: true}, // rule_id: null
		},
	}
}

func main() {
	_, _ = client().Get("https://example.invalid")
}
