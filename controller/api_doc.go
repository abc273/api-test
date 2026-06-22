package controller

import (
	"errors"
	"net/http"
	"strings"

	"github.com/QuantumNous/new-api/common"
	"github.com/QuantumNous/new-api/model"
	"github.com/gin-gonic/gin"
)

type apiDocPublishRequest struct {
	Version         string                        `json:"version"`
	Title           string                        `json:"title"`
	Summary         string                        `json:"summary"`
	ChangedSections []string                      `json:"changed_sections"`
	ChangeItems     []model.ApiDocChangeItemInput `json:"change_items"`
	Content         string                        `json:"content"`
	SourceCommit    string                        `json:"source_commit"`
}

type apiDocPreviewRequest struct {
	Content string `json:"content"`
}

type apiDocDiffRequest struct {
	FromVersion string `json:"from_version"`
	FromContent string `json:"from_content"`
	ToVersion   string `json:"to_version"`
	ToContent   string `json:"to_content"`
}

func getApiDocsFallbackContent() string {
	common.OptionMapRWMutex.RLock()
	defer common.OptionMapRWMutex.RUnlock()
	return common.OptionMap["ApiDocs"]
}

func GetApiDocsMeta(c *gin.Context) {
	revision, err := model.GetLatestApiDocRevision(false)
	if errors.Is(err, model.ErrApiDocRevisionNotFound) {
		common.ApiSuccess(c, gin.H{
			"version":          "legacy",
			"title":            "接口文档",
			"summary":          "当前线上接口文档",
			"changed_sections": []string{},
			"content_sha256":   "",
			"published_at":     int64(0),
			"change_items":     []model.ApiDocChangeItemView{},
		})
		return
	}
	if err != nil {
		common.ApiError(c, err)
		return
	}
	common.ApiSuccess(c, revision)
}

func ListApiDocChangelog(c *gin.Context) {
	pageInfo := common.GetPageQuery(c)
	revisions, total, err := model.ListApiDocRevisions(pageInfo.GetStartIdx(), pageInfo.GetPageSize(), false)
	if err != nil {
		common.ApiError(c, err)
		return
	}
	pageInfo.SetTotal(int(total))
	pageInfo.SetItems(revisions)
	common.ApiSuccess(c, pageInfo)
}

func GetApiDocRevision(c *gin.Context) {
	version := strings.TrimSpace(c.Param("version"))
	revision, err := model.GetApiDocRevisionByVersion(version, true)
	if errors.Is(err, model.ErrApiDocRevisionNotFound) {
		c.JSON(http.StatusOK, gin.H{
			"success": false,
			"message": "文档版本不存在",
		})
		return
	}
	if err != nil {
		common.ApiError(c, err)
		return
	}
	common.ApiSuccess(c, revision)
}

func GetApiDocRevisionDiff(c *gin.Context) {
	toVersion := strings.TrimSpace(c.Param("version"))
	toRevision, err := model.GetApiDocRevisionByVersion(toVersion, true)
	if errors.Is(err, model.ErrApiDocRevisionNotFound) {
		c.JSON(http.StatusOK, gin.H{
			"success": false,
			"message": "文档版本不存在",
		})
		return
	}
	if err != nil {
		common.ApiError(c, err)
		return
	}

	fromVersion := strings.TrimSpace(c.Query("from"))
	var fromContent string
	if fromVersion != "" {
		fromRevision, err := model.GetApiDocRevisionByVersion(fromVersion, true)
		if errors.Is(err, model.ErrApiDocRevisionNotFound) {
			c.JSON(http.StatusOK, gin.H{
				"success": false,
				"message": "对比版本不存在",
			})
			return
		}
		if err != nil {
			common.ApiError(c, err)
			return
		}
		fromContent = fromRevision.Content
	} else {
		previous, err := model.GetPreviousApiDocRevision(toVersion)
		if err == nil {
			fromVersion = previous.Version
			fromContent = previous.Content
		} else {
			fromVersion = "empty"
			fromContent = ""
		}
	}

	diff := model.BuildApiDocDiff(fromVersion, fromContent, toVersion, toRevision.Content)
	common.ApiSuccess(c, diff)
}

func AdminListApiDocRevisions(c *gin.Context) {
	pageInfo := common.GetPageQuery(c)
	includeContent := c.Query("include_content") == "true"
	revisions, total, err := model.ListApiDocRevisions(pageInfo.GetStartIdx(), pageInfo.GetPageSize(), includeContent)
	if err != nil {
		common.ApiError(c, err)
		return
	}
	pageInfo.SetTotal(int(total))
	pageInfo.SetItems(revisions)
	common.ApiSuccess(c, pageInfo)
}

func PublishApiDocs(c *gin.Context) {
	var req apiDocPublishRequest
	if err := common.DecodeJson(c.Request.Body, &req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"success": false,
			"message": "无效的参数",
		})
		return
	}
	revision, err := model.PublishApiDocRevision(model.ApiDocPublishInput{
		Version:         req.Version,
		Title:           req.Title,
		Summary:         req.Summary,
		ChangedSections: req.ChangedSections,
		ChangeItems:     req.ChangeItems,
		Content:         req.Content,
		SourceCommit:    req.SourceCommit,
		PublishedBy:     c.GetInt("id"),
	})
	if err != nil {
		c.JSON(http.StatusOK, gin.H{
			"success": false,
			"message": err.Error(),
		})
		return
	}
	common.ApiSuccess(c, revision)
}

func PreviewApiDocs(c *gin.Context) {
	var req apiDocPreviewRequest
	if err := common.DecodeJson(c.Request.Body, &req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"success": false,
			"message": "无效的参数",
		})
		return
	}
	common.ApiSuccess(c, gin.H{
		"content":      req.Content,
		"content_size": len(req.Content),
	})
}

func DiffApiDocs(c *gin.Context) {
	var req apiDocDiffRequest
	if err := common.DecodeJson(c.Request.Body, &req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"success": false,
			"message": "无效的参数",
		})
		return
	}
	fromVersion := strings.TrimSpace(req.FromVersion)
	fromContent := req.FromContent
	if fromContent == "" {
		fromContent = getApiDocsFallbackContent()
	}
	if fromVersion == "" {
		fromVersion = "current"
	}
	toVersion := strings.TrimSpace(req.ToVersion)
	if toVersion == "" {
		toVersion = "draft"
	}
	diff := model.BuildApiDocDiff(fromVersion, fromContent, toVersion, req.ToContent)
	common.ApiSuccess(c, diff)
}
