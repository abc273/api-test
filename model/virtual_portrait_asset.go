package model

import (
	"strings"

	"github.com/QuantumNous/new-api/common"
	"gorm.io/gorm"
)

type AdminVirtualPortraitAssetQuery struct {
	UserID         int
	ExternalUserID string
	Status         string
	FolderID       int
	FilterFolderID bool
	Keyword        string
	AssetID        string
	CreatedFrom    int64
	CreatedTo      int64
}

const (
	VirtualPortraitAssetGroupStatusCreating = "creating"
	VirtualPortraitAssetGroupStatusActive   = "active"
	VirtualPortraitAssetGroupStatusFailed   = "failed"
)

const (
	VirtualPortraitAssetStatusProcessing = "processing"
	VirtualPortraitAssetStatusActive     = "active"
	VirtualPortraitAssetStatusFailed     = "failed"
)

type VirtualPortraitAssetGroup struct {
	Id             int    `json:"id"`
	UserId         int    `json:"user_id" gorm:"uniqueIndex:idx_virtual_portrait_asset_groups_user_external"`
	ExternalUserID string `json:"external_user_id" gorm:"type:varchar(128);default:'';uniqueIndex:idx_virtual_portrait_asset_groups_user_external;index"`
	Name           string `json:"name" gorm:"type:varchar(255)"`
	Description    string `json:"description" gorm:"type:text"`
	ProjectName    string `json:"project_name" gorm:"type:varchar(128);index"`
	VolcGroupID    string `json:"volc_group_id" gorm:"type:varchar(128);index"`
	Status         string `json:"status" gorm:"type:varchar(32);index"`
	ErrorMessage   string `json:"error_message" gorm:"type:text"`
	CreatedTime    int64  `json:"created_time" gorm:"bigint"`
	UpdatedTime    int64  `json:"updated_time" gorm:"bigint"`
}

type VirtualPortraitAsset struct {
	Id             int            `json:"id"`
	UserId         int            `json:"user_id" gorm:"index"`
	ExternalUserID string         `json:"external_user_id" gorm:"type:varchar(128);default:'';index"`
	FolderID       int            `json:"folder_id" gorm:"index"`
	GroupId        int            `json:"group_id" gorm:"index"`
	Name           string         `json:"name" gorm:"type:varchar(255)"`
	AssetType      string         `json:"asset_type" gorm:"type:varchar(32)"`
	SourceURL      string         `json:"source_url" gorm:"type:text"`
	PreviewURL     string         `json:"preview_url" gorm:"type:text"`
	ProjectName    string         `json:"project_name" gorm:"type:varchar(128);index"`
	VolcGroupID    string         `json:"volc_group_id" gorm:"type:varchar(128);index"`
	VolcAssetID    string         `json:"volc_asset_id" gorm:"type:varchar(128);index"`
	Status         string         `json:"status" gorm:"type:varchar(32);index"`
	VolcStatus     string         `json:"volc_status" gorm:"type:varchar(64)"`
	ErrorMessage   string         `json:"error_message" gorm:"type:text"`
	CreatedTime    int64          `json:"created_time" gorm:"bigint"`
	UpdatedTime    int64          `json:"updated_time" gorm:"bigint"`
	ReadyTime      int64          `json:"ready_time" gorm:"bigint"`
	DeletedAt      gorm.DeletedAt `json:"-" gorm:"index"`
}

func NormalizeVirtualPortraitAssetGroupStatus(status string) string {
	switch strings.ToLower(strings.TrimSpace(status)) {
	case VirtualPortraitAssetGroupStatusActive:
		return VirtualPortraitAssetGroupStatusActive
	case VirtualPortraitAssetGroupStatusFailed:
		return VirtualPortraitAssetGroupStatusFailed
	default:
		return VirtualPortraitAssetGroupStatusCreating
	}
}

func NormalizeVirtualPortraitAssetStatus(status string) string {
	switch strings.ToLower(strings.TrimSpace(status)) {
	case "active":
		return VirtualPortraitAssetStatusActive
	case "failed":
		return VirtualPortraitAssetStatusFailed
	default:
		return VirtualPortraitAssetStatusProcessing
	}
}

func GetUserVirtualPortraitAssetGroup(userId int) (*VirtualPortraitAssetGroup, error) {
	return GetUserVirtualPortraitAssetGroupForExternalUser(userId, "")
}

func GetUserVirtualPortraitAssetGroupForExternalUser(userId int, externalUserID string) (*VirtualPortraitAssetGroup, error) {
	var group VirtualPortraitAssetGroup
	err := externalUserIDQuery(DB.Where("user_id = ?", userId), externalUserID).First(&group).Error
	return &group, err
}

func SaveVirtualPortraitAssetGroup(group *VirtualPortraitAssetGroup) error {
	if group == nil {
		return gorm.ErrInvalidData
	}
	group.Status = NormalizeVirtualPortraitAssetGroupStatus(group.Status)
	group.ExternalUserID = NormalizeExternalUserID(group.ExternalUserID)
	group.Name = strings.TrimSpace(group.Name)
	group.Description = strings.TrimSpace(group.Description)
	group.ProjectName = strings.TrimSpace(group.ProjectName)
	group.VolcGroupID = strings.TrimSpace(group.VolcGroupID)
	group.ErrorMessage = strings.TrimSpace(group.ErrorMessage)
	now := common.GetTimestamp()
	if group.CreatedTime == 0 {
		group.CreatedTime = now
	}
	group.UpdatedTime = now
	return DB.Save(group).Error
}

func CreateUserVirtualPortraitAsset(
	userId int,
	group *VirtualPortraitAssetGroup,
	name string,
	assetType string,
	sourceURL string,
	volcAssetID string,
	folderID int,
) (*VirtualPortraitAsset, error) {
	if group == nil {
		return nil, gorm.ErrInvalidData
	}
	if folderID > 0 {
		if _, err := GetUserPortraitAssetFolderByIDForExternalUser(userId, folderID, PortraitAssetFolderKindVirtual, group.ExternalUserID, true); err != nil {
			return nil, err
		}
	}
	name = strings.TrimSpace(name)
	if name == "" {
		name = "虚拟人像素材"
	}
	now := common.GetTimestamp()
	asset := &VirtualPortraitAsset{
		UserId:         userId,
		ExternalUserID: NormalizeExternalUserID(group.ExternalUserID),
		FolderID:       folderID,
		GroupId:        group.Id,
		Name:           name,
		AssetType:      strings.TrimSpace(assetType),
		SourceURL:      strings.TrimSpace(sourceURL),
		PreviewURL:     strings.TrimSpace(sourceURL),
		ProjectName:    strings.TrimSpace(group.ProjectName),
		VolcGroupID:    strings.TrimSpace(group.VolcGroupID),
		VolcAssetID:    NormalizePortraitAssetID(volcAssetID),
		Status:         VirtualPortraitAssetStatusProcessing,
		VolcStatus:     "Processing",
		CreatedTime:    now,
		UpdatedTime:    now,
		ErrorMessage:   "",
	}
	return asset, DB.Create(asset).Error
}

func GetUserVirtualPortraitAssets(userId int, startIdx int, num int) ([]*VirtualPortraitAsset, error) {
	return GetUserVirtualPortraitAssetsForExternalUser(userId, "", false, startIdx, num)
}

func GetUserVirtualPortraitAssetsForExternalUser(userId int, externalUserID string, filterExternalUserID bool, startIdx int, num int) ([]*VirtualPortraitAsset, error) {
	return GetUserVirtualPortraitAssetsWithFolderForExternalUser(userId, externalUserID, filterExternalUserID, 0, false, startIdx, num)
}

func GetUserVirtualPortraitAssetsWithFolderForExternalUser(userId int, externalUserID string, filterExternalUserID bool, folderID int, filterFolderID bool, startIdx int, num int) ([]*VirtualPortraitAsset, error) {
	var assets []*VirtualPortraitAsset
	query := DB.Where("user_id = ?", userId)
	if filterExternalUserID {
		query = externalUserIDQuery(query, externalUserID)
	}
	if filterFolderID {
		query = portraitAssetFolderIDQuery(query, folderID)
	}
	err := query.
		Order("id desc").
		Limit(num).
		Offset(startIdx).
		Find(&assets).Error
	return assets, err
}

func CountUserVirtualPortraitAssets(userId int) (int64, error) {
	return CountUserVirtualPortraitAssetsForExternalUser(userId, "", false)
}

func CountUserVirtualPortraitAssetsForExternalUser(userId int, externalUserID string, filterExternalUserID bool) (int64, error) {
	return CountUserVirtualPortraitAssetsWithFolderForExternalUser(userId, externalUserID, filterExternalUserID, 0, false)
}

func CountUserVirtualPortraitAssetsWithFolderForExternalUser(userId int, externalUserID string, filterExternalUserID bool, folderID int, filterFolderID bool) (int64, error) {
	var count int64
	query := DB.Model(&VirtualPortraitAsset{}).Where("user_id = ?", userId)
	if filterExternalUserID {
		query = externalUserIDQuery(query, externalUserID)
	}
	if filterFolderID {
		query = portraitAssetFolderIDQuery(query, folderID)
	}
	err := query.Count(&count).Error
	return count, err
}

func GetUserVirtualPortraitAssetByID(userId int, id int) (*VirtualPortraitAsset, error) {
	return GetUserVirtualPortraitAssetByIDForExternalUser(userId, id, "", false)
}

func GetUserVirtualPortraitAssetByIDForExternalUser(userId int, id int, externalUserID string, filterExternalUserID bool) (*VirtualPortraitAsset, error) {
	var asset VirtualPortraitAsset
	query := DB.Where("user_id = ? and id = ?", userId, id)
	if filterExternalUserID {
		query = externalUserIDQuery(query, externalUserID)
	}
	err := query.First(&asset).Error
	return &asset, err
}

func GetActiveUserVirtualPortraitAssetByAssetID(userId int, assetID string) (*VirtualPortraitAsset, error) {
	return GetActiveUserVirtualPortraitAssetByAssetIDForExternalUser(userId, assetID, "")
}

func GetUserVirtualPortraitAssetByAssetIDForExternalUser(userId int, assetID string, externalUserID string) (*VirtualPortraitAsset, error) {
	assetID = NormalizePortraitAssetID(assetID)
	if assetID == "" {
		return nil, gorm.ErrRecordNotFound
	}
	var asset VirtualPortraitAsset
	err := externalUserIDQuery(DB.Where("user_id = ? and volc_asset_id = ?", userId, assetID), externalUserID).First(&asset).Error
	return &asset, err
}

func GetActiveUserVirtualPortraitAssetByAssetIDForExternalUser(userId int, assetID string, externalUserID string) (*VirtualPortraitAsset, error) {
	assetID = NormalizePortraitAssetID(assetID)
	if assetID == "" {
		return nil, gorm.ErrRecordNotFound
	}
	var asset VirtualPortraitAsset
	err := externalUserIDQuery(DB.Where("user_id = ? and volc_asset_id = ? and status = ?", userId, assetID, VirtualPortraitAssetStatusActive), externalUserID).First(&asset).Error
	return &asset, err
}

func GetUserVirtualPortraitAssetByAssetIDIgnoringExternalUser(userId int, assetID string) (*VirtualPortraitAsset, error) {
	assetID = NormalizePortraitAssetID(assetID)
	if assetID == "" {
		return nil, gorm.ErrRecordNotFound
	}
	var asset VirtualPortraitAsset
	err := DB.Where("user_id = ? and volc_asset_id = ?", userId, assetID).First(&asset).Error
	return &asset, err
}

func GetActiveUserVirtualPortraitAssetByAssetIDIgnoringExternalUser(userId int, assetID string) (*VirtualPortraitAsset, error) {
	assetID = NormalizePortraitAssetID(assetID)
	if assetID == "" {
		return nil, gorm.ErrRecordNotFound
	}
	var asset VirtualPortraitAsset
	err := DB.Where("user_id = ? and volc_asset_id = ? and status = ?", userId, assetID, VirtualPortraitAssetStatusActive).First(&asset).Error
	return &asset, err
}

func GetVirtualPortraitAssetByID(id int) (*VirtualPortraitAsset, error) {
	var asset VirtualPortraitAsset
	err := DB.Where("id = ?", id).First(&asset).Error
	return &asset, err
}

func DeleteUserVirtualPortraitAssetForExternalUser(userId int, id int, externalUserID string, filterExternalUserID bool) error {
	asset, err := GetUserVirtualPortraitAssetByIDForExternalUser(userId, id, externalUserID, filterExternalUserID)
	if err != nil {
		return err
	}
	return DB.Delete(asset).Error
}

func GetAdminVirtualPortraitAssets(queryParams AdminVirtualPortraitAssetQuery, startIdx int, num int) ([]*VirtualPortraitAsset, error) {
	var assets []*VirtualPortraitAsset
	query := buildAdminVirtualPortraitAssetQuery(queryParams)
	err := query.Order("id desc").Limit(num).Offset(startIdx).Find(&assets).Error
	return assets, err
}

func CountAdminVirtualPortraitAssets(queryParams AdminVirtualPortraitAssetQuery) (int64, error) {
	var count int64
	query := buildAdminVirtualPortraitAssetQuery(queryParams).Model(&VirtualPortraitAsset{})
	err := query.Count(&count).Error
	return count, err
}

func DeleteVirtualPortraitAssetByID(id int) error {
	asset, err := GetVirtualPortraitAssetByID(id)
	if err != nil {
		return err
	}
	return DB.Delete(asset).Error
}

func buildAdminVirtualPortraitAssetQuery(queryParams AdminVirtualPortraitAssetQuery) *gorm.DB {
	query := DB
	if queryParams.UserID > 0 {
		query = query.Where("user_id = ?", queryParams.UserID)
	}
	if NormalizeExternalUserID(queryParams.ExternalUserID) != "" {
		query = externalUserIDQuery(query, queryParams.ExternalUserID)
	}
	if status := strings.TrimSpace(queryParams.Status); status != "" {
		query = query.Where("status = ?", status)
	}
	if queryParams.FilterFolderID {
		query = portraitAssetFolderIDQuery(query, queryParams.FolderID)
	}
	if assetID := NormalizePortraitAssetID(queryParams.AssetID); assetID != "" {
		query = query.Where("volc_asset_id = ?", assetID)
	}
	if keyword := strings.TrimSpace(queryParams.Keyword); keyword != "" {
		like := "%" + keyword + "%"
		query = query.Where("name LIKE ? OR volc_asset_id LIKE ? OR volc_group_id LIKE ?", like, like, like)
	}
	if queryParams.CreatedFrom > 0 {
		query = query.Where("created_time >= ?", queryParams.CreatedFrom)
	}
	if queryParams.CreatedTo > 0 {
		query = query.Where("created_time <= ?", queryParams.CreatedTo)
	}
	return query
}

func SaveVirtualPortraitAsset(asset *VirtualPortraitAsset) error {
	if asset == nil {
		return gorm.ErrInvalidData
	}
	asset.Name = strings.TrimSpace(asset.Name)
	asset.ExternalUserID = NormalizeExternalUserID(asset.ExternalUserID)
	asset.AssetType = strings.TrimSpace(asset.AssetType)
	asset.SourceURL = strings.TrimSpace(asset.SourceURL)
	asset.PreviewURL = strings.TrimSpace(asset.PreviewURL)
	asset.ProjectName = strings.TrimSpace(asset.ProjectName)
	asset.VolcGroupID = strings.TrimSpace(asset.VolcGroupID)
	asset.VolcAssetID = NormalizePortraitAssetID(asset.VolcAssetID)
	asset.Status = NormalizeVirtualPortraitAssetStatus(asset.Status)
	asset.VolcStatus = strings.TrimSpace(asset.VolcStatus)
	asset.ErrorMessage = strings.TrimSpace(asset.ErrorMessage)
	asset.UpdatedTime = common.GetTimestamp()
	if asset.Status == VirtualPortraitAssetStatusActive && asset.ReadyTime == 0 {
		asset.ReadyTime = asset.UpdatedTime
	}
	return DB.Save(asset).Error
}
