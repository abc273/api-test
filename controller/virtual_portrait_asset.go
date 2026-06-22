package controller

import (
	"errors"
	"fmt"
	"net/http"
	"strconv"
	"strings"

	"github.com/QuantumNous/new-api/common"
	"github.com/QuantumNous/new-api/model"
	"github.com/QuantumNous/new-api/service"
	"github.com/gin-gonic/gin"
	"gorm.io/gorm"
)

type virtualPortraitAssetCreateRequest struct {
	Name           string `json:"name"`
	AssetURL       string `json:"asset_url"`
	AssetType      string `json:"asset_type"`
	ExternalUserID string `json:"external_user_id"`
	FolderID       int    `json:"folder_id"`
}

func GetVirtualPortraitAssetConfig(c *gin.Context) {
	common.ApiSuccess(c, gin.H{
		"configured":   service.IsVolcPortraitConfigured(),
		"project_name": service.GetVolcPortraitProjectName(),
	})
}

func GetUserVirtualPortraitAssetGroup(c *gin.Context) {
	externalUserID, err := getExternalUserIDFromQuery(c)
	if err != nil {
		common.ApiErrorMsg(c, err.Error())
		return
	}
	group, err := model.GetUserVirtualPortraitAssetGroupForExternalUser(c.GetInt("id"), externalUserID)
	if errors.Is(err, gorm.ErrRecordNotFound) {
		common.ApiSuccess(c, nil)
		return
	}
	if err != nil {
		common.ApiError(c, err)
		return
	}
	common.ApiSuccess(c, group)
}

func ListUserVirtualPortraitAssets(c *gin.Context) {
	userId := c.GetInt("id")
	externalUserID, err := getExternalUserIDFromQuery(c)
	if err != nil {
		common.ApiErrorMsg(c, err.Error())
		return
	}
	pageInfo := common.GetPageQuery(c)
	folderID, filterFolderID, err := getFolderIDFilterFromQuery(c)
	if err != nil {
		common.ApiErrorMsg(c, err.Error())
		return
	}
	assets, err := model.GetUserVirtualPortraitAssetsWithFolderForExternalUser(userId, externalUserID, externalUserID != "", folderID, filterFolderID, pageInfo.GetStartIdx(), pageInfo.GetPageSize())
	if err != nil {
		common.ApiError(c, err)
		return
	}
	service.SyncUserVirtualPortraitAssetsForDisplay(assets)
	decorateUserVirtualPortraitAssetsResponse(c, assets)
	total, _ := model.CountUserVirtualPortraitAssetsWithFolderForExternalUser(userId, externalUserID, externalUserID != "", folderID, filterFolderID)
	pageInfo.SetTotal(int(total))
	pageInfo.SetItems(assets)
	common.ApiSuccess(c, pageInfo)
}

func UploadVirtualPortraitAssetMaterial(c *gin.Context) {
	uploadPortraitAssetMaterial(c)
}

func CreateUserVirtualPortraitAsset(c *gin.Context) {
	if !service.IsVolcPortraitConfigured() {
		common.ApiErrorMsg(c, "火山虚拟人像资产 API 未配置，请在管理员后台或环境变量中配置 AK/SK")
		return
	}

	req := virtualPortraitAssetCreateRequest{}
	if err := c.ShouldBindJSON(&req); err != nil {
		common.ApiError(c, err)
		return
	}

	req.Name = strings.TrimSpace(req.Name)
	req.AssetURL = strings.TrimSpace(req.AssetURL)
	req.AssetType = strings.TrimSpace(req.AssetType)
	var err error
	req.ExternalUserID, err = normalizeExternalUserIDInput(req.ExternalUserID)
	if err != nil {
		common.ApiErrorMsg(c, err.Error())
		return
	}

	if len([]rune(req.Name)) > 50 {
		common.ApiErrorMsg(c, "资产名称不能超过 50 个字符")
		return
	}
	if req.FolderID < 0 {
		common.ApiErrorMsg(c, "folder_id 不能小于 0")
		return
	}
	if !isPublicHTTPURL(req.AssetURL) {
		common.ApiErrorMsg(c, "请填写可被火山访问的 http/https 素材 URL")
		return
	}

	assetType, ok := normalizeVirtualPortraitAssetType(req.AssetType)
	if !ok {
		common.ApiErrorMsg(c, "素材类型仅支持 Image、Video、Audio")
		return
	}

	asset, err := service.CreateUserVirtualPortraitAsset(c.GetInt("id"), req.ExternalUserID, req.Name, req.AssetURL, assetType, req.FolderID)
	if err != nil {
		common.ApiError(c, err)
		return
	}
	decorateUserVirtualPortraitAssetResponse(c, asset)
	common.ApiSuccess(c, asset)
}

func SyncUserVirtualPortraitAsset(c *gin.Context) {
	id, err := strconv.Atoi(c.Param("id"))
	if err != nil {
		common.ApiError(c, err)
		return
	}
	externalUserID, filterExternalUserID, err := getExternalUserIDFilterFromQuery(c)
	if err != nil {
		common.ApiErrorMsg(c, err.Error())
		return
	}
	asset, err := service.SyncUserVirtualPortraitAssetForExternalUser(c.GetInt("id"), id, externalUserID, filterExternalUserID)
	if err != nil {
		common.ApiError(c, err)
		return
	}
	decorateUserVirtualPortraitAssetResponse(c, asset)
	common.ApiSuccess(c, asset)
}

func DeleteUserVirtualPortraitAsset(c *gin.Context) {
	id, err := strconv.Atoi(c.Param("id"))
	if err != nil {
		common.ApiError(c, err)
		return
	}
	externalUserID, filterExternalUserID, err := getExternalUserIDFilterFromQuery(c)
	if err != nil {
		common.ApiErrorMsg(c, err.Error())
		return
	}
	if err := model.DeleteUserVirtualPortraitAssetForExternalUser(c.GetInt("id"), id, externalUserID, filterExternalUserID); err != nil {
		common.ApiError(c, err)
		return
	}
	common.ApiSuccess(c, nil)
}

func UserVirtualPortraitAssetPreview(c *gin.Context) {
	id, err := strconv.Atoi(c.Param("id"))
	if err != nil {
		c.String(http.StatusBadRequest, "invalid virtual portrait asset id")
		return
	}
	asset, err := model.GetVirtualPortraitAssetByID(id)
	if err != nil {
		c.String(http.StatusNotFound, "virtual portrait asset not found")
		return
	}
	if c.Param("state") != virtualPortraitAssetPreviewState(asset) {
		c.String(http.StatusUnauthorized, "invalid virtual portrait asset preview state")
		return
	}

	previewURL, err := resolveUserVirtualPortraitAssetPreviewURL(asset)
	if err != nil {
		c.String(http.StatusNotFound, err.Error())
		return
	}

	c.Header("Cache-Control", "no-store, no-cache, max-age=0, must-revalidate")
	c.Header("Pragma", "no-cache")
	c.Redirect(http.StatusFound, previewURL)
}

func decorateUserVirtualPortraitAssetsResponse(c *gin.Context, assets []*model.VirtualPortraitAsset) {
	for _, asset := range assets {
		decorateUserVirtualPortraitAssetResponse(c, asset)
	}
}

func decorateUserVirtualPortraitAssetResponse(c *gin.Context, asset *model.VirtualPortraitAsset) {
	if asset == nil {
		return
	}
	if previewURL := buildUserVirtualPortraitAssetPreviewURL(c, asset); previewURL != "" {
		asset.PreviewURL = previewURL
	}
}

func resolveUserVirtualPortraitAssetPreviewURL(asset *model.VirtualPortraitAsset) (string, error) {
	if asset == nil {
		return "", errors.New("virtual portrait asset is nil")
	}

	previewURL := ""
	if model.NormalizePortraitAssetID(asset.VolcAssetID) != "" {
		info, err := service.GetVolcPortraitAsset(asset.VolcAssetID, asset.ProjectName)
		if err != nil {
			common.SysError(fmt.Sprintf("failed to refresh virtual portrait asset preview for asset %d: %v", asset.Id, err))
		} else {
			previewURL = strings.TrimSpace(info.URL)
		}
	}
	previewURL = firstNonEmpty(previewURL, asset.PreviewURL, asset.SourceURL)
	if previewURL == "" {
		return "", errors.New("virtual portrait asset preview not found")
	}
	if !isPublicHTTPURL(previewURL) {
		return "", errors.New("virtual portrait asset preview url is invalid")
	}
	return previewURL, nil
}

func buildUserVirtualPortraitAssetPreviewURL(c *gin.Context, asset *model.VirtualPortraitAsset) string {
	if asset == nil {
		return ""
	}
	if model.NormalizePortraitAssetID(asset.VolcAssetID) == "" &&
		strings.TrimSpace(asset.PreviewURL) == "" &&
		strings.TrimSpace(asset.SourceURL) == "" {
		return ""
	}
	return buildPublicURL(c, fmt.Sprintf(
		"/api/portrait_assets/virtual/assets/%d/preview/%s",
		asset.Id,
		virtualPortraitAssetPreviewState(asset),
	))
}

func virtualPortraitAssetPreviewState(asset *model.VirtualPortraitAsset) string {
	return common.GenerateHMAC(fmt.Sprintf("portrait-virtual-preview:%d:%d:%d", asset.Id, asset.UserId, asset.CreatedTime))
}

func normalizeVirtualPortraitAssetType(assetType string) (string, bool) {
	switch strings.ToLower(strings.TrimSpace(assetType)) {
	case "image":
		return "Image", true
	case "video":
		return "Video", true
	case "audio":
		return "Audio", true
	default:
		return "", false
	}
}
