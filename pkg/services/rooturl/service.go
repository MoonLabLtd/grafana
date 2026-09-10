package rooturl

import (
	"fmt"
	"net/http"
	"net/url"
	"strings"

	"github.com/grafana/grafana/pkg/infra/log"
	"github.com/grafana/grafana/pkg/setting"
)

const (
	RootURLModeStatic  = "static"
	RootURLModeDynamic = "dynamic"
)

type Service interface {
	// GetRootURL returns the root URL for the given request.
	// If dynamic mode is enabled, it extracts the URL from the request and validates it against trusted origins.
	// If static mode is enabled or validation fails, it falls back to the configured AppURL.
	GetRootURL(r *http.Request) string
}

type service struct {
	cfg  *setting.Cfg
	log  log.Logger
}

func ProvideService(cfg *setting.Cfg) Service {
	return &service{
		cfg: cfg,
		log: log.New("rooturl"),
	}
}

func (s *service) GetRootURL(r *http.Request) string {
	// If static mode or no request, return the configured AppURL
	if s.cfg.RootURLMode == RootURLModeStatic || r == nil {
		return s.cfg.AppURL
	}

	// Extract the scheme
	scheme := "http"
	if r.TLS != nil {
		scheme = "https"
	}
	
	// Check for proxy headers
	if forwardedProto := r.Header.Get("X-Forwarded-Proto"); forwardedProto != "" {
		scheme = forwardedProto
	}

	// Extract the host
	host := r.Host
	if forwardedHost := r.Header.Get("X-Forwarded-Host"); forwardedHost != "" {
		host = forwardedHost
	}

	// Construct the root URL
	rootURL := fmt.Sprintf("%s://%s", scheme, host)
	
	// Add the subpath if configured
	if s.cfg.AppSubURL != "" {
		rootURL += s.cfg.AppSubURL
	}

	// Ensure trailing slash
	if !strings.HasSuffix(rootURL, "/") {
		rootURL += "/"
	}

	// Validate against trusted origins
	if !s.isTrustedOrigin(rootURL) {
		s.log.Debug("Dynamic root URL not in trusted origins, falling back to static", 
			"dynamic", rootURL, 
			"static", s.cfg.AppURL,
			"trusted_origins", s.cfg.RootURLTrustedOrigins)
		return s.cfg.AppURL
	}

	s.log.Debug("Using dynamic root URL", "url", rootURL)
	return rootURL
}

func (s *service) isTrustedOrigin(rootURL string) bool {
	// If no trusted origins configured, reject all dynamic URLs
	if len(s.cfg.RootURLTrustedOrigins) == 0 {
		return false
	}

	// Parse the root URL
	parsedURL, err := url.Parse(rootURL)
	if err != nil {
		return false
	}

	// Check if the origin matches any trusted origin
	origin := fmt.Sprintf("%s://%s", parsedURL.Scheme, parsedURL.Host)
	for _, trusted := range s.cfg.RootURLTrustedOrigins {
		// Normalize trusted origin
		trusted = strings.TrimSpace(trusted)
		if trusted == "" {
			continue
		}
		
		// Parse trusted origin
		trustedURL, err := url.Parse(trusted)
		if err != nil {
			continue
		}

		// Compare scheme and host (case-insensitive)
		if strings.EqualFold(parsedURL.Scheme, trustedURL.Scheme) &&
			strings.EqualFold(parsedURL.Host, trustedURL.Host) {
			return true
		}
	}

	return false
}
