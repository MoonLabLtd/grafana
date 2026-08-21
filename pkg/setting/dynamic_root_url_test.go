package setting

import (
	"testing"

	"gopkg.in/ini.v1"

	"github.com/stretchr/testify/require"
)

const dynamicRootURLTestINI = `
[server]
root_url = https://internal.example/grafana/

[security]
csrf_trusted_origins = internal.example public.example
`

func newDynamicRootURLCfg(t *testing.T, iniContent string) *Cfg {
	t.Helper()
	cfg, err := NewCfgFromBytes([]byte(iniContent))
	require.NoError(t, err)
	return cfg
}

func TestResolveRootURL(t *testing.T) {

	tests := []struct {
		name          string
		iniContent    string
		host          string
		expected      string
	}{
		{
			name:       "feature disabled uses static root url",
			iniContent: dynamicRootURLTestINI,
			host:       "public.example",
			expected:   "https://internal.example/grafana/",
		},
		{
			name:       "enabled with trusted internal host resolves dynamically",
			iniContent: "[server]\nroot_url = https://internal.example/grafana/\ndynamic_root_url_enabled = true\n\n[security]\ncsrf_trusted_origins = internal.example public.example\n",
			host:       "internal.example",
			expected:   "https://internal.example/grafana/",
		},
		{
			name:       "enabled with trusted public host resolves dynamically",
			iniContent: "[server]\nroot_url = https://internal.example/grafana/\ndynamic_root_url_enabled = true\n\n[security]\ncsrf_trusted_origins = internal.example public.example\n",
			host:       "public.example",
			expected:   "https://public.example/grafana/",
		},
		{
			name:       "enabled with untrusted host falls back to static",
			iniContent: "[server]\nroot_url = https://internal.example/grafana/\ndynamic_root_url_enabled = true\n\n[security]\ncsrf_trusted_origins = internal.example public.example\n",
			host:       "untrusted.example",
			expected:   "https://internal.example/grafana/",
		},
		{
			name:       "missing host falls back to static",
			iniContent: "[server]\nroot_url = https://internal.example/grafana/\ndynamic_root_url_enabled = true\n\n[security]\ncsrf_trusted_origins = internal.example public.example\n",
			host:       "",
			expected:   "https://internal.example/grafana/",
		},
		{
			name:       "malformed host falls back to static",
			iniContent: "[server]\nroot_url = https://internal.example/grafana/\ndynamic_root_url_enabled = true\n\n[security]\ncsrf_trusted_origins = internal.example public.example\n",
			host:       "::%%%%:::",
			expected:   "https://internal.example/grafana/",
		},
		{
			name:       "empty trusted origins never resolves dynamically",
			iniContent: "[server]\nroot_url = https://internal.example/grafana/\ndynamic_root_url_enabled = true\n\n[security]\ncsrf_trusted_origins =\n",
			host:       "internal.example",
			expected:   "https://internal.example/grafana/",
		},
		{
			name:       "host with port matches trusted hostname",
			iniContent: "[server]\nroot_url = https://internal.example/grafana/\ndynamic_root_url_enabled = true\n\n[security]\ncsrf_trusted_origins = internal.example public.example\n",
			host:       "public.example:443",
			expected:   "https://public.example/grafana/",
		},
		{
			name:       "preserves static scheme when dynamic host resolved",
			iniContent: "[server]\nroot_url = http://internal.example/\ndynamic_root_url_enabled = true\n\n[security]\ncsrf_trusted_origins = public.example\n",
			host:       "public.example",
			expected:   "http://public.example/",
		},
		{
			name:       "root url without subpath resolves to bare host",
			iniContent: "[server]\nroot_url = https://internal.example/\ndynamic_root_url_enabled = true\n\n[security]\ncsrf_trusted_origins = public.example\n",
			host:       "public.example",
			expected:   "https://public.example/",
		},
		{
			name:       "does not match by substring",
			iniContent: "[server]\nroot_url = https://internal.example/grafana/\ndynamic_root_url_enabled = true\n\n[security]\ncsrf_trusted_origins = internal.example\n",
			host:       "internal.example.evil.com",
			expected:   "https://internal.example/grafana/",
		},
	}

	for _, tc := range tests {
		tc := tc
		t.Run(tc.name, func(t *testing.T) {
			cfg := newDynamicRootURLCfg(t, tc.iniContent)
			require.Equal(t, tc.expected, cfg.ResolveRootURL(tc.host))
		})
	}
}

func TestResolveRootURLDoesNotMutateGlobalConfig(t *testing.T) {

	cfg := newDynamicRootURLCfg(t, "[server]\nroot_url = https://internal.example/grafana/\ndynamic_root_url_enabled = true\n\n[security]\ncsrf_trusted_origins = internal.example public.example\n")

	require.Equal(t, "https://internal.example/grafana/", cfg.AppURL)
	resolved := cfg.ResolveRootURL("public.example")
	require.Equal(t, "https://public.example/grafana/", resolved)
	// Global config is unchanged.
	require.Equal(t, "https://internal.example/grafana/", cfg.AppURL)
}

func TestIsTrustedOrigin(t *testing.T) {

	cfg := newDynamicRootURLCfg(t, dynamicRootURLTestINI)

	require.True(t, cfg.IsTrustedOrigin("internal.example"))
	require.True(t, cfg.IsTrustedOrigin("public.example"))
	require.False(t, cfg.IsTrustedOrigin("untrusted.example"))
	require.False(t, cfg.IsTrustedOrigin(""))
}

func TestDynamicRootURLDisabledByDefault(t *testing.T) {

	cfg := newDynamicRootURLCfg(t, dynamicRootURLTestINI)
	require.False(t, cfg.DynamicRootURLEnabled)
}

func TestDynamicRootURLReadFromEnvironmentOverride(t *testing.T) {
	t.Setenv("GF_SERVER_DYNAMIC_ROOT_URL_ENABLED", "true")

	cfg := NewCfg()
	parsedFile, err := ini.Load([]byte(dynamicRootURLTestINI))
	require.NoError(t, err)
	require.NoError(t, cfg.applyEnvVariableOverrides(parsedFile))
	require.NoError(t, cfg.parseINIFile(parsedFile))

	require.True(t, cfg.DynamicRootURLEnabled)
	require.Equal(t, "https://public.example/grafana/", cfg.ResolveRootURL("public.example"))
}
