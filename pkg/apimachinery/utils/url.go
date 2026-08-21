package utils

import (
	"net/url"
	"strings"
)

// Subpath returns the Grafana application subpath (for example "/grafana") parsed from
// an application base URL such as "http://localhost:3000/grafana/". An empty string is
// returned when the app is served from the root or appURL has no path component. The
// returned value has no trailing slash. All trailing slashes on the path are stripped so
// concatenation never produces a repeated separator.
func Subpath(appURL string) string {
	if appURL == "" {
		return ""
	}
	u, err := url.Parse(appURL)
	if err != nil || u.Path == "" || u.Path == "/" {
		return ""
	}
	return strings.TrimRight(u.Path, "/")
}

// JoinShortURLTarget builds the path-relative redirect target for a short URL whose stored
// resource path is resourcePath. The application subpath (derived from appURL) is prepended
// exactly once, the resource path is normalized to remove a leading slash, and the result
// always begins with "/". This is used to build the HTTP Location header for the Kubernetes
// ShortURL /goto redirect.
func JoinShortURLTarget(appURL, resourcePath string) string {
	subpath := Subpath(appURL)
	p := strings.TrimPrefix(resourcePath, "/")
	if subpath == "" {
		return "/" + p
	}
	return subpath + "/" + p
}

// TrimTrailingSlash returns baseURL with any trailing slashes removed. Shared by the short URL
// DTO builders so the "/goto/<uid>" segment is joined without a repeated separator.
func TrimTrailingSlash(baseURL string) string {
	return strings.TrimSuffix(baseURL, "/")
}
