package rooturl

import (
	"crypto/tls"
	"net/http/httptest"
	"testing"

	"github.com/stretchr/testify/assert"

	"github.com/grafana/grafana/pkg/setting"
)

func TestGetRootURL_StaticMode(t *testing.T) {
	cfg := &setting.Cfg{
		AppURL:              "http://static.example.com:3000/",
		AppSubURL:           "",
		RootURLMode:         RootURLModeStatic,
		RootURLTrustedOrigins: []string{},
	}

	svc := ProvideService(cfg)

	// Should return static URL in static mode
	req := httptest.NewRequest("GET", "http://dynamic.example.com/test", nil)
	rootURL := svc.GetRootURL(req)
	assert.Equal(t, "http://static.example.com:3000/", rootURL)

	// Should return static URL even when request is nil
	rootURL = svc.GetRootURL(nil)
	assert.Equal(t, "http://static.example.com:3000/", rootURL)
}

func TestGetRootURL_DynamicMode_NoTrustedOrigins(t *testing.T) {
	cfg := &setting.Cfg{
		AppURL:              "http://static.example.com:3000/",
		AppSubURL:           "",
		RootURLMode:         RootURLModeDynamic,
		RootURLTrustedOrigins: []string{},
	}

	svc := ProvideService(cfg)

	// Should fall back to static URL when no trusted origins configured
	req := httptest.NewRequest("GET", "http://dynamic.example.com/test", nil)
	rootURL := svc.GetRootURL(req)
	assert.Equal(t, "http://static.example.com:3000/", rootURL)
}

func TestGetRootURL_DynamicMode_WithTrustedOrigins(t *testing.T) {
	cfg := &setting.Cfg{
		AppURL:              "http://static.example.com:3000/",
		AppSubURL:           "",
		RootURLMode:         RootURLModeDynamic,
		RootURLTrustedOrigins: []string{"http://dynamic.example.com:3000"},
	}

	svc := ProvideService(cfg)

	req := httptest.NewRequest("GET", "http://dynamic.example.com:3000/test", nil)
	rootURL := svc.GetRootURL(req)
	assert.Equal(t, "http://dynamic.example.com:3000/", rootURL)
}

func TestGetRootURL_DynamicMode_WithProxyHeaders(t *testing.T) {
	cfg := &setting.Cfg{
		AppURL:              "http://static.example.com:3000/",
		AppSubURL:           "",
		RootURLMode:         RootURLModeDynamic,
		RootURLTrustedOrigins: []string{"https://public.example.com"},
	}

	svc := ProvideService(cfg)

	req := httptest.NewRequest("GET", "http://internal.example.com/test", nil)
	req.Header.Set("X-Forwarded-Proto", "https")
	req.Header.Set("X-Forwarded-Host", "public.example.com")
	
	rootURL := svc.GetRootURL(req)
	assert.Equal(t, "https://public.example.com/", rootURL)
}

func TestGetRootURL_DynamicMode_UntrustedOrigin(t *testing.T) {
	cfg := &setting.Cfg{
		AppURL:              "http://static.example.com:3000/",
		AppSubURL:           "",
		RootURLMode:         RootURLModeDynamic,
		RootURLTrustedOrigins: []string{"http://trusted.example.com"},
	}

	svc := ProvideService(cfg)

	// Should fall back to static URL for untrusted origin
	req := httptest.NewRequest("GET", "http://untrusted.example.com/test", nil)
	rootURL := svc.GetRootURL(req)
	assert.Equal(t, "http://static.example.com:3000/", rootURL)
}

func TestGetRootURL_DynamicMode_WithSubpath(t *testing.T) {
	cfg := &setting.Cfg{
		AppURL:              "http://static.example.com:3000/grafana/",
		AppSubURL:           "/grafana",
		RootURLMode:         RootURLModeDynamic,
		RootURLTrustedOrigins: []string{"http://dynamic.example.com:3000"},
	}

	svc := ProvideService(cfg)

	req := httptest.NewRequest("GET", "http://dynamic.example.com:3000/grafana/test", nil)
	rootURL := svc.GetRootURL(req)
	assert.Equal(t, "http://dynamic.example.com:3000/grafana/", rootURL)
}

func TestGetRootURL_DynamicMode_TLSDetection(t *testing.T) {
	cfg := &setting.Cfg{
		AppURL:              "http://static.example.com:3000/",
		AppSubURL:           "",
		RootURLMode:         RootURLModeDynamic,
		RootURLTrustedOrigins: []string{"https://secure.example.com"},
	}

	svc := ProvideService(cfg)

	// Test TLS detection
	req := httptest.NewRequest("GET", "https://secure.example.com/test", nil)
	req.TLS = &tls.ConnectionState{}
	rootURL := svc.GetRootURL(req)
	assert.Equal(t, "https://secure.example.com/", rootURL)
}

func TestIsTrustedOrigin(t *testing.T) {
	tests := []struct {
		name           string
		trustedOrigins []string
		rootURL        string
		expected       bool
	}{
		{
			name:           "exact match",
			trustedOrigins: []string{"http://example.com"},
			rootURL:        "http://example.com/",
			expected:       true,
		},
		{
			name:           "no match",
			trustedOrigins: []string{"http://trusted.com"},
			rootURL:        "http://untrusted.com/",
			expected:       false,
		},
		{
			name:           "different port",
			trustedOrigins: []string{"http://example.com:3000"},
			rootURL:        "http://example.com:8080/",
			expected:       false,
		},
		{
			name:           "different scheme",
			trustedOrigins: []string{"http://example.com"},
			rootURL:        "https://example.com/",
			expected:       false,
		},
		{
			name:           "case insensitive match",
			trustedOrigins: []string{"HTTP://Example.COM"},
			rootURL:        "http://example.com/",
			expected:       true,
		},
		{
			name:           "multiple trusted origins - match",
			trustedOrigins: []string{"http://a.com", "http://b.com", "http://c.com"},
			rootURL:        "http://b.com/",
			expected:       true,
		},
		{
			name:           "invalid URL in trusted origins",
			trustedOrigins: []string{"not a url", "http://valid.com"},
			rootURL:        "http://valid.com/",
			expected:       true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			cfg := &setting.Cfg{
				RootURLTrustedOrigins: tt.trustedOrigins,
			}

			svc := &service{
				cfg: cfg,
			}

			result := svc.isTrustedOrigin(tt.rootURL)
			assert.Equal(t, tt.expected, result)
		})
	}
}
