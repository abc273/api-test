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
	Name      string `json:"name"`
	AssetURL  string `json:"asset_url"`
	AssetType string `json:"asset_type"`
}

func GetVirtualPortraitAssetConfig(c *gin.Context) {
	common.ApiSuccess(c, gin.H{
		"configured":   service.IsVolcPortraitConfigured(),
		"project_name": service.GetVolcPortraitProjectName(),
	})
}

func GetUserVirtualPortraitAssetGroup(c *gin.Context) {
	group, err := model.GetUserVirtualPortraitAssetGroup(c.GetInt("id"))
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
	pageInfo := common.GetPageQuery(c)
	assets, err := model.GetUserVirtualPortraitAssets(userId, pageInfo.GetStartIdx(), pageInfo.GetPageSize())
	if err != nil {
		common.ApiError(c, err)
		return
	}
	decorateUserVirtualPortraitAssetsResponse(c, assets)
	total, _ := model.CountUserVirtualPortraitAssets(userId)
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

	if len([]rune(req.Name)) > 50 {
		common.ApiErrorMsg(c, "资产名称不能超过 50 个字符")
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

	asset, err := service.CreateUserVirtualPortraitAsset(c.GetInt("id"), req.Name, req.AssetURL, assetType)
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
	asset, err := service.SyncUserVirtualPortraitAsset(c.GetInt("id"), id)
	if err != nil {
		common.ApiError(c, err)
		return
	}
	decorateUserVirtualPortraitAssetResponse(c, asset)
	common.ApiSuccess(c, asset)
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
