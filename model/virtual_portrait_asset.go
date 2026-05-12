package model

import (
	"strings"

	"github.com/QuantumNous/new-api/common"
	"gorm.io/gorm"
)

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
	Id           int    `json:"id"`
	UserId       int    `json:"user_id" gorm:"uniqueIndex"`
	Name         string `json:"name" gorm:"type:varchar(255)"`
	Description  string `json:"description" gorm:"type:text"`
	ProjectName  string `json:"project_name" gorm:"type:varchar(128);index"`
	VolcGroupID  string `json:"volc_group_id" gorm:"type:varchar(128);index"`
	Status       string `json:"status" gorm:"type:varchar(32);index"`
	ErrorMessage string `json:"error_message" gorm:"type:text"`
	CreatedTime  int64  `json:"created_time" gorm:"bigint"`
	UpdatedTime  int64  `json:"updated_time" gorm:"bigint"`
}

type VirtualPortraitAsset struct {
	Id           int    `json:"id"`
	UserId       int    `json:"user_id" gorm:"index"`
	GroupId      int    `json:"group_id" gorm:"index"`
	Name         string `json:"name" gorm:"type:varchar(255)"`
	AssetType    string `json:"asset_type" gorm:"type:varchar(32)"`
	SourceURL    string `json:"source_url" gorm:"type:text"`
	PreviewURL   string `json:"preview_url" gorm:"type:text"`
	ProjectName  string `json:"project_name" gorm:"type:varchar(128);index"`
	VolcGroupID  string `json:"volc_group_id" gorm:"type:varchar(128);index"`
	VolcAssetID  string `json:"volc_asset_id" gorm:"type:varchar(128);index"`
	Status       string `json:"status" gorm:"type:varchar(32);index"`
	VolcStatus   string `json:"volc_status" gorm:"type:varchar(64)"`
	ErrorMessage string `json:"error_message" gorm:"type:text"`
	CreatedTime  int64  `json:"created_time" gorm:"bigint"`
	UpdatedTime  int64  `json:"updated_time" gorm:"bigint"`
	ReadyTime    int64  `json:"ready_time" gorm:"bigint"`
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
	var group VirtualPortraitAssetGroup
	err := DB.Where("user_id = ?", userId).First(&group).Error
	return &group, err
}

func SaveVirtualPortraitAssetGroup(group *VirtualPortraitAssetGroup) error {
	if group == nil {
		return gorm.ErrInvalidData
	}
	group.Status = NormalizeVirtualPortraitAssetGroupStatus(group.Status)
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
) (*VirtualPortraitAsset, error) {
	if group == nil {
		return nil, gorm.ErrInvalidData
	}
	name = strings.TrimSpace(name)
	if name == "" {
		name = "虚拟人像素材"
	}
	now := common.GetTimestamp()
	asset := &VirtualPortraitAsset{
		UserId:       userId,
		GroupId:      group.Id,
		Name:         name,
		AssetType:    strings.TrimSpace(assetType),
		SourceURL:    strings.TrimSpace(sourceURL),
		PreviewURL:   strings.TrimSpace(sourceURL),
		ProjectName:  strings.TrimSpace(group.ProjectName),
		VolcGroupID:  strings.TrimSpace(group.VolcGroupID),
		VolcAssetID:  NormalizePortraitAssetID(volcAssetID),
		Status:       VirtualPortraitAssetStatusProcessing,
		VolcStatus:   "Processing",
		CreatedTime:  now,
		UpdatedTime:  now,
		ErrorMessage: "",
	}
	return asset, DB.Create(asset).Error
}

func GetUserVirtualPortraitAssets(userId int, startIdx int, num int) ([]*VirtualPortraitAsset, error) {
	var assets []*VirtualPortraitAsset
	err := DB.Where("user_id = ?", userId).
		Order("id desc").
		Limit(num).
		Offset(startIdx).
		Find(&assets).Error
	return assets, err
}

func CountUserVirtualPortraitAssets(userId int) (int64, error) {
	var count int64
	err := DB.Model(&VirtualPortraitAsset{}).Where("user_id = ?", userId).Count(&count).Error
	return count, err
}

func GetUserVirtualPortraitAssetByID(userId int, id int) (*VirtualPortraitAsset, error) {
	var asset VirtualPortraitAsset
	err := DB.Where("user_id = ? and id = ?", userId, id).First(&asset).Error
	return &asset, err
}

func GetVirtualPortraitAssetByID(id int) (*VirtualPortraitAsset, error) {
	var asset VirtualPortraitAsset
	err := DB.Where("id = ?", id).First(&asset).Error
	return &asset, err
}

func SaveVirtualPortraitAsset(asset *VirtualPortraitAsset) error {
	if asset == nil {
		return gorm.ErrInvalidData
	}
	asset.Name = strings.TrimSpace(asset.Name)
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
