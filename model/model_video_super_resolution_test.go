package model

import (
	"fmt"
	"strings"
	"testing"

	"github.com/QuantumNous/new-api/common"
	"github.com/glebarez/sqlite"
	"github.com/stretchr/testify/require"
	"gorm.io/gorm"
)

func setupVideoSuperResolutionModelTestDB(t *testing.T) *gorm.DB {
	t.Helper()

	originalDB := DB
	originalLogDB := LOG_DB
	originalUsingSQLite := common.UsingSQLite
	originalUsingMySQL := common.UsingMySQL
	originalUsingPostgreSQL := common.UsingPostgreSQL

	common.UsingSQLite = true
	common.UsingMySQL = false
	common.UsingPostgreSQL = false

	dsn := fmt.Sprintf("file:%s?mode=memory&cache=shared", strings.ReplaceAll(t.Name(), "/", "_"))
	db, err := gorm.Open(sqlite.Open(dsn), &gorm.Config{})
	require.NoError(t, err)
	DB = db
	LOG_DB = db
	require.NoError(t, db.AutoMigrate(&Model{}))

	t.Cleanup(func() {
		DB = originalDB
		LOG_DB = originalLogDB
		common.UsingSQLite = originalUsingSQLite
		common.UsingMySQL = originalUsingMySQL
		common.UsingPostgreSQL = originalUsingPostgreSQL

		sqlDB, err := db.DB()
		if err == nil {
			_ = sqlDB.Close()
		}
	})

	return db
}

func TestIsVideoSuperResolutionModel(t *testing.T) {
	db := setupVideoSuperResolutionModelTestDB(t)

	require.NoError(t, db.Create(&Model{
		ModelName:                   "seedance2.0fast-sr",
		Status:                      1,
		SyncOfficial:                1,
		VideoSuperResolutionEnabled: true,
	}).Error)
	require.NoError(t, db.Create(&Model{
		ModelName:                   "seedance2-normal",
		Status:                      1,
		SyncOfficial:                1,
		VideoSuperResolutionEnabled: false,
	}).Error)

	enabled, err := IsVideoSuperResolutionModel("seedance2.0fast-sr")
	require.NoError(t, err)
	require.True(t, enabled)

	enabled, err = IsVideoSuperResolutionModel("seedance2.0fast")
	require.NoError(t, err)
	require.False(t, enabled)

	enabled, err = IsVideoSuperResolutionModel("sd2.0fast-sr")
	require.NoError(t, err)
	require.False(t, enabled)

	require.NoError(t, db.Create(&Model{
		ModelName:                   "sd2.0fast",
		Status:                      1,
		SyncOfficial:                1,
		VideoSuperResolutionEnabled: true,
	}).Error)

	enabled, err = IsVideoSuperResolutionModel("sd2.0fast")
	require.NoError(t, err)
	require.True(t, enabled)

	enabled, err = IsVideoSuperResolutionModel("seedance2-normal")
	require.NoError(t, err)
	require.False(t, enabled)

	enabled, err = IsVideoSuperResolutionModel("unknown-model")
	require.NoError(t, err)
	require.False(t, enabled)
}
