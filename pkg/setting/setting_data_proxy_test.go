package setting

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestDataProxyForwardAuthHeaders(t *testing.T) {
	t.Run("should default to true when the key is not present", func(t *testing.T) {
		cfg := NewCfg()

		err := readDataProxySettings(cfg.Raw, cfg)
		require.NoError(t, err)

		assert.True(t, cfg.DataProxyForwardAuthHeaders)
	})

	t.Run("should parse accepted boolean values", func(t *testing.T) {
		testCases := []struct {
			name          string
			value         string
			expectedValue bool
		}{
			{name: "should parse true", value: "true", expectedValue: true},
			{name: "should parse 1", value: "1", expectedValue: true},
			{name: "should parse yes", value: "yes", expectedValue: true},
			{name: "should parse on", value: "on", expectedValue: true},
			{name: "should parse false", value: "false", expectedValue: false},
			{name: "should parse 0", value: "0", expectedValue: false},
			{name: "should parse no", value: "no", expectedValue: false},
			{name: "should parse off", value: "off", expectedValue: false},
		}

		for _, tc := range testCases {
			t.Run(tc.name, func(t *testing.T) {
				cfg := NewCfg()

				dataproxy, err := cfg.Raw.NewSection("dataproxy")
				require.NoError(t, err)
				_, err = dataproxy.NewKey("forward_grafana_auth_headers", tc.value)
				require.NoError(t, err)

				err = readDataProxySettings(cfg.Raw, cfg)
				require.NoError(t, err)

				assert.Equal(t, tc.expectedValue, cfg.DataProxyForwardAuthHeaders)
			})
		}
	})

	t.Run("should use the default value when the configured value is invalid", func(t *testing.T) {
		cfg := NewCfg()

		dataproxy, err := cfg.Raw.NewSection("dataproxy")
		require.NoError(t, err)
		_, err = dataproxy.NewKey("forward_grafana_auth_headers", "invalid-bool")
		require.NoError(t, err)

		err = readDataProxySettings(cfg.Raw, cfg)
		require.NoError(t, err)

		assert.True(t, cfg.DataProxyForwardAuthHeaders)
	})

	t.Run("should propagate the value into the configuration", func(t *testing.T) {
		cfg, err := NewCfgFromBytes([]byte(`
[dataproxy]
forward_grafana_auth_headers = false
`))
		require.NoError(t, err)

		assert.False(t, cfg.DataProxyForwardAuthHeaders)
	})
}
