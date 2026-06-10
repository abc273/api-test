package model

import (
	"fmt"
	"os"
	"strings"
	"testing"

	"github.com/QuantumNous/new-api/common"
	"github.com/stretchr/testify/require"
)

func TestUpdateOfficialPortraitAssetStatusPreservesOriginalAssetURL(t *testing.T) {
	originalIsMasterNode := common.IsMasterNode
	originalSQLitePath := common.SQLitePath
	originalUsingSQLite := common.UsingSQLite
	originalUsingMySQL := common.UsingMySQL
	originalUsingPostgreSQL := common.UsingPostgreSQL
	originalDSN := os.Getenv("SQL_DSN")
	t.Cleanup(func() {
		common.IsMasterNode = originalIsMasterNode
		common.SQLitePath = originalSQLitePath
		common.UsingSQLite = originalUsingSQLite
		common.UsingMySQL = originalUsingMySQL
		common.UsingPostgreSQL = originalUsingPostgreSQL
		if DB != nil {
			sqlDB, err := DB.DB()
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
	common.SQLitePath = fmt.Sprintf("file:%s_portrait_asset?mode=memory&cache=shared", strings.ReplaceAll(t.Name(), "/", "_"))
	common.UsingSQLite = false
	common.UsingMySQL = false
	common.UsingPostgreSQL = false
	require.NoError(t, os.Setenv("SQL_DSN", "local"))
	require.NoError(t, InitDB())
	require.NoError(t, DB.AutoMigrate(&PortraitAssetJob{}))

	job := &PortraitAssetJob{
		AssetURL: "https://example.com/uploads/original.jpg",
	}

	err := UpdateOfficialPortraitAssetStatus(job, "Active", "https://tos.example.com/signed-preview")

	require.NoError(t, err)
	require.Equal(t, "https://example.com/uploads/original.jpg", job.AssetURL)
	require.Equal(t, "https://tos.example.com/signed-preview", job.AssetPreview)
	require.Equal(t, PortraitAssetStatusPendingConfirm, job.Status)
}
