package util

import (
	"net/http"
	"testing"
)

func TestDynamicRootURLService_DetectRootURL(t *testing.T) {
	tests := []struct {
		name               string
		trustedOrigins     []string
		defaultRootURL     string
		appSubPath         string
		request            *http.Request
		expectedRootURL    string
	}{
		{
			name:            "no trusted origins - uses default",
			trustedOrigins:  []string{},
			defaultRootURL:  "http://localhost:3000/",
			appSubPath:      "",
			request:         newTestRequest("https://external.example.com", "external.example.com", nil),
			expectedRootURL: "http://localhost:3000/",
		},
		{
			name:            "trusted origin matches - returns detected URL",
			trustedOrigins:  []string{"https://external.example.com"},
			defaultRootURL:  "http://localhost:3000/",
			appSubPath:      "",
			request:         newTestRequest("https", "external.example.com", nil),
			expectedRootURL: "https://external.example.com/",
		},
		{
			name:            "trusted origin with port matches",
			trustedOrigins:  []string{"https://external.example.com:443"},
			defaultRootURL:  "http://localhost:3000/",
			appSubPath:      "",
			request:         newTestRequest("https", "external.example.com:443", nil),
			expectedRootURL: "https://external.example.com:443/",
		},
		{
			name:            "trusted origin without port matches request with port",
			trustedOrigins:  []string{"https://external.example.com"},
			defaultRootURL:  "http://localhost:3000/",
			appSubPath:      "",
			request:         newTestRequest("https", "external.example.com:443", nil),
			expectedRootURL: "https://external.example.com:443/",
		},
		{
			name:            "detected origin not trusted - falls back to default",
			trustedOrigins:  []string{"https://trusted.example.com"},
			defaultRootURL:  "http://localhost:3000/",
			appSubPath:      "",
			request:         newTestRequest("https", "untrusted.example.com", nil),
			expectedRootURL: "http://localhost:3000/",
		},
		{
			name:            "with subpath",
			trustedOrigins:  []string{"https://external.example.com"},
			defaultRootURL:  "http://localhost:3000/grafana/",
			appSubPath:      "/grafana",
			request:         newTestRequest("https", "external.example.com", nil),
			expectedRootURL: "https://external.example.com/grafana/",
		},
		{
			name:            "uses X-Forwarded-Proto header",
			trustedOrigins:  []string{"https://external.example.com"},
			defaultRootURL:  "http://localhost:3000/",
			appSubPath:      "",
			request:         newTestRequestWithHeaders("", "external.example.com", map[string]string{"X-Forwarded-Proto": "https"}),
			expectedRootURL: "https://external.example.com/",
		},
		{
			name:            "uses X-Forwarded-Host header",
			trustedOrigins:  []string{"https://proxy.example.com"},
			defaultRootURL:  "http://localhost:3000/",
			appSubPath:      "",
			request:         newTestRequestWithHeaders("http", "internal.example.com", map[string]string{"X-Forwarded-Host": "proxy.example.com", "X-Forwarded-Proto": "https"}),
			expectedRootURL: "https://proxy.example.com/",
		},
		{
			name:            "X-Forwarded-SSL header",
			trustedOrigins:  []string{"https://external.example.com"},
			defaultRootURL:  "http://localhost:3000/",
			appSubPath:      "",
			request:         newTestRequestWithHeaders("", "external.example.com", map[string]string{"X-Forwarded-SSL": "on"}),
			expectedRootURL: "https://external.example.com/",
		},
		{
			name:            "no trusted origins - matches default root URL",
			trustedOrigins:  []string{},
			defaultRootURL:  "http://localhost:3000/",
			appSubPath:      "",
			request:         newTestRequest("http", "localhost:3000", nil),
			expectedRootURL: "http://localhost:3000/",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			service := NewDynamicRootURLService(tt.trustedOrigins, tt.defaultRootURL, tt.appSubPath)
			result := service.DetectRootURL(tt.request)
			if result != tt.expectedRootURL {
				t.Errorf("DetectRootURL() = %v, want %v", result, tt.expectedRootURL)
			}
		})
	}
}

func TestDynamicRootURLService_IsTrustedOrigin(t *testing.T) {
	service := NewDynamicRootURLService(
		[]string{"https://trusted.example.com", "http://localhost:3000"},
		"http://localhost:3000/",
		"",
	)

	tests := []struct {
		origin   string
		expected bool
	}{
		{"https://trusted.example.com", true},
		{"http://localhost:3000", true},
		{"https://untrusted.example.com", false},
		{"", false},
	}

	for _, tt := range tests {
		t.Run(tt.origin, func(t *testing.T) {
			result := service.IsTrustedOrigin(tt.origin)
			if result != tt.expected {
				t.Errorf("IsTrustedOrigin(%v) = %v, want %v", tt.origin, result, tt.expected)
			}
		})
	}
}

// Helper functions

func newTestRequest(scheme, host string, headers map[string]string) *http.Request {
	return newTestRequestWithHeaders(scheme, host, headers)
}

func newTestRequestWithHeaders(scheme, host string, headers map[string]string) *http.Request {
	req, _ := http.NewRequest("GET", "http://"+host+"/", nil)
	req.Host = host

	for k, v := range headers {
		req.Header.Set(k, v)
	}

	return req
}
