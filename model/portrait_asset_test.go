package model

import (
	"fmt"
	"os"
	"strings"
	"testing"

	"github.com/QuantumNous/new-api/common"
	"github.com/stretchr/testify/require"
	"gorm.io/gorm"
)

func TestUpdateOfficialPortraitAssetStatusPreservesOriginalAssetURL(t *testing.T) {
	originalDB := DB
	originalLogDB := LOG_DB
	originalIsMasterNode := common.IsMasterNode
	originalSQLitePath := common.SQLitePath
	originalUsingSQLite := common.UsingSQLite
	originalUsingMySQL := common.UsingMySQL
	originalUsingPostgreSQL := common.UsingPostgreSQL
	originalDSN := os.Getenv("SQL_DSN")
	var testDB *gorm.DB
	t.Cleanup(func() {
		DB = originalDB
		LOG_DB = originalLogDB
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
	common.SQLitePath = fmt.Sprintf("file:%s_portrait_asset?mode=memory&cache=shared", strings.ReplaceAll(t.Name(), "/", "_"))
	common.UsingSQLite = false
	common.UsingMySQL = false
	common.UsingPostgreSQL = false
	require.NoError(t, os.Setenv("SQL_DSN", "local"))
	require.NoError(t, InitDB())
	testDB = DB
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

func TestValidateUserOwnedPortraitAssetByAssetIDSupportsVirtualAssets(t *testing.T) {
	originalDB := DB
	originalLogDB := LOG_DB
	originalIsMasterNode := common.IsMasterNode
	originalSQLitePath := common.SQLitePath
	originalUsingSQLite := common.UsingSQLite
	originalUsingMySQL := common.UsingMySQL
	originalUsingPostgreSQL := common.UsingPostgreSQL
	originalDSN := os.Getenv("SQL_DSN")
	var testDB *gorm.DB
	t.Cleanup(func() {
		DB = originalDB
		LOG_DB = originalLogDB
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
	common.SQLitePath = fmt.Sprintf("file:%s_virtual_portrait_asset?mode=memory&cache=shared", strings.ReplaceAll(t.Name(), "/", "_"))
	common.UsingSQLite = false
	common.UsingMySQL = false
	common.UsingPostgreSQL = false
	require.NoError(t, os.Setenv("SQL_DSN", "local"))
	require.NoError(t, InitDB())
	testDB = DB
	require.NoError(t, DB.AutoMigrate(&PortraitAssetJob{}, &VirtualPortraitAsset{}))

	asset := &VirtualPortraitAsset{
		UserId:      42,
		Name:        "virtual-ref",
		VolcAssetID: "asset-20260610162557-sgp6f",
		Status:      VirtualPortraitAssetStatusActive,
		CreatedTime: 1,
		UpdatedTime: 1,
	}
	require.NoError(t, DB.Create(asset).Error)

	require.NoError(t, ValidateUserOwnedPortraitAssetByAssetID(42, "asset://asset-20260610162557-sgp6f"))
	require.Error(t, ValidateUserOwnedPortraitAssetByAssetID(7, "asset://asset-20260610162557-sgp6f"))
}

func TestOfficialPortraitAssetActiveJobScopedByExternalUserID(t *testing.T) {
	setupPortraitAssetModelTestDB(t, &PortraitAssetJob{})

	jobA, err := CreateOfficialPortraitAssetJob(42, "asset-a", "default", "", "customer-user-a", 0)
	require.NoError(t, err)
	require.Equal(t, "customer-user-a", jobA.ExternalUserID)

	jobB, err := CreateOfficialPortraitAssetJob(42, "asset-b", "default", "", "customer-user-b", 0)
	require.NoError(t, err)
	require.Equal(t, "customer-user-b", jobB.ExternalUserID)

	_, err = CreateOfficialPortraitAssetJob(42, "asset-a-2", "default", "", "customer-user-a", 0)
	require.ErrorIs(t, err, ErrPortraitAssetActiveJob)

	legacyJob, err := CreateOfficialPortraitAssetJob(42, "legacy", "default", "", "", 0)
	require.NoError(t, err)
	require.Empty(t, legacyJob.ExternalUserID)

	_, err = CreateOfficialPortraitAssetJob(42, "legacy-2", "default", "", "", 0)
	require.ErrorIs(t, err, ErrPortraitAssetActiveJob)
}

func TestValidateUserOwnedPortraitAssetByAssetIDScopesExternalUserID(t *testing.T) {
	setupPortraitAssetModelTestDB(t, &PortraitAssetJob{}, &VirtualPortraitAsset{})

	require.NoError(t, DB.Create(&PortraitAssetJob{
		UserId:         42,
		ExternalUserID: "customer-user-a",
		Source:         PortraitAssetSourceOfficial,
		Status:         PortraitAssetStatusReady,
		AssetID:        "asset-official-a",
		CreatedTime:    1,
		UpdatedTime:    1,
	}).Error)
	require.NoError(t, DB.Create(&PortraitAssetJob{
		UserId:      42,
		Source:      PortraitAssetSourceOfficial,
		Status:      PortraitAssetStatusReady,
		AssetID:     "asset-legacy",
		CreatedTime: 1,
		UpdatedTime: 1,
	}).Error)
	require.NoError(t, DB.Create(&VirtualPortraitAsset{
		UserId:         42,
		ExternalUserID: "customer-user-a",
		Name:           "virtual-a",
		VolcAssetID:    "asset-virtual-a",
		Status:         VirtualPortraitAssetStatusActive,
		CreatedTime:    1,
		UpdatedTime:    1,
	}).Error)
	require.NoError(t, DB.Create(&VirtualPortraitAsset{
		UserId:         42,
		ExternalUserID: "customer-user-b",
		Name:           "virtual-b",
		VolcAssetID:    "asset-virtual-b",
		Status:         VirtualPortraitAssetStatusActive,
		CreatedTime:    1,
		UpdatedTime:    1,
	}).Error)

	require.NoError(t, ValidateUserOwnedPortraitAssetByAssetIDForExternalUser(42, "asset://asset-official-a", "customer-user-a"))
	require.Error(t, ValidateUserOwnedPortraitAssetByAssetIDForExternalUser(42, "asset://asset-official-a", "customer-user-b"))
	require.NoError(t, ValidateUserOwnedPortraitAssetByAssetID(42, "asset://asset-official-a"))

	require.NoError(t, ValidateUserOwnedPortraitAssetByAssetID(42, "asset://asset-legacy"))
	require.Error(t, ValidateUserOwnedPortraitAssetByAssetIDForExternalUser(42, "asset://asset-legacy", "customer-user-a"))

	require.NoError(t, ValidateUserOwnedPortraitAssetByAssetIDForExternalUser(42, "asset://asset-virtual-a", "customer-user-a"))
	require.Error(t, ValidateUserOwnedPortraitAssetByAssetIDForExternalUser(42, "asset://asset-virtual-a", "customer-user-b"))
	require.NoError(t, ValidateUserOwnedPortraitAssetByAssetID(42, "asset://asset-virtual-a"))
}

func TestVirtualPortraitAssetGroupScopedByExternalUserID(t *testing.T) {
	setupPortraitAssetModelTestDB(t, &VirtualPortraitAssetGroup{})

	require.NoError(t, SaveVirtualPortraitAssetGroup(&VirtualPortraitAssetGroup{
		UserId:         42,
		ExternalUserID: "customer-user-a",
		Name:           "group-a",
		ProjectName:    "default",
		VolcGroupID:    "group-a",
		Status:         VirtualPortraitAssetGroupStatusActive,
	}))
	require.NoError(t, SaveVirtualPortraitAssetGroup(&VirtualPortraitAssetGroup{
		UserId:         42,
		ExternalUserID: "customer-user-b",
		Name:           "group-b",
		ProjectName:    "default",
		VolcGroupID:    "group-b",
		Status:         VirtualPortraitAssetGroupStatusActive,
	}))

	groupA, err := GetUserVirtualPortraitAssetGroupForExternalUser(42, "customer-user-a")
	require.NoError(t, err)
	require.Equal(t, "group-a", groupA.VolcGroupID)

	groupB, err := GetUserVirtualPortraitAssetGroupForExternalUser(42, "customer-user-b")
	require.NoError(t, err)
	require.Equal(t, "group-b", groupB.VolcGroupID)

	_, err = GetUserVirtualPortraitAssetGroupForExternalUser(42, "")
	require.ErrorIs(t, err, gorm.ErrRecordNotFound)
}

func TestPortraitAssetFoldersScopedByKindAndExternalUserID(t *testing.T) {
	setupPortraitAssetModelTestDB(t, &PortraitAssetFolder{})

	officialA, err := CreateUserPortraitAssetFolder(42, PortraitAssetFolderKindOfficial, "分镜图", "customer-user-a", 0)
	require.NoError(t, err)
	require.Equal(t, PortraitAssetFolderKindOfficial, officialA.AssetKind)
	require.Equal(t, "customer-user-a", officialA.ExternalUserID)

	virtualA, err := CreateUserPortraitAssetFolder(42, PortraitAssetFolderKindVirtual, "分镜图", "customer-user-a", 0)
	require.NoError(t, err)
	require.Equal(t, PortraitAssetFolderKindVirtual, virtualA.AssetKind)

	_, err = CreateUserPortraitAssetFolder(42, PortraitAssetFolderKindOfficial, "分镜图", "customer-user-a", 0)
	require.ErrorIs(t, err, ErrPortraitAssetFolderExists)

	officialB, err := CreateUserPortraitAssetFolder(42, PortraitAssetFolderKindOfficial, "分镜图", "customer-user-b", 0)
	require.NoError(t, err)
	require.Equal(t, "customer-user-b", officialB.ExternalUserID)

	folders, err := ListUserPortraitAssetFolders(42, PortraitAssetFolderKindOfficial, "customer-user-a")
	require.NoError(t, err)
	require.Len(t, folders, 1)
	require.Equal(t, officialA.Id, folders[0].Id)
}

func TestPortraitAssetFolderFilteringAndDeleteUngroupsAssets(t *testing.T) {
	setupPortraitAssetModelTestDB(t, &PortraitAssetFolder{}, &PortraitAssetJob{})

	folder, err := CreateUserPortraitAssetFolder(42, PortraitAssetFolderKindOfficial, "项目A", "", 0)
	require.NoError(t, err)
	require.NoError(t, DB.Create(&PortraitAssetJob{
		UserId:      42,
		FolderID:    folder.Id,
		Source:      PortraitAssetSourceOfficial,
		Status:      PortraitAssetStatusReady,
		AssetID:     "asset-folder",
		CreatedTime: 1,
		UpdatedTime: 1,
	}).Error)
	require.NoError(t, DB.Create(&PortraitAssetJob{
		UserId:      42,
		Source:      PortraitAssetSourceOfficial,
		Status:      PortraitAssetStatusReady,
		AssetID:     "asset-ungrouped",
		CreatedTime: 1,
		UpdatedTime: 1,
	}).Error)
	require.NoError(t, DB.Create(&PortraitAssetJob{
		UserId:      42,
		Source:      PortraitAssetSourceOfficial,
		Status:      PortraitAssetStatusReady,
		AssetID:     "asset-null-folder",
		CreatedTime: 1,
		UpdatedTime: 1,
	}).Error)
	require.NoError(t, DB.Model(&PortraitAssetJob{}).
		Where("asset_id = ?", "asset-null-folder").
		Update("folder_id", gorm.Expr("NULL")).Error)

	allJobs, err := GetUserOfficialPortraitAssetJobsWithFolder(42, "", 0, false, 0, 10)
	require.NoError(t, err)
	require.Len(t, allJobs, 3)

	ungroupedJobs, err := GetUserOfficialPortraitAssetJobsWithFolder(42, "", 0, true, 0, 10)
	require.NoError(t, err)
	require.Len(t, ungroupedJobs, 2)
	require.ElementsMatch(t, []string{"asset-ungrouped", "asset-null-folder"}, portraitAssetJobIDs(ungroupedJobs))
	ungroupedCount, err := CountUserOfficialPortraitAssetJobsWithFolder(42, "", 0, true)
	require.NoError(t, err)
	require.EqualValues(t, 2, ungroupedCount)

	folderJobs, err := GetUserOfficialPortraitAssetJobsWithFolder(42, "", folder.Id, true, 0, 10)
	require.NoError(t, err)
	require.Len(t, folderJobs, 1)
	require.Equal(t, "asset-folder", folderJobs[0].AssetID)

	require.NoError(t, DeleteUserPortraitAssetFolder(42, folder.Id))
	folderJobs, err = GetUserOfficialPortraitAssetJobsWithFolder(42, "", folder.Id, true, 0, 10)
	require.NoError(t, err)
	require.Empty(t, folderJobs)
	ungroupedJobs, err = GetUserOfficialPortraitAssetJobsWithFolder(42, "", 0, true, 0, 10)
	require.NoError(t, err)
	require.Len(t, ungroupedJobs, 3)
}

func TestPortraitAssetFolderMoveScopesExternalUserID(t *testing.T) {
	setupPortraitAssetModelTestDB(t, &PortraitAssetFolder{}, &VirtualPortraitAsset{})

	folderA, err := CreateUserPortraitAssetFolder(42, PortraitAssetFolderKindVirtual, "用户A", "customer-user-a", 0)
	require.NoError(t, err)
	folderB, err := CreateUserPortraitAssetFolder(42, PortraitAssetFolderKindVirtual, "用户B", "customer-user-b", 0)
	require.NoError(t, err)

	assetA := &VirtualPortraitAsset{
		UserId:         42,
		ExternalUserID: "customer-user-a",
		Name:           "asset-a",
		VolcAssetID:    "asset-a",
		Status:         VirtualPortraitAssetStatusActive,
		CreatedTime:    1,
		UpdatedTime:    1,
	}
	assetB := &VirtualPortraitAsset{
		UserId:         42,
		ExternalUserID: "customer-user-b",
		Name:           "asset-b",
		FolderID:       folderB.Id,
		VolcAssetID:    "asset-b",
		Status:         VirtualPortraitAssetStatusActive,
		CreatedTime:    1,
		UpdatedTime:    1,
	}
	require.NoError(t, DB.Create(assetA).Error)
	require.NoError(t, DB.Create(assetB).Error)

	require.NoError(t, MoveUserPortraitAssetsToFolder(42, PortraitAssetFolderKindVirtual, []int{assetA.Id}, folderA.Id, ""))
	movedA, err := GetUserVirtualPortraitAssetByID(42, assetA.Id)
	require.NoError(t, err)
	require.Equal(t, folderA.Id, movedA.FolderID)

	require.ErrorIs(t, MoveUserPortraitAssetsToFolder(42, PortraitAssetFolderKindVirtual, []int{assetB.Id}, folderA.Id, ""), gorm.ErrRecordNotFound)
	movedB, err := GetUserVirtualPortraitAssetByID(42, assetB.Id)
	require.NoError(t, err)
	require.Equal(t, folderB.Id, movedB.FolderID)

	require.NoError(t, MoveUserPortraitAssetsToFolder(42, PortraitAssetFolderKindVirtual, []int{assetA.Id}, 0, "customer-user-a"))
	movedA, err = GetUserVirtualPortraitAssetByID(42, assetA.Id)
	require.NoError(t, err)
	require.Zero(t, movedA.FolderID)
}

func TestVirtualPortraitAssetFolderFilteringTreatsNullAsUngrouped(t *testing.T) {
	setupPortraitAssetModelTestDB(t, &PortraitAssetFolder{}, &VirtualPortraitAsset{})

	folder, err := CreateUserPortraitAssetFolder(42, PortraitAssetFolderKindVirtual, "项目A", "", 0)
	require.NoError(t, err)
	require.NoError(t, DB.Create(&VirtualPortraitAsset{
		UserId:      42,
		FolderID:    folder.Id,
		Name:        "folder-asset",
		VolcAssetID: "asset-folder",
		Status:      VirtualPortraitAssetStatusActive,
		CreatedTime: 1,
		UpdatedTime: 1,
	}).Error)
	require.NoError(t, DB.Create(&VirtualPortraitAsset{
		UserId:      42,
		Name:        "ungrouped-asset",
		VolcAssetID: "asset-ungrouped",
		Status:      VirtualPortraitAssetStatusActive,
		CreatedTime: 1,
		UpdatedTime: 1,
	}).Error)
	require.NoError(t, DB.Create(&VirtualPortraitAsset{
		UserId:      42,
		Name:        "null-folder-asset",
		VolcAssetID: "asset-null-folder",
		Status:      VirtualPortraitAssetStatusActive,
		CreatedTime: 1,
		UpdatedTime: 1,
	}).Error)
	require.NoError(t, DB.Model(&VirtualPortraitAsset{}).
		Where("volc_asset_id = ?", "asset-null-folder").
		Update("folder_id", gorm.Expr("NULL")).Error)

	allAssets, err := GetUserVirtualPortraitAssetsWithFolderForExternalUser(42, "", false, 0, false, 0, 10)
	require.NoError(t, err)
	require.Len(t, allAssets, 3)

	ungroupedAssets, err := GetUserVirtualPortraitAssetsWithFolderForExternalUser(42, "", false, 0, true, 0, 10)
	require.NoError(t, err)
	require.Len(t, ungroupedAssets, 2)
	require.ElementsMatch(t, []string{"asset-ungrouped", "asset-null-folder"}, virtualPortraitAssetIDs(ungroupedAssets))
	ungroupedCount, err := CountUserVirtualPortraitAssetsWithFolderForExternalUser(42, "", false, 0, true)
	require.NoError(t, err)
	require.EqualValues(t, 2, ungroupedCount)

	folderAssets, err := GetUserVirtualPortraitAssetsWithFolderForExternalUser(42, "", false, folder.Id, true, 0, 10)
	require.NoError(t, err)
	require.Len(t, folderAssets, 1)
	require.Equal(t, "asset-folder", folderAssets[0].VolcAssetID)
}

func TestDeleteOfficialPortraitAssetJobSoftDeletesAsset(t *testing.T) {
	setupPortraitAssetModelTestDB(t, &PortraitAssetJob{})

	job := &PortraitAssetJob{
		UserId:      42,
		Source:      PortraitAssetSourceOfficial,
		Status:      PortraitAssetStatusReady,
		AssetID:     "asset-delete-official",
		CreatedTime: 1,
		UpdatedTime: 1,
	}
	require.NoError(t, DB.Create(job).Error)

	require.NoError(t, DeleteUserOfficialPortraitAssetJobForExternalUser(42, job.Id, "", false))

	jobs, err := GetUserOfficialPortraitAssetJobsWithFolder(42, "", 0, false, 0, 10)
	require.NoError(t, err)
	require.Empty(t, jobs)
	_, err = GetReadyUserPortraitAssetByAssetID(42, "asset-delete-official")
	require.ErrorIs(t, err, gorm.ErrRecordNotFound)

	var deleted PortraitAssetJob
	require.NoError(t, DB.Unscoped().Where("id = ?", job.Id).First(&deleted).Error)
	require.True(t, deleted.DeletedAt.Valid)
}

func TestDeleteVirtualPortraitAssetSoftDeletesAsset(t *testing.T) {
	setupPortraitAssetModelTestDB(t, &VirtualPortraitAsset{})

	asset := &VirtualPortraitAsset{
		UserId:      42,
		Name:        "delete-virtual",
		VolcAssetID: "asset-delete-virtual",
		Status:      VirtualPortraitAssetStatusActive,
		CreatedTime: 1,
		UpdatedTime: 1,
	}
	require.NoError(t, DB.Create(asset).Error)

	require.NoError(t, DeleteUserVirtualPortraitAssetForExternalUser(42, asset.Id, "", false))

	assets, err := GetUserVirtualPortraitAssetsWithFolderForExternalUser(42, "", false, 0, false, 0, 10)
	require.NoError(t, err)
	require.Empty(t, assets)
	_, err = GetActiveUserVirtualPortraitAssetByAssetID(42, "asset-delete-virtual")
	require.ErrorIs(t, err, gorm.ErrRecordNotFound)

	var deleted VirtualPortraitAsset
	require.NoError(t, DB.Unscoped().Where("id = ?", asset.Id).First(&deleted).Error)
	require.True(t, deleted.DeletedAt.Valid)
}

func portraitAssetJobIDs(jobs []*PortraitAssetJob) []string {
	ids := make([]string, 0, len(jobs))
	for _, job := range jobs {
		ids = append(ids, job.AssetID)
	}
	return ids
}

func virtualPortraitAssetIDs(assets []*VirtualPortraitAsset) []string {
	ids := make([]string, 0, len(assets))
	for _, asset := range assets {
		ids = append(ids, asset.VolcAssetID)
	}
	return ids
}

func setupPortraitAssetModelTestDB(t *testing.T, models ...interface{}) {
	originalDB := DB
	originalLogDB := LOG_DB
	originalIsMasterNode := common.IsMasterNode
	originalSQLitePath := common.SQLitePath
	originalUsingSQLite := common.UsingSQLite
	originalUsingMySQL := common.UsingMySQL
	originalUsingPostgreSQL := common.UsingPostgreSQL
	originalDSN := os.Getenv("SQL_DSN")
	var testDB *gorm.DB
	t.Cleanup(func() {
		DB = originalDB
		LOG_DB = originalLogDB
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
	common.SQLitePath = fmt.Sprintf("file:%s_external_user_id?mode=memory&cache=shared", strings.ReplaceAll(t.Name(), "/", "_"))
	common.UsingSQLite = false
	common.UsingMySQL = false
	common.UsingPostgreSQL = false
	require.NoError(t, os.Setenv("SQL_DSN", "local"))
	require.NoError(t, InitDB())
	testDB = DB
	require.NoError(t, DB.AutoMigrate(models...))
}
