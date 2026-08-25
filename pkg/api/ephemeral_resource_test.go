package api

import (
	"context"
	"encoding/json"
	"io"
	"net/http"
	"path/filepath"
	"strings"
	"testing"

	"github.com/grafana/grafana-azure-sdk-go/v2/azsettings"
	"github.com/stretchr/testify/require"

	"github.com/grafana/grafana/pkg/infra/db"
	"github.com/grafana/grafana/pkg/infra/localcache"
	"github.com/grafana/grafana/pkg/infra/log"
	"github.com/grafana/grafana/pkg/infra/tracing"
	"github.com/grafana/grafana/pkg/services/accesscontrol"
	"github.com/grafana/grafana/pkg/services/datasources"
	datasourcesfakes "github.com/grafana/grafana/pkg/services/datasources/fakes"
	"github.com/grafana/grafana/pkg/services/featuremgmt"
	"github.com/grafana/grafana/pkg/services/pluginsintegration"
	"github.com/grafana/grafana/pkg/services/pluginsintegration/coreplugin"
	"github.com/grafana/grafana/pkg/services/pluginsintegration/pluginaccesscontrol"
	"github.com/grafana/grafana/pkg/services/pluginsintegration/pluginconfig"
	"github.com/grafana/grafana/pkg/services/pluginsintegration/plugincontext"
	pluginSettings "github.com/grafana/grafana/pkg/services/pluginsintegration/pluginsettings/service"
	"github.com/grafana/grafana/pkg/services/quota/quotatest"
	fakeSecrets "github.com/grafana/grafana/pkg/services/secrets/fakes"
	"github.com/grafana/grafana/pkg/services/user"
	"github.com/grafana/grafana/pkg/setting"
	"github.com/grafana/grafana/pkg/tsdb/cloudwatch"
	testdatasource "github.com/grafana/grafana/pkg/tsdb/grafana-testdata-datasource"
	"github.com/grafana/grafana/pkg/util/testutil"
	"github.com/grafana/grafana/pkg/web/webtest"
)

// dataSourceRequestValidatorStub accepts all datasource resource requests.
type dataSourceRequestValidatorStub struct{}

func (*dataSourceRequestValidatorStub) Validate(string, map[string]any, *http.Request) error {
	return nil
}

func TestIntegrationCallDatasourceResourceWithEphemeralSettings(t *testing.T) {
	testutil.SkipIntegrationTestInShortMode(t)

	staticRootPath, err := filepath.Abs("../../public/")
	require.NoError(t, err)

	cfg := setting.NewCfg()
	cfg.StaticRootPath = staticRootPath
	cfg.Azure = &azsettings.AzureSettings{}

	coreRegistry := coreplugin.ProvideCoreRegistry(tracing.InitializeTracerForTest(), nil, &cloudwatch.Service{}, nil, nil, nil,
		nil, testdatasource.ProvideService(), nil, nil, nil, nil)

	testCtx := pluginsintegration.CreateIntegrationTestCtx(t, cfg, coreRegistry)

	pcp := plugincontext.ProvideService(cfg, localcache.ProvideService(), testCtx.PluginStore, &datasourcesfakes.FakeCacheService{},
		&datasourcesfakes.FakeDataSourceService{}, pluginSettings.ProvideService(db.InitTestDB(t), fakeSecrets.NewFakeSecretsService()), pluginconfig.NewFakePluginRequestConfigProvider())

	queryPermission := func() map[int64]map[string][]string {
		return map[int64]map[string][]string{
			1: accesscontrol.GroupScopesByActionContext(context.Background(), []accesscontrol.Permission{
				{Action: datasources.ActionQuery, Scope: datasources.ScopeAll},
			}),
		}
	}

	setup := func(featuresEnabled bool) *webtest.Server {
		var features featuremgmt.FeatureToggles
		if featuresEnabled {
			features = featuremgmt.WithFeatures(featuremgmt.FlagUnsavedDatasourceResourceLookup)
		} else {
			features = featuremgmt.WithFeatures()
		}
		return SetupAPITestServer(t, func(hs *HTTPServer) {
			hs.Cfg = cfg
			hs.Features = features
			hs.pluginContextProvider = pcp
			hs.QuotaService = quotatest.New(false, nil)
			hs.pluginStore = testCtx.PluginStore
			hs.pluginClient = testCtx.PluginClient
			hs.DataSourceRequestValidator = &dataSourceRequestValidatorStub{}
			hs.log = log.New("test")
		})
	}

	t.Run("disabled feature toggle returns 400", func(t *testing.T) {
		server := setup(false)
		req := server.NewPostRequest("/api/datasources/uid/__ephemeral__/resources/test",
			strings.NewReader(`{"type":"grafana-testdata-datasource"}`))
		webtest.RequestWithSignedInUser(req, &user.SignedInUser{UserID: 1, OrgID: 1, Permissions: queryPermission()})
		resp, err := server.SendJSON(req)
		require.NoError(t, err)
		defer resp.Body.Close()
		require.Equal(t, http.StatusBadRequest, resp.StatusCode)
	})

	t.Run("successful resource lookup on unsaved settings", func(t *testing.T) {
		server := setup(true)
		body := `{"type":"grafana-testdata-datasource","access":"proxy","url":"","jsonData":{"defaultRegion":"eu-central-1"},"secureJsonData":{"accessKey":"AKIA","secretKey":"secret"}}`
		req := server.NewPostRequest("/api/datasources/uid/__ephemeral__/resources/test", strings.NewReader(body))
		webtest.RequestWithSignedInUser(req, &user.SignedInUser{UserID: 1, OrgID: 1, Permissions: queryPermission()})
		resp, err := server.SendJSON(req)
		require.NoError(t, err)
		defer resp.Body.Close()

		b, err := io.ReadAll(resp.Body)
		require.NoError(t, err)
		var result map[string]any
		require.NoError(t, json.Unmarshal(b, &result))

		require.Equal(t, http.StatusOK, resp.StatusCode)
		require.Equal(t, "Hello world from test datasource!", result["message"])
	})

	t.Run("missing type returns 400", func(t *testing.T) {
		server := setup(true)
		req := server.NewPostRequest("/api/datasources/uid/__ephemeral__/resources/test", strings.NewReader(`{}`))
		webtest.RequestWithSignedInUser(req, &user.SignedInUser{UserID: 1, OrgID: 1, Permissions: queryPermission()})
		resp, err := server.SendJSON(req)
		require.NoError(t, err)
		defer resp.Body.Close()
		require.Equal(t, http.StatusBadRequest, resp.StatusCode)
	})

	t.Run("invalid plugin type returns 400", func(t *testing.T) {
		server := setup(true)
		req := server.NewPostRequest("/api/datasources/uid/__ephemeral__/resources/test",
			strings.NewReader(`{"type":"does-not-exist-plugin"}`))
		webtest.RequestWithSignedInUser(req, &user.SignedInUser{UserID: 1, OrgID: 1, Permissions: queryPermission()})
		resp, err := server.SendJSON(req)
		require.NoError(t, err)
		defer resp.Body.Close()
		require.Equal(t, http.StatusBadRequest, resp.StatusCode)
	})

	t.Run("insufficient permissions returns 403", func(t *testing.T) {
		server := setup(true)
		req := server.NewPostRequest("/api/datasources/uid/__ephemeral__/resources/test",
			strings.NewReader(`{"type":"grafana-testdata-datasource"}`))
		// Only app-access permission, no datasources:query.
		webtest.RequestWithSignedInUser(req, &user.SignedInUser{UserID: 1, OrgID: 1, Permissions: map[int64]map[string][]string{
			1: accesscontrol.GroupScopesByActionContext(context.Background(), []accesscontrol.Permission{
				{Action: pluginaccesscontrol.ActionAppAccess, Scope: pluginaccesscontrol.ScopeProvider.GetResourceAllScope()},
			}),
		}})
		resp, err := server.SendJSON(req)
		require.NoError(t, err)
		defer resp.Body.Close()
		require.Equal(t, http.StatusForbidden, resp.StatusCode)
	})
}
