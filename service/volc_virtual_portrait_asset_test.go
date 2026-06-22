package service

import (
	"fmt"
	"net/http"
	"net/http/httptest"
	"net/url"
	"os"
	"strings"
	"sync/atomic"
	"testing"

	"github.com/QuantumNous/new-api/common"
	"github.com/QuantumNous/new-api/model"
	"github.com/stretchr/testify/require"
	"gorm.io/gorm"
)

func TestValidateUserOwnedVirtualPortraitAssetAutoSyncsProcessingAsset(t *testing.T) {
	setupVolcVirtualPortraitAssetServiceTestDB(t)
	setupVolcPortraitAssetMockServer(t, func(w http.ResponseWriter, r *http.Request) {
		require.Equal(t, "GetAsset", r.URL.Query().Get("Action"))
		w.Header().Set("Content-Type", "application/json")
		_, _ = w.Write([]byte(`{"Result":{"Id":"asset-auto-sync","GroupId":"group-1","Status":"Active","AssetType":"Image","ProjectName":"default","URL":"https://example.com/preview.png"}}`))
	})

	require.NoError(t, model.DB.Create(&model.VirtualPortraitAsset{
		UserId:         42,
		ExternalUserID: "sub-a",
		Name:           "processing",
		VolcAssetID:    "asset-auto-sync",
		Status:         model.VirtualPortraitAssetStatusProcessing,
		VolcStatus:     "Processing",
		ProjectName:    "default",
		CreatedTime:    1,
		UpdatedTime:    1,
	}).Error)

	err := ValidateUserOwnedPortraitAssetByAssetIDForExternalUser(42, "asset://asset-auto-sync", "sub-a")
	require.NoError(t, err)

	asset, err := model.GetActiveUserVirtualPortraitAssetByAssetIDForExternalUser(42, "asset-auto-sync", "sub-a")
	require.NoError(t, err)
	require.Equal(t, model.VirtualPortraitAssetStatusActive, asset.Status)
	require.Equal(t, "Active", asset.VolcStatus)
	require.Equal(t, "https://example.com/preview.png", asset.PreviewURL)
	require.NotZero(t, asset.ReadyTime)
}

func TestValidateUserOwnedVirtualPortraitAssetAutoSyncKeepsExternalUserBoundary(t *testing.T) {
	setupVolcVirtualPortraitAssetServiceTestDB(t)
	var requestCount int32
	setupVolcPortraitAssetMockServer(t, func(w http.ResponseWriter, r *http.Request) {
		atomic.AddInt32(&requestCount, 1)
		w.Header().Set("Content-Type", "application/json")
		_, _ = w.Write([]byte(`{"Result":{"Id":"asset-isolated","Status":"Active"}}`))
	})

	require.NoError(t, model.DB.Create(&model.VirtualPortraitAsset{
		UserId:         42,
		ExternalUserID: "sub-b",
		Name:           "isolated",
		VolcAssetID:    "asset-isolated",
		Status:         model.VirtualPortraitAssetStatusProcessing,
		VolcStatus:     "Processing",
		ProjectName:    "default",
		CreatedTime:    1,
		UpdatedTime:    1,
	}).Error)

	err := ValidateUserOwnedPortraitAssetByAssetIDForExternalUser(42, "asset://asset-isolated", "sub-a")
	require.Error(t, err)
	require.Equal(t, int32(0), atomic.LoadInt32(&requestCount))
}

func TestValidateUserOwnedVirtualPortraitAssetAutoSyncSupportsLegacyRequestWithoutExternalUserID(t *testing.T) {
	setupVolcVirtualPortraitAssetServiceTestDB(t)
	setupVolcPortraitAssetMockServer(t, func(w http.ResponseWriter, r *http.Request) {
		require.Equal(t, "GetAsset", r.URL.Query().Get("Action"))
		w.Header().Set("Content-Type", "application/json")
		_, _ = w.Write([]byte(`{"Result":{"Id":"asset-legacy-compatible","GroupId":"group-2","Status":"Active","AssetType":"Image","ProjectName":"default","URL":"https://example.com/legacy.png"}}`))
	})

	require.NoError(t, model.DB.Create(&model.VirtualPortraitAsset{
		UserId:         42,
		ExternalUserID: "sub-legacy",
		Name:           "legacy-compatible",
		VolcAssetID:    "asset-legacy-compatible",
		Status:         model.VirtualPortraitAssetStatusProcessing,
		VolcStatus:     "Processing",
		ProjectName:    "default",
		CreatedTime:    1,
		UpdatedTime:    1,
	}).Error)

	err := ValidateUserOwnedPortraitAssetByAssetIDForExternalUser(42, "asset://asset-legacy-compatible", "")
	require.NoError(t, err)

	asset, err := model.GetActiveUserVirtualPortraitAssetByAssetIDForExternalUser(42, "asset-legacy-compatible", "sub-legacy")
	require.NoError(t, err)
	require.Equal(t, model.VirtualPortraitAssetStatusActive, asset.Status)
	require.Equal(t, "https://example.com/legacy.png", asset.PreviewURL)
}

func setupVolcPortraitAssetMockServer(t *testing.T, handler http.HandlerFunc) {
	server := httptest.NewServer(handler)
	t.Cleanup(server.Close)

	parsed, err := url.Parse(server.URL)
	require.NoError(t, err)
	setTestEnv(t, "VOLC_PORTRAIT_OPENAPI_SCHEME", parsed.Scheme)
	setTestEnv(t, "VOLC_PORTRAIT_OPENAPI_HOST", parsed.Host)
	setTestEnv(t, "VOLC_PORTRAIT_AK", "test-ak")
	setTestEnv(t, "VOLC_PORTRAIT_SK", "test-sk")
	setTestEnv(t, "VOLC_PORTRAIT_REGION", "cn-beijing")
}

func setupVolcVirtualPortraitAssetServiceTestDB(t *testing.T) {
	originalDB := model.DB
	originalLogDB := model.LOG_DB
	originalIsMasterNode := common.IsMasterNode
	originalSQLitePath := common.SQLitePath
	originalUsingSQLite := common.UsingSQLite
	originalUsingMySQL := common.UsingMySQL
	originalUsingPostgreSQL := common.UsingPostgreSQL
	originalDSN := os.Getenv("SQL_DSN")
	var testDB *gorm.DB
	t.Cleanup(func() {
		model.DB = originalDB
		model.LOG_DB = originalLogDB
		common.IsMasterNode = originalIsMasterNode
		common.SQLitePath = originalSQLitePath
		common.UsingSQLite = originalUsingSQLite
		common.UsingMySQL = originalUsingMySQL
		common.UsingPostgreSQL = originalUsingPostgreSQL
		if testDB != nil {
			sqlDB, err := testDB.DB()
			if err == nil {
				_ = sqlDB.Close()
			}
		}
		if originalDSN == "" {
			require.NoError(t, os.Unsetenv("SQL_DSN"))
		} else {
			require.NoError(t, os.Setenv("SQL_DSN", originalDSN))
		}
	})

	common.IsMasterNode = false
	common.SQLitePath = fmt.Sprintf("file:%s_volc_virtual_portrait_asset?mode=memory&cache=shared", strings.ReplaceAll(t.Name(), "/", "_"))
	common.UsingSQLite = false
	common.UsingMySQL = false
	common.UsingPostgreSQL = false
	require.NoError(t, os.Setenv("SQL_DSN", "local"))
	require.NoError(t, model.InitDB())
	testDB = model.DB
	require.NoError(t, model.DB.AutoMigrate(&model.PortraitAssetJob{}, &model.VirtualPortraitAsset{}))
}

func setTestEnv(t *testing.T, key string, value string) {
	originalValue, hadValue := os.LookupEnv(key)
	require.NoError(t, os.Setenv(key, value))
	t.Cleanup(func() {
		if hadValue {
			require.NoError(t, os.Setenv(key, originalValue))
			return
		}
		require.NoError(t, os.Unsetenv(key))
	})
}
