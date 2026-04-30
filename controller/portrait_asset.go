package controller

import (
	"crypto/subtle"
	"errors"
	"net/http"
	"os"
	"strconv"
	"strings"

	"github.com/QuantumNous/new-api/common"
	"github.com/QuantumNous/new-api/model"
	"github.com/gin-gonic/gin"
	"gorm.io/gorm"
)

type portraitAssetCreateRequest struct {
	Name string `json:"name"`
}

type portraitAssetRPAUpdateRequest struct {
	Status       string `json:"status"`
	InviteURL    string `json:"invite_url"`
	QRImage      string `json:"qr_image"`
	VolcGroupID  string `json:"volc_group_id"`
	AssetID      string `json:"asset_id"`
	AssetStatus  string `json:"asset_status"`
	AssetPreview string `json:"asset_preview"`
	ErrorMessage string `json:"error_message"`
}

func ListPortraitAssetJobs(c *gin.Context) {
	userId := c.GetInt("id")
	pageInfo := common.GetPageQuery(c)
	jobs, err := model.GetUserPortraitAssetJobs(userId, pageInfo.GetStartIdx(), pageInfo.GetPageSize())
	if err != nil {
		common.ApiError(c, err)
		return
	}
	total, _ := model.CountUserPortraitAssetJobs(userId)
	pageInfo.SetTotal(int(total))
	pageInfo.SetItems(jobs)
	common.ApiSuccess(c, pageInfo)
}

func CreatePortraitAssetJob(c *gin.Context) {
	req := portraitAssetCreateRequest{}
	if err := c.ShouldBindJSON(&req); err != nil {
		common.ApiError(c, err)
		return
	}
	req.Name = strings.TrimSpace(req.Name)
	if len([]rune(req.Name)) > 50 {
		common.ApiErrorMsg(c, "资产名称不能超过 50 个字符")
		return
	}
	job, err := model.CreatePortraitAssetJob(c.GetInt("id"), req.Name)
	if err != nil {
		if errors.Is(err, model.ErrPortraitAssetActiveJob) {
			common.ApiErrorMsg(c, "你已有进行中的真人资产任务，请完成后再创建")
		} else {
			common.ApiError(c, err)
		}
		return
	}
	common.ApiSuccess(c, job)
}

func GetPortraitAssetJob(c *gin.Context) {
	id, err := strconv.Atoi(c.Param("id"))
	if err != nil {
		common.ApiError(c, err)
		return
	}
	job, err := model.GetUserPortraitAssetJobByID(c.GetInt("id"), id)
	if err != nil {
		common.ApiError(c, err)
		return
	}
	common.ApiSuccess(c, job)
}

func RequestPortraitAssetAccept(c *gin.Context) {
	id, err := strconv.Atoi(c.Param("id"))
	if err != nil {
		common.ApiError(c, err)
		return
	}
	job, err := model.RequestUserPortraitAssetAccept(c.GetInt("id"), id)
	if err != nil {
		if errors.Is(err, model.ErrPortraitAssetQRCodeExpired) {
			common.ApiErrorMsg(c, "二维码已超时，请重新排队")
		} else if errors.Is(err, model.ErrPortraitAssetNotReady) {
			common.ApiErrorMsg(c, "请先等待二维码生成，或确认授权素材已上传完成")
		} else {
			common.ApiError(c, err)
		}
		return
	}
	common.ApiSuccess(c, job)
}

func ConfirmPortraitAsset(c *gin.Context) {
	id, err := strconv.Atoi(c.Param("id"))
	if err != nil {
		common.ApiError(c, err)
		return
	}
	job, err := model.ConfirmUserPortraitAsset(c.GetInt("id"), id)
	if err != nil {
		if errors.Is(err, model.ErrPortraitAssetNotReady) {
			common.ApiErrorMsg(c, "请先等待资产审核通过并确认缩略图")
		} else {
			common.ApiError(c, err)
		}
		return
	}
	common.ApiSuccess(c, job)
}

func RejectPortraitAsset(c *gin.Context) {
	id, err := strconv.Atoi(c.Param("id"))
	if err != nil {
		common.ApiError(c, err)
		return
	}
	job, err := model.RejectUserPortraitAsset(c.GetInt("id"), id)
	if err != nil {
		if errors.Is(err, model.ErrPortraitAssetNotReady) {
			common.ApiErrorMsg(c, "当前资产不需要确认")
		} else {
			common.ApiError(c, err)
		}
		return
	}
	common.ApiSuccess(c, job)
}

func requirePortraitRPAAuth(c *gin.Context) bool {
	secret := os.Getenv("PORTRAIT_RPA_SECRET")
	if strings.TrimSpace(secret) == "" {
		c.JSON(http.StatusForbidden, gin.H{
			"success": false,
			"message": "PORTRAIT_RPA_SECRET is not configured",
		})
		return false
	}
	token := c.GetHeader("Authorization")
	if strings.HasPrefix(strings.ToLower(token), "bearer ") {
		token = strings.TrimSpace(token[7:])
	}
	if subtle.ConstantTimeCompare([]byte(token), []byte(secret)) != 1 {
		c.JSON(http.StatusUnauthorized, gin.H{
			"success": false,
			"message": "invalid portrait RPA secret",
		})
		return false
	}
	return true
}

func ListPortraitAssetJobsForRPA(c *gin.Context) {
	if !requirePortraitRPAAuth(c) {
		return
	}
	limit, _ := strconv.Atoi(c.DefaultQuery("limit", "10"))
	jobs, err := model.ListPortraitAssetJobsForRPA(c.Query("status"), limit)
	if err != nil {
		common.ApiError(c, err)
		return
	}
	common.ApiSuccess(c, jobs)
}

func UpdatePortraitAssetJobFromRPA(c *gin.Context) {
	if !requirePortraitRPAAuth(c) {
		return
	}
	id, err := strconv.Atoi(c.Param("id"))
	if err != nil {
		common.ApiError(c, err)
		return
	}
	var job model.PortraitAssetJob
	if err := model.DB.First(&job, id).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			common.ApiErrorMsg(c, "portrait asset job not found")
		} else {
			common.ApiError(c, err)
		}
		return
	}
	req := portraitAssetRPAUpdateRequest{}
	if err := c.ShouldBindJSON(&req); err != nil {
		common.ApiError(c, err)
		return
	}
	if req.Status != "" {
		job.Status = strings.TrimSpace(req.Status)
	}
	if req.InviteURL != "" {
		job.InviteURL = strings.TrimSpace(req.InviteURL)
	}
	if req.QRImage != "" {
		job.QRImage = strings.TrimSpace(req.QRImage)
	}
	if req.VolcGroupID != "" {
		job.VolcGroupID = strings.TrimSpace(req.VolcGroupID)
	}
	if req.AssetID != "" {
		job.AssetID = model.NormalizePortraitAssetID(req.AssetID)
	}
	if req.AssetStatus != "" {
		job.AssetStatus = strings.TrimSpace(req.AssetStatus)
	}
	if req.AssetPreview != "" {
		job.AssetPreview = strings.TrimSpace(req.AssetPreview)
	}
	if req.ErrorMessage != "" {
		job.ErrorMessage = strings.TrimSpace(req.ErrorMessage)
	} else if req.Status != "" && strings.TrimSpace(req.Status) != model.PortraitAssetStatusFailed {
		job.ErrorMessage = ""
	}
	if job.Status == "" {
		job.Status = model.PortraitAssetStatusPending
	}
	if job.AssetID != "" && req.Status == "" && job.Status != model.PortraitAssetStatusFailed && job.Status != model.PortraitAssetStatusDisabled {
		job.Status = model.PortraitAssetStatusReady
	}
	if err := model.UpdatePortraitAssetJobFromRPA(&job); err != nil {
		common.ApiError(c, err)
		return
	}
	common.ApiSuccess(c, &job)
}
