package controller

import (
	"fmt"
	"os"
	"testing"
	"time"

	"github.com/QuantumNous/new-api/common"
	"github.com/QuantumNous/new-api/model"
	"github.com/stretchr/testify/require"
	"gorm.io/gorm"
)

func TestNormalizePortraitPreviewCandidateFallsBackWhenSignedURLExpired(t *testing.T) {
	expired := time.Now().UTC().Add(-2 * time.Minute).Format("20060102T150405Z")
	fallback := "https://example.com/uploads/source.jpg"
	candidate := "https://tos.example.com/object?X-Tos-Date=" + expired + "&X-Tos-Expires=60"

	result := normalizePortraitPreviewCandidate(candidate, fallback)

	require.Equal(t, fallback, result)
}

func TestNormalizePortraitPreviewCandidateKeepsValidSignedURL(t *testing.T) {
	future := time.Now().UTC().Add(-30 * time.Second).Format("20060102T150405Z")
	fallback := "https://example.com/uploads/source.jpg"
	candidate := "https://tos.example.com/object?X-Tos-Date=" + future + "&X-Tos-Expires=300"

	result := normalizePortraitPreviewCandidate(candidate, fallback)

	require.Equal(t, candidate, result)
}

func TestBuildOfficialPortraitFinalRedirectURLUsesCustomCallback(t *testing.T) {
	job := &model.PortraitAssetJob{
		Id:          123,
		CallbackURL: "https://client.example.com/portrait/callback?source=scan",
	}

	result := buildOfficialPortraitFinalRedirectURL(job, "10000")

	require.Equal(t, "https://client.example.com/portrait/callback?job_id=123&result_code=10000&source=scan", result)
}

func TestBuildOfficialPortraitFinalRedirectURLFallsBackToConsole(t *testing.T) {
	job := &model.PortraitAssetJob{
		Id:          123,
		CallbackURL: "javascript:alert(1)",
	}

	result := buildOfficialPortraitFinalRedirectURL(job, "10001")

	require.Equal(t, "/portrait-assets-official?job_id=123&result_code=10001", result)
}

func TestMarkOfficialPortraitGroupExpired(t *testing.T) {
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
	common.SQLitePath = fmt.Sprintf("file:%s_group_expired?mode=memory&cache=shared", t.Name())
	common.UsingSQLite = false
	common.UsingMySQL = false
	common.UsingPostgreSQL = false
	require.NoError(t, os.Setenv("SQL_DSN", "local"))
	require.NoError(t, model.InitDB())
	testDB = model.DB
	require.NoError(t, model.DB.AutoMigrate(&model.PortraitAssetJob{}))

	job := &model.PortraitAssetJob{
		UserId:             1,
		Name:               "official-asset",
		Source:             model.PortraitAssetSourceOfficial,
		Status:             model.PortraitAssetStatusValidated,
		ValidateResultCode: "10000",
		ValidateToken:      "token",
		VolcGroupID:        "group-legacy",
		CreatedTime:        1,
		UpdatedTime:        1,
	}
	require.NoError(t, model.DB.Create(job).Error)

	markOfficialPortraitGroupExpired(job)

	refreshed, err := model.GetOfficialPortraitAssetJobByID(job.Id)
	require.NoError(t, err)
	require.Equal(t, model.PortraitAssetStatusExpired, refreshed.Status)
	require.Empty(t, refreshed.VolcGroupID)
	require.Equal(t, "真人认证分组已失效，请重新生成人脸认证链接后再上传素材", refreshed.ErrorMessage)
}
