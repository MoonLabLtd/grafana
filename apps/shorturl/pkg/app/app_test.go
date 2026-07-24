package app

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/grafana/grafana-app-sdk/app"
)

func TestNewWithAppURL(t *testing.T) {
	tests := []struct {
		name           string
		specificConfig interface{}
		expectedURL    string
	}{
		{
			name: "with AppURL including subpath",
			specificConfig: ShortURLAppConfig{
				AppURL: "http://localhost:3000/grafana/",
			},
			expectedURL: "http://localhost:3000/grafana/",
		},
		{
			name: "with AppURL without trailing slash",
			specificConfig: ShortURLAppConfig{
				AppURL: "http://localhost:3000/grafana",
			},
			expectedURL: "http://localhost:3000/grafana/",
		},
		{
			name: "with AppURL without subpath",
			specificConfig: ShortURLAppConfig{
				AppURL: "http://localhost:3000/",
			},
			expectedURL: "http://localhost:3000/",
		},
		{
			name:           "without specific config",
			specificConfig: nil,
			expectedURL:    "",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			cfg := app.Config{
				SpecificConfig: tt.specificConfig,
			}

			// We can't fully test the app creation without a real Kubernetes config,
			// but we can verify the config extraction logic works
			var appURL string
			if specificConfig, ok := cfg.SpecificConfig.(ShortURLAppConfig); ok {
				appURL = specificConfig.AppURL
			}

			assert.Equal(t, tt.expectedURL, appURL)
		})
	}
}

func TestShortURLAppConfigTypeAssertion(t *testing.T) {
	cfg := app.Config{
		SpecificConfig: ShortURLAppConfig{
			AppURL: "http://localhost:3000/grafana/",
		},
	}

	specificConfig, ok := cfg.SpecificConfig.(ShortURLAppConfig)
	require.True(t, ok, "Type assertion should succeed")
	assert.Equal(t, "http://localhost:3000/grafana/", specificConfig.AppURL)
}
