package utils

import (
	"testing"

	"github.com/stretchr/testify/require"
)

func TestSubpath(t *testing.T) {
	tests := []struct {
		name   string
		appURL string
		expect string
	}{
		{name: "no trailing slash", appURL: "http://localhost:3000/grafana", expect: "/grafana"},
		{name: "trailing slash", appURL: "http://localhost:3000/grafana/", expect: "/grafana"},
		{name: "repeated trailing slashes", appURL: "http://localhost:3000/grafana//", expect: "/grafana"},
		{name: "nested subpath", appURL: "https://example.com/org/grafana/", expect: "/org/grafana"},
		{name: "root path", appURL: "http://localhost:3000/", expect: ""},
		{name: "no path component", appURL: "http://localhost:3000", expect: ""},
		{name: "empty", appURL: "", expect: ""},
	}
	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			require.Equal(t, tc.expect, Subpath(tc.appURL))
		})
	}
}

func TestJoinShortURLTarget(t *testing.T) {
	tests := []struct {
		name         string
		appURL       string
		resourcePath string
		expect       string
	}{
		{name: "subpath + unix path", appURL: "http://localhost:3000/grafana/", resourcePath: "d/abc/dashboard", expect: "/grafana/d/abc/dashboard"},
		{name: "subpath + leading slash path no double slash", appURL: "http://localhost:3000/grafana", resourcePath: "/d/abc/dashboard", expect: "/grafana/d/abc/dashboard"},
		{name: "root app no subpath", appURL: "http://localhost:3000/", resourcePath: "d/abc/dashboard", expect: "/d/abc/dashboard"},
		{name: "no path no subpath", appURL: "http://localhost:3000", resourcePath: "d/abc/dashboard", expect: "/d/abc/dashboard"},
		{name: "nested subpath", appURL: "https://example.com/org/grafana/", resourcePath: "d/abc/dashboard", expect: "/org/grafana/d/abc/dashboard"},
	}
	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			require.Equal(t, tc.expect, JoinShortURLTarget(tc.appURL, tc.resourcePath))
		})
	}
}

func TestTrimTrailingSlash(t *testing.T) {
	require.Equal(t, "http://localhost:3000/grafana", TrimTrailingSlash("http://localhost:3000/grafana/"))
	require.Equal(t, "http://localhost:3000/grafana", TrimTrailingSlash("http://localhost:3000/grafana"))
	require.Equal(t, "http://localhost:3000", TrimTrailingSlash("http://localhost:3000/"))
}
