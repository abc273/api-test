package controller

import (
	"errors"
	"strconv"

	"github.com/QuantumNous/new-api/common"
	"github.com/QuantumNous/new-api/model"
	"github.com/gin-gonic/gin"
	"gorm.io/gorm"
)

type portraitAssetFolderCreateRequest struct {
	AssetKind      string `json:"asset_kind"`
	Name           string `json:"name"`
	ExternalUserID string `json:"external_user_id"`
	SortOrder      int    `json:"sort_order"`
}

type portraitAssetFolderUpdateRequest struct {
	Name           string  `json:"name"`
	ExternalUserID *string `json:"external_user_id"`
	SortOrder      *int    `json:"sort_order"`
}

type portraitAssetFolderMoveRequest struct {
	AssetKind      string `json:"asset_kind"`
	AssetIDs       []int  `json:"asset_ids"`
	FolderID       int    `json:"folder_id"`
	ExternalUserID string `json:"external_user_id"`
}

func ListPortraitAssetFolders(c *gin.Context) {
	assetKind := c.Query("asset_kind")
	externalUserID, err := getExternalUserIDFromQuery(c)
	if err != nil {
		common.ApiErrorMsg(c, err.Error())
		return
	}
	folders, err := model.ListUserPortraitAssetFolders(c.GetInt("id"), assetKind, externalUserID)
	if err != nil {
		respondPortraitAssetFolderError(c, err)
		return
	}
	common.ApiSuccess(c, folders)
}

func CreatePortraitAssetFolder(c *gin.Context) {
	req := portraitAssetFolderCreateRequest{}
	if err := c.ShouldBindJSON(&req); err != nil {
		common.ApiError(c, err)
		return
	}
	externalUserID, err := normalizeExternalUserIDInput(req.ExternalUserID)
	if err != nil {
		common.ApiErrorMsg(c, err.Error())
		return
	}
	folder, err := model.CreateUserPortraitAssetFolder(c.GetInt("id"), req.AssetKind, req.Name, externalUserID, req.SortOrder)
	if err != nil {
		respondPortraitAssetFolderError(c, err)
		return
	}
	common.ApiSuccess(c, folder)
}

func UpdatePortraitAssetFolder(c *gin.Context) {
	folderID, err := strconv.Atoi(c.Param("folder_id"))
	if err != nil || folderID <= 0 {
		common.ApiErrorMsg(c, "无效的文件夹 ID")
		return
	}
	req := portraitAssetFolderUpdateRequest{}
	if err := c.ShouldBindJSON(&req); err != nil {
		common.ApiError(c, err)
		return
	}
	externalUserID, filterExternalUserID, err := getExternalUserIDFilterFromQuery(c)
	if err != nil {
		common.ApiErrorMsg(c, err.Error())
		return
	}
	if !filterExternalUserID && req.ExternalUserID != nil {
		externalUserID, err = normalizeExternalUserIDInput(*req.ExternalUserID)
		if err != nil {
			common.ApiErrorMsg(c, err.Error())
			return
		}
		filterExternalUserID = true
	}
	folder, err := model.UpdateUserPortraitAssetFolderForExternalUser(c.GetInt("id"), folderID, req.Name, req.SortOrder, externalUserID, filterExternalUserID)
	if err != nil {
		respondPortraitAssetFolderError(c, err)
		return
	}
	common.ApiSuccess(c, folder)
}

func DeletePortraitAssetFolder(c *gin.Context) {
	folderID, err := strconv.Atoi(c.Param("folder_id"))
	if err != nil || folderID <= 0 {
		common.ApiErrorMsg(c, "无效的文件夹 ID")
		return
	}
	externalUserID, filterExternalUserID, err := getExternalUserIDFilterFromQuery(c)
	if err != nil {
		common.ApiErrorMsg(c, err.Error())
		return
	}
	if err := model.DeleteUserPortraitAssetFolderForExternalUser(c.GetInt("id"), folderID, externalUserID, filterExternalUserID); err != nil {
		respondPortraitAssetFolderError(c, err)
		return
	}
	common.ApiSuccess(c, nil)
}

func MovePortraitAssetsToFolder(c *gin.Context) {
	req := portraitAssetFolderMoveRequest{}
	if err := c.ShouldBindJSON(&req); err != nil {
		common.ApiError(c, err)
		return
	}
	if req.FolderID < 0 {
		common.ApiErrorMsg(c, "folder_id 不能小于 0")
		return
	}
	externalUserID, err := normalizeExternalUserIDInput(req.ExternalUserID)
	if err != nil {
		common.ApiErrorMsg(c, err.Error())
		return
	}
	assetIDs := normalizePositiveIDs(req.AssetIDs)
	if err := model.MoveUserPortraitAssetsToFolder(c.GetInt("id"), req.AssetKind, assetIDs, req.FolderID, externalUserID); err != nil {
		respondPortraitAssetFolderError(c, err)
		return
	}
	common.ApiSuccess(c, gin.H{"moved": len(assetIDs)})
}

func normalizePositiveIDs(ids []int) []int {
	seen := make(map[int]struct{}, len(ids))
	result := make([]int, 0, len(ids))
	for _, id := range ids {
		if id <= 0 {
			continue
		}
		if _, exists := seen[id]; exists {
			continue
		}
		seen[id] = struct{}{}
		result = append(result, id)
	}
	return result
}

func respondPortraitAssetFolderError(c *gin.Context, err error) {
	switch {
	case errors.Is(err, model.ErrPortraitAssetFolderInvalidKind):
		common.ApiErrorMsg(c, "asset_kind 仅支持 official 或 virtual")
	case errors.Is(err, model.ErrPortraitAssetFolderNameEmpty):
		common.ApiErrorMsg(c, "文件夹名称不能为空")
	case errors.Is(err, model.ErrPortraitAssetFolderNameTooLong):
		common.ApiErrorMsg(c, "文件夹名称不能超过 128 个字符")
	case errors.Is(err, model.ErrPortraitAssetFolderExists):
		common.ApiErrorMsg(c, "同名文件夹已存在")
	case errors.Is(err, gorm.ErrRecordNotFound):
		common.ApiErrorMsg(c, "文件夹或资产不存在")
	default:
		common.ApiError(c, err)
	}
}
