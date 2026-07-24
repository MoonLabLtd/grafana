package util

import (
	"net/http"
	"net/url"
	"strings"
)

// DynamicRootURLService provides functionality to detect and validate
// the root URL from incoming HTTP requests based on trusted origins.
type DynamicRootURLService struct {
	trustedOrigins map[string]struct{}
	defaultRootURL string
	appSubPath     string
}

// NewDynamicRootURLService creates a new service for dynamic root URL detection.
func NewDynamicRootURLService(trustedOrigins []string, defaultRootURL, appSubPath string) *DynamicRootURLService {
	originsMap := make(map[string]struct{})
	for _, origin := range trustedOrigins {
		origin = strings.TrimSpace(origin)
		if origin != "" {
			// Store both with and without trailing slash for flexible matching
			originsMap[origin] = struct{}{}
			originsMap[strings.TrimSuffix(origin, "/")] = struct{}{}
		}
	}

	return &DynamicRootURLService{
		trustedOrigins: originsMap,
		defaultRootURL: defaultRootURL,
		appSubPath:     appSubPath,
	}
}

// DetectRootURL attempts to determine the root URL from the incoming request.
// It checks the X-Forwarded-Host, X-Forwarded-Proto, and Host headers to construct
// the URL. If the detected origin is not in the trusted origins list, it falls back
// to the default root URL.
func (s *DynamicRootURLService) DetectRootURL(r *http.Request) string {
	// Get the scheme (protocol)
	scheme := s.detectScheme(r)

	// Get the host
	host := s.detectHost(r)

	if host == "" {
		return s.defaultRootURL
	}

	// Construct the detected URL
	detectedURL := scheme + "://" + host

	// Check if the detected origin is trusted
	if !s.isTrustedOrigin(detectedURL) {
		return s.defaultRootURL
	}

	// Append subpath if configured
	if s.appSubPath != "" && s.appSubPath != "/" {
		detectedURL += s.appSubPath
	}

	// Ensure trailing slash
	if !strings.HasSuffix(detectedURL, "/") {
		detectedURL += "/"
	}

	return detectedURL
}

// detectScheme determines the URL scheme from the request.
// It checks X-Forwarded-Proto first, then falls back to inferring from the request.
func (s *DynamicRootURLService) detectScheme(r *http.Request) string {
	// Check X-Forwarded-Proto header (standard proxy header)
	if proto := r.Header.Get("X-Forwarded-Proto"); proto != "" {
		proto = strings.ToLower(strings.TrimSpace(proto))
		if proto == "http" || proto == "https" {
			return proto
		}
	}

	// Check X-Forwarded-SSL header (some proxies use this)
	if ssl := r.Header.Get("X-Forwarded-SSL"); strings.EqualFold(ssl, "on") {
		return "https"
	}

	// Check if TLS is enabled on the connection
	if r.TLS != nil {
		return "https"
	}

	// Default to http
	return "http"
}

// detectHost determines the host from the request.
// It checks X-Forwarded-Host first, then falls back to the Host header.
func (s *DynamicRootURLService) detectHost(r *http.Request) string {
	// Check X-Forwarded-Host header (standard proxy header)
	if fwdHost := r.Header.Get("X-Forwarded-Host"); fwdHost != "" {
		return strings.TrimSpace(fwdHost)
	}

	// Fall back to the Host header
	if r.Host != "" {
		return r.Host
	}

	return ""
}

// isTrustedOrigin checks if the detected origin is in the trusted origins list.
// If no trusted origins are configured, it uses the default root URL's origin.
func (s *DynamicRootURLService) isTrustedOrigin(detectedURL string) bool {
	// Parse the detected URL to get the origin (scheme + host)
	parsed, err := url.Parse(detectedURL)
	if err != nil {
		return false
	}

	// Construct origin (scheme://host[:port])
	origin := parsed.Scheme + "://" + parsed.Host

	// Check if in trusted origins
	if _, ok := s.trustedOrigins[origin]; ok {
		return true
	}

	// Check without port if origin has a port
	if parsed.Port() != "" {
		originNoPort := parsed.Scheme + "://" + parsed.Hostname()
		if _, ok := s.trustedOrigins[originNoPort]; ok {
			return true
		}
	}

	// If no trusted origins configured, check against default root URL
	if len(s.trustedOrigins) == 0 {
		defaultParsed, err := url.Parse(s.defaultRootURL)
		if err != nil {
			return false
		}
		defaultOrigin := defaultParsed.Scheme + "://" + defaultParsed.Host
		return origin == defaultOrigin
	}

	return false
}

// IsTrustedOrigin is a public helper to check if an origin string is trusted.
func (s *DynamicRootURLService) IsTrustedOrigin(origin string) bool {
	_, ok := s.trustedOrigins[origin]
	return ok
}
