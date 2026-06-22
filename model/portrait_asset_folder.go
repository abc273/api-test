package model

import (
	"errors"
	"strings"

	"github.com/QuantumNous/new-api/common"
	"gorm.io/gorm"
)

const (
	PortraitAssetFolderKindOfficial = "official"
	PortraitAssetFolderKindVirtual  = "virtual"
	PortraitAssetFolderNameMaxRunes = 128
	PortraitAssetFolderMoveMaxItems = 1000
)

var (
	ErrPortraitAssetFolderInvalidKind = errors.New("invalid portrait asset folder kind")
	ErrPortraitAssetFolderNameEmpty   = errors.New("portrait asset folder name is empty")
	ErrPortraitAssetFolderNameTooLong = errors.New("portrait asset folder name is too long")
	ErrPortraitAssetFolderExists      = errors.New("portrait asset folder already exists")
)

type PortraitAssetFolder struct {
	Id             int            `json:"id"`
	UserId         int            `json:"user_id" gorm:"index"`
	ExternalUserID string         `json:"external_user_id" gorm:"type:varchar(128);default:'';index"`
	AssetKind      string         `json:"asset_kind" gorm:"type:varchar(32);index"`
	Name           string         `json:"name" gorm:"type:varchar(128);index"`
	SortOrder      int            `json:"sort_order" gorm:"index"`
	CreatedTime    int64          `json:"created_time" gorm:"bigint"`
	UpdatedTime    int64          `json:"updated_time" gorm:"bigint"`
	DeletedAt      gorm.DeletedAt `json:"-" gorm:"index"`
}

func NormalizePortraitAssetFolderKind(assetKind string) (string, bool) {
	switch strings.ToLower(strings.TrimSpace(assetKind)) {
	case PortraitAssetFolderKindOfficial:
		return PortraitAssetFolderKindOfficial, true
	case PortraitAssetFolderKindVirtual:
		return PortraitAssetFolderKindVirtual, true
	default:
		return "", false
	}
}

func normalizePortraitAssetFolderName(name string) (string, error) {
	name = strings.TrimSpace(name)
	if name == "" {
		return "", ErrPortraitAssetFolderNameEmpty
	}
	if len([]rune(name)) > PortraitAssetFolderNameMaxRunes {
		return "", ErrPortraitAssetFolderNameTooLong
	}
	return name, nil
}

func portraitAssetFolderQuery(userId int, assetKind string, externalUserID string, filterExternalUserID bool) (*gorm.DB, error) {
	assetKind, ok := NormalizePortraitAssetFolderKind(assetKind)
	if !ok {
		return nil, ErrPortraitAssetFolderInvalidKind
	}
	query := DB.Where("user_id = ? and asset_kind = ?", userId, assetKind)
	if filterExternalUserID {
		query = externalUserIDQuery(query, externalUserID)
	} else if NormalizeExternalUserID(externalUserID) != "" {
		query = externalUserIDQuery(query, externalUserID)
	}
	return query, nil
}

func ListUserPortraitAssetFolders(userId int, assetKind string, externalUserID string) ([]*PortraitAssetFolder, error) {
	query, err := portraitAssetFolderQuery(userId, assetKind, externalUserID, false)
	if err != nil {
		return nil, err
	}
	var folders []*PortraitAssetFolder
	err = query.Order("sort_order asc, id asc").Find(&folders).Error
	return folders, err
}

func CreateUserPortraitAssetFolder(userId int, assetKind string, name string, externalUserID string, sortOrder int) (*PortraitAssetFolder, error) {
	assetKind, ok := NormalizePortraitAssetFolderKind(assetKind)
	if !ok {
		return nil, ErrPortraitAssetFolderInvalidKind
	}
	name, err := normalizePortraitAssetFolderName(name)
	if err != nil {
		return nil, err
	}
	externalUserID = NormalizeExternalUserID(externalUserID)
	if exists, err := hasUserPortraitAssetFolderName(userId, assetKind, name, externalUserID, 0); err != nil {
		return nil, err
	} else if exists {
		return nil, ErrPortraitAssetFolderExists
	}
	now := common.GetTimestamp()
	folder := &PortraitAssetFolder{
		UserId:         userId,
		ExternalUserID: externalUserID,
		AssetKind:      assetKind,
		Name:           name,
		SortOrder:      sortOrder,
		CreatedTime:    now,
		UpdatedTime:    now,
	}
	return folder, DB.Create(folder).Error
}

func GetUserPortraitAssetFolderByID(userId int, id int) (*PortraitAssetFolder, error) {
	var folder PortraitAssetFolder
	err := DB.Where("user_id = ? and id = ?", userId, id).First(&folder).Error
	return &folder, err
}

func GetUserPortraitAssetFolderByIDForExternalUser(userId int, id int, assetKind string, externalUserID string, filterExternalUserID bool) (*PortraitAssetFolder, error) {
	query, err := portraitAssetFolderQuery(userId, assetKind, externalUserID, filterExternalUserID)
	if err != nil {
		return nil, err
	}
	var folder PortraitAssetFolder
	err = query.Where("id = ?", id).First(&folder).Error
	return &folder, err
}

func UpdateUserPortraitAssetFolder(userId int, id int, name string, sortOrder *int) (*PortraitAssetFolder, error) {
	return UpdateUserPortraitAssetFolderForExternalUser(userId, id, name, sortOrder, "", false)
}

func UpdateUserPortraitAssetFolderForExternalUser(userId int, id int, name string, sortOrder *int, externalUserID string, filterExternalUserID bool) (*PortraitAssetFolder, error) {
	folder, err := getUserPortraitAssetFolderByIDWithOptionalExternalUser(userId, id, externalUserID, filterExternalUserID)
	if err != nil {
		return nil, err
	}
	name, err = normalizePortraitAssetFolderName(name)
	if err != nil {
		return nil, err
	}
	if exists, err := hasUserPortraitAssetFolderName(userId, folder.AssetKind, name, folder.ExternalUserID, folder.Id); err != nil {
		return nil, err
	} else if exists {
		return nil, ErrPortraitAssetFolderExists
	}
	folder.Name = name
	if sortOrder != nil {
		folder.SortOrder = *sortOrder
	}
	folder.UpdatedTime = common.GetTimestamp()
	return folder, DB.Save(folder).Error
}

func DeleteUserPortraitAssetFolder(userId int, id int) error {
	return DeleteUserPortraitAssetFolderForExternalUser(userId, id, "", false)
}

func DeleteUserPortraitAssetFolderForExternalUser(userId int, id int, externalUserID string, filterExternalUserID bool) error {
	folder, err := getUserPortraitAssetFolderByIDWithOptionalExternalUser(userId, id, externalUserID, filterExternalUserID)
	if err != nil {
		return err
	}
	return DB.Transaction(func(tx *gorm.DB) error {
		switch folder.AssetKind {
		case PortraitAssetFolderKindOfficial:
			if err := tx.Model(&PortraitAssetJob{}).
				Where("user_id = ? and source = ? and folder_id = ?", userId, PortraitAssetSourceOfficial, id).
				Update("folder_id", 0).Error; err != nil {
				return err
			}
		case PortraitAssetFolderKindVirtual:
			if err := tx.Model(&VirtualPortraitAsset{}).
				Where("user_id = ? and folder_id = ?", userId, id).
				Update("folder_id", 0).Error; err != nil {
				return err
			}
		default:
			return ErrPortraitAssetFolderInvalidKind
		}
		return tx.Delete(folder).Error
	})
}

func getUserPortraitAssetFolderByIDWithOptionalExternalUser(userId int, id int, externalUserID string, filterExternalUserID bool) (*PortraitAssetFolder, error) {
	if !filterExternalUserID {
		return GetUserPortraitAssetFolderByID(userId, id)
	}
	var folder PortraitAssetFolder
	err := externalUserIDQuery(DB.Where("user_id = ? and id = ?", userId, id), externalUserID).First(&folder).Error
	return &folder, err
}

func MoveUserPortraitAssetsToFolder(userId int, assetKind string, assetIDs []int, folderID int, externalUserID string) error {
	assetKind, ok := NormalizePortraitAssetFolderKind(assetKind)
	if !ok {
		return ErrPortraitAssetFolderInvalidKind
	}
	if len(assetIDs) == 0 {
		return gorm.ErrRecordNotFound
	}
	if len(assetIDs) > PortraitAssetFolderMoveMaxItems {
		return errors.New("too many portrait assets to move")
	}
	externalUserID = NormalizeExternalUserID(externalUserID)
	filterExternalUserID := externalUserID != ""
	if folderID > 0 {
		folder, err := GetUserPortraitAssetFolderByIDForExternalUser(userId, folderID, assetKind, externalUserID, filterExternalUserID)
		if err != nil {
			return err
		}
		externalUserID = NormalizeExternalUserID(folder.ExternalUserID)
		filterExternalUserID = true
	}

	var result *gorm.DB
	switch assetKind {
	case PortraitAssetFolderKindOfficial:
		query := portraitOfficialSourceQuery(DB.Model(&PortraitAssetJob{}).
			Where("user_id = ? and id in ?", userId, assetIDs))
		if filterExternalUserID {
			query = externalUserIDQuery(query, externalUserID)
		}
		result = query.Update("folder_id", folderID)
	case PortraitAssetFolderKindVirtual:
		query := DB.Model(&VirtualPortraitAsset{}).Where("user_id = ? and id in ?", userId, assetIDs)
		if filterExternalUserID {
			query = externalUserIDQuery(query, externalUserID)
		}
		result = query.Update("folder_id", folderID)
	default:
		return ErrPortraitAssetFolderInvalidKind
	}
	if result.Error != nil {
		return result.Error
	}
	if result.RowsAffected != int64(len(assetIDs)) {
		return gorm.ErrRecordNotFound
	}
	return nil
}

func portraitAssetFolderIDQuery(query *gorm.DB, folderID int) *gorm.DB {
	if folderID == 0 {
		return query.Where("(folder_id = ? OR folder_id IS NULL)", 0)
	}
	return query.Where("folder_id = ?", folderID)
}

func hasUserPortraitAssetFolderName(userId int, assetKind string, name string, externalUserID string, excludeID int) (bool, error) {
	query, err := portraitAssetFolderQuery(userId, assetKind, externalUserID, true)
	if err != nil {
		return false, err
	}
	query = query.Model(&PortraitAssetFolder{}).Where("name = ?", name)
	if excludeID > 0 {
		query = query.Where("id <> ?", excludeID)
	}
	var count int64
	if err := query.Count(&count).Error; err != nil {
		return false, err
	}
	return count > 0, nil
}
