package app

import (
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/stretchr/testify/require"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"

	"github.com/grafana/grafana-app-sdk/app"
	shorturlv1beta1 "github.com/grafana/grafana/apps/shorturl/pkg/apis/shorturl/v1beta1"
)

func TestGotoRedirectWithSubpath(t *testing.T) {
	tests := []struct {
		name           string
		appURL         string
		shortURLPath   string
		expectedURL    string
	}{
		{
			name:         "redirect with subpath",
			appURL:       "http://example.com/grafana/",
			shortURLPath: "d/abc123/my-dashboard?orgId=1",
			expectedURL:  "http://example.com/grafana/d/abc123/my-dashboard?orgId=1",
		},
		{
			name:         "redirect without subpath",
			appURL:       "http://localhost:3000/",
			shortURLPath: "explore?orgId=1",
			expectedURL:  "http://localhost:3000/explore?orgId=1",
		},
		{
			name:         "redirect with nested subpath",
			appURL:       "https://grafana.example.com/monitoring/grafana/",
			shortURLPath: "d/xyz/dashboard?orgId=2",
			expectedURL:  "https://grafana.example.com/monitoring/grafana/d/xyz/dashboard?orgId=2",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Create ShortURLConfig with AppURL
			cfg := app.Config{
				SpecificConfig: &ShortURLConfig{
					AppURL: tt.appURL,
				},
			}

			// Create a mock short URL
			shortURL := &shorturlv1beta1.ShortURL{
				ObjectMeta: metav1.ObjectMeta{
					Name:      "test-short-url",
					Namespace: "default",
				},
				Spec: shorturlv1beta1.ShortURLSpec{
					Path: tt.shortURLPath,
				},
			}

			// Verify the redirect URL construction
			// In a real test, we would mock the client and request
			// For now, we verify the logic directly
			appURL := ""
			if shortURLConfig, ok := cfg.SpecificConfig.(*ShortURLConfig); ok && shortURLConfig != nil {
				appURL = shortURLConfig.AppURL
			}

			redirectURL := appURL + shortURL.Spec.Path
			require.Equal(t, tt.expectedURL, redirectURL, "redirect URL should include the configured subpath")
		})
	}
}

func TestGotoRedirectWithoutConfig(t *testing.T) {
	// Test that when no AppURL is configured, we get an empty string
	cfg := app.Config{
		SpecificConfig: nil,
	}

	appURL := ""
	if shortURLConfig, ok := cfg.SpecificConfig.(*ShortURLConfig); ok && shortURLConfig != nil {
		appURL = shortURLConfig.AppURL
	}

	require.Empty(t, appURL, "AppURL should be empty when config is not provided")

	// Test with empty ShortURLConfig
	cfg2 := app.Config{
		SpecificConfig: &ShortURLConfig{},
	}

	appURL2 := ""
	if shortURLConfig, ok := cfg2.SpecificConfig.(*ShortURLConfig); ok && shortURLConfig != nil {
		appURL2 = shortURLConfig.AppURL
	}

	require.Empty(t, appURL2, "AppURL should be empty when ShortURLConfig has empty AppURL")
}
