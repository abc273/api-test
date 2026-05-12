package controller

import (
	"bytes"
	"crypto/subtle"
	"errors"
	"fmt"
	"io"
	"mime"
	"net/http"
	"net/url"
	"os"
	"path"
	"path/filepath"
	"strconv"
	"strings"
	"time"

	"github.com/QuantumNous/new-api/common"
	"github.com/QuantumNous/new-api/model"
	"github.com/QuantumNous/new-api/service"
	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
	"gorm.io/gorm"
)

type portraitAssetCreateRequest struct {
	Name string `json:"name"`
}

type portraitAssetOfficialCreateRequest struct {
	Name string `json:"name"`
}

type portraitAssetOfficialAssetRequest struct {
	AssetURL  string `json:"asset_url"`
	AssetType string `json:"asset_type"`
	Name      string `json:"name"`
}

type portraitAssetOfficialUploadResponse struct {
	URL         string `json:"url"`
	FileName    string `json:"file_name"`
	ContentType string `json:"content_type"`
	AssetType   string `json:"asset_type"`
	Size        int64  `json:"size"`
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

type officialPortraitUploadFormat struct {
	AssetType   string
	ContentType string
	Ext         string
}

const officialPortraitUploadMaxBytes int64 = 100 << 20

var officialPortraitUploadFormatsByExt = map[string]officialPortraitUploadFormat{
	".jpg":  {AssetType: "Image", ContentType: "image/jpeg", Ext: ".jpg"},
	".jpeg": {AssetType: "Image", ContentType: "image/jpeg", Ext: ".jpeg"},
	".png":  {AssetType: "Image", ContentType: "image/png", Ext: ".png"},
	".webp": {AssetType: "Image", ContentType: "image/webp", Ext: ".webp"},
	".gif":  {AssetType: "Image", ContentType: "image/gif", Ext: ".gif"},
	".bmp":  {AssetType: "Image", ContentType: "image/bmp", Ext: ".bmp"},
	".mp4":  {AssetType: "Video", ContentType: "video/mp4", Ext: ".mp4"},
	".mov":  {AssetType: "Video", ContentType: "video/quicktime", Ext: ".mov"},
	".webm": {AssetType: "Video", ContentType: "video/webm", Ext: ".webm"},
	".mp3":  {AssetType: "Audio", ContentType: "audio/mpeg", Ext: ".mp3"},
	".wav":  {AssetType: "Audio", ContentType: "audio/wav", Ext: ".wav"},
	".m4a":  {AssetType: "Audio", ContentType: "audio/mp4", Ext: ".m4a"},
	".ogg":  {AssetType: "Audio", ContentType: "audio/ogg", Ext: ".ogg"},
	".aac":  {AssetType: "Audio", ContentType: "audio/aac", Ext: ".aac"},
	".flac": {AssetType: "Audio", ContentType: "audio/flac", Ext: ".flac"},
}

var officialPortraitUploadFormatsByContentType = map[string]officialPortraitUploadFormat{
	"image/jpeg":      {AssetType: "Image", ContentType: "image/jpeg", Ext: ".jpg"},
	"image/png":       {AssetType: "Image", ContentType: "image/png", Ext: ".png"},
	"image/webp":      {AssetType: "Image", ContentType: "image/webp", Ext: ".webp"},
	"image/gif":       {AssetType: "Image", ContentType: "image/gif", Ext: ".gif"},
	"image/bmp":       {AssetType: "Image", ContentType: "image/bmp", Ext: ".bmp"},
	"video/mp4":       {AssetType: "Video", ContentType: "video/mp4", Ext: ".mp4"},
	"video/webm":      {AssetType: "Video", ContentType: "video/webm", Ext: ".webm"},
	"video/quicktime": {AssetType: "Video", ContentType: "video/quicktime", Ext: ".mov"},
	"audio/mpeg":      {AssetType: "Audio", ContentType: "audio/mpeg", Ext: ".mp3"},
	"audio/wav":       {AssetType: "Audio", ContentType: "audio/wav", Ext: ".wav"},
	"audio/x-wav":     {AssetType: "Audio", ContentType: "audio/wav", Ext: ".wav"},
	"audio/wave":      {AssetType: "Audio", ContentType: "audio/wav", Ext: ".wav"},
	"audio/mp4":       {AssetType: "Audio", ContentType: "audio/mp4", Ext: ".m4a"},
	"audio/aac":       {AssetType: "Audio", ContentType: "audio/aac", Ext: ".aac"},
	"audio/ogg":       {AssetType: "Audio", ContentType: "audio/ogg", Ext: ".ogg"},
	"audio/flac":      {AssetType: "Audio", ContentType: "audio/flac", Ext: ".flac"},
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

func GetOfficialPortraitAssetConfig(c *gin.Context) {
	common.ApiSuccess(c, gin.H{
		"configured":   service.IsVolcPortraitConfigured(),
		"project_name": service.GetVolcPortraitProjectName(),
	})
}

func ListOfficialPortraitAssetJobs(c *gin.Context) {
	userId := c.GetInt("id")
	pageInfo := common.GetPageQuery(c)
	jobs, err := model.GetUserOfficialPortraitAssetJobs(userId, pageInfo.GetStartIdx(), pageInfo.GetPageSize())
	if err != nil {
		common.ApiError(c, err)
		return
	}
	decorateOfficialPortraitAssetJobsResponse(c, jobs)
	total, _ := model.CountUserOfficialPortraitAssetJobs(userId)
	pageInfo.SetTotal(int(total))
	pageInfo.SetItems(jobs)
	common.ApiSuccess(c, pageInfo)
}

func CreateOfficialPortraitAssetJob(c *gin.Context) {
	if !service.IsVolcPortraitConfigured() {
		common.ApiErrorMsg(c, "火山真人资产 API 未配置，请在管理员后台或环境变量中配置 AK/SK")
		return
	}
	req := portraitAssetOfficialCreateRequest{}
	if err := c.ShouldBindJSON(&req); err != nil {
		common.ApiError(c, err)
		return
	}
	req.Name = strings.TrimSpace(req.Name)
	if len([]rune(req.Name)) > 50 {
		common.ApiErrorMsg(c, "资产名称不能超过 50 个字符")
		return
	}
	job, err := model.CreateOfficialPortraitAssetJob(c.GetInt("id"), req.Name, service.GetVolcPortraitProjectName())
	if err != nil {
		if errors.Is(err, model.ErrPortraitAssetActiveJob) {
			common.ApiErrorMsg(c, "你已有进行中的官方真人资产任务，请完成后再创建")
		} else {
			common.ApiError(c, err)
		}
		return
	}
	if err := startOfficialPortraitValidationSession(c, job); err != nil {
		_ = model.FailOfficialPortraitAssetJob(job, err.Error())
		common.ApiError(c, err)
		return
	}
	decorateOfficialPortraitAssetJobResponse(c, job)
	common.ApiSuccess(c, job)
}

func RefreshOfficialPortraitValidation(c *gin.Context) {
	if !service.IsVolcPortraitConfigured() {
		common.ApiErrorMsg(c, "火山真人资产 API 未配置，请在管理员后台或环境变量中配置 AK/SK")
		return
	}
	id, err := strconv.Atoi(c.Param("id"))
	if err != nil {
		common.ApiError(c, err)
		return
	}
	job, err := model.GetUserOfficialPortraitAssetJobByID(c.GetInt("id"), id)
	if err != nil {
		common.ApiError(c, err)
		return
	}
	if job.Status == model.PortraitAssetStatusAssetProcessing ||
		job.Status == model.PortraitAssetStatusPendingConfirm ||
		job.Status == model.PortraitAssetStatusReady {
		common.ApiErrorMsg(c, "当前任务已进入素材处理阶段，不能重新生成认证链接")
		return
	}
	if err := startOfficialPortraitValidationSession(c, job); err != nil {
		common.ApiError(c, err)
		return
	}
	decorateOfficialPortraitAssetJobResponse(c, job)
	common.ApiSuccess(c, job)
}

func UploadOfficialPortraitAssetMaterial(c *gin.Context) {
	uploadPortraitAssetMaterial(c)
}

func uploadPortraitAssetMaterial(c *gin.Context) {
	c.Request.Body = http.MaxBytesReader(c.Writer, c.Request.Body, officialPortraitUploadMaxBytes)
	fileHeader, err := c.FormFile("file")
	if err != nil {
		var maxBytesErr *http.MaxBytesError
		if errors.As(err, &maxBytesErr) {
			common.ApiErrorMsg(c, "上传素材不能超过 100 MB")
			return
		}
		common.ApiErrorMsg(c, "请选择要上传的素材文件")
		return
	}
	if fileHeader.Size <= 0 {
		common.ApiErrorMsg(c, "上传文件不能为空")
		return
	}

	src, err := fileHeader.Open()
	if err != nil {
		common.ApiError(c, err)
		return
	}
	defer src.Close()

	head := make([]byte, 512)
	n, err := io.ReadFull(src, head)
	if err != nil && err != io.ErrUnexpectedEOF && err != io.EOF {
		common.ApiError(c, err)
		return
	}
	head = head[:n]

	detectedContentType := strings.TrimSpace(http.DetectContentType(head))
	declaredContentType := strings.TrimSpace(fileHeader.Header.Get("Content-Type"))
	format, err := resolveOfficialPortraitUploadFormat(fileHeader.Filename, detectedContentType, declaredContentType)
	if err != nil {
		common.ApiErrorMsg(c, "仅支持上传图片、视频或音频素材")
		return
	}

	reader := io.Reader(src)
	if seeker, ok := src.(io.Seeker); ok {
		if _, err := seeker.Seek(0, io.SeekStart); err != nil {
			common.ApiError(c, err)
			return
		}
		reader = src
	} else {
		reader = io.MultiReader(bytes.NewReader(head), src)
	}

	dateDir := time.Now().UTC().Format("20060102")
	fileName := uuid.NewString() + format.Ext
	uploadRoot := common.GetPortraitAssetUploadRoot()
	saveDir := filepath.Join(uploadRoot, dateDir)
	if err := os.MkdirAll(saveDir, 0o755); err != nil {
		common.ApiError(c, err)
		return
	}

	dstPath := filepath.Join(saveDir, fileName)
	dst, err := os.Create(dstPath)
	if err != nil {
		common.ApiError(c, err)
		return
	}
	defer dst.Close()

	if _, err := io.Copy(dst, reader); err != nil {
		_ = os.Remove(dstPath)
		common.ApiError(c, err)
		return
	}

	publicPath := path.Join(common.PortraitAssetUploadRoutePrefix, dateDir, fileName)
	common.ApiSuccess(c, portraitAssetOfficialUploadResponse{
		URL:         buildPublicURL(c, publicPath),
		FileName:    strings.TrimSpace(fileHeader.Filename),
		ContentType: format.ContentType,
		AssetType:   format.AssetType,
		Size:        fileHeader.Size,
	})
}

func SubmitOfficialPortraitAsset(c *gin.Context) {
	id, err := strconv.Atoi(c.Param("id"))
	if err != nil {
		common.ApiError(c, err)
		return
	}
	job, err := model.GetUserOfficialPortraitAssetJobByID(c.GetInt("id"), id)
	if err != nil {
		common.ApiError(c, err)
		return
	}
	req := portraitAssetOfficialAssetRequest{}
	if err := c.ShouldBindJSON(&req); err != nil {
		common.ApiError(c, err)
		return
	}
	req.AssetURL = strings.TrimSpace(req.AssetURL)
	req.AssetType = strings.TrimSpace(req.AssetType)
	req.Name = strings.TrimSpace(req.Name)
	if !isPublicHTTPURL(req.AssetURL) {
		common.ApiErrorMsg(c, "请填写可被火山访问的 http/https 素材 URL")
		return
	}
	if job.VolcGroupID == "" {
		if err := syncOfficialPortraitAssetJob(job); err != nil {
			common.ApiError(c, err)
			return
		}
	}
	if job.VolcGroupID == "" || job.Status != model.PortraitAssetStatusValidated {
		common.ApiErrorMsg(c, "请先完成真人认证")
		return
	}
	assetID, err := service.CreateVolcPortraitAsset(job.VolcGroupID, req.AssetURL, req.AssetType, req.Name, job.ProjectName)
	if err != nil {
		common.ApiError(c, err)
		return
	}
	if err := model.SetOfficialPortraitAssetUpload(job, assetID, req.AssetURL, req.AssetType); err != nil {
		common.ApiError(c, err)
		return
	}
	_ = syncOfficialPortraitAssetJob(job)
	decorateOfficialPortraitAssetJobResponse(c, job)
	common.ApiSuccess(c, job)
}

func SyncOfficialPortraitAssetJob(c *gin.Context) {
	id, err := strconv.Atoi(c.Param("id"))
	if err != nil {
		common.ApiError(c, err)
		return
	}
	job, err := model.GetUserOfficialPortraitAssetJobByID(c.GetInt("id"), id)
	if err != nil {
		common.ApiError(c, err)
		return
	}
	if err := syncOfficialPortraitAssetJob(job); err != nil {
		common.ApiError(c, err)
		return
	}
	decorateOfficialPortraitAssetJobResponse(c, job)
	common.ApiSuccess(c, job)
}

func ConfirmOfficialPortraitAsset(c *gin.Context) {
	id, err := strconv.Atoi(c.Param("id"))
	if err != nil {
		common.ApiError(c, err)
		return
	}
	job, err := model.ConfirmUserOfficialPortraitAsset(c.GetInt("id"), id)
	if err != nil {
		if errors.Is(err, model.ErrPortraitAssetNotReady) {
			common.ApiErrorMsg(c, "请先等待素材入库并确认缩略图")
		} else {
			common.ApiError(c, err)
		}
		return
	}
	decorateOfficialPortraitAssetJobResponse(c, job)
	common.ApiSuccess(c, job)
}

func RejectOfficialPortraitAsset(c *gin.Context) {
	id, err := strconv.Atoi(c.Param("id"))
	if err != nil {
		common.ApiError(c, err)
		return
	}
	job, err := model.RejectUserOfficialPortraitAsset(c.GetInt("id"), id)
	if err != nil {
		if errors.Is(err, model.ErrPortraitAssetNotReady) {
			common.ApiErrorMsg(c, "当前素材不需要确认")
		} else {
			common.ApiError(c, err)
		}
		return
	}
	decorateOfficialPortraitAssetJobResponse(c, job)
	common.ApiSuccess(c, job)
}

func OfficialPortraitAssetPreview(c *gin.Context) {
	id, err := strconv.Atoi(c.Param("id"))
	if err != nil {
		c.String(http.StatusBadRequest, "invalid portrait asset job id")
		return
	}
	job, err := model.GetOfficialPortraitAssetJobByID(id)
	if err != nil {
		c.String(http.StatusNotFound, "portrait asset job not found")
		return
	}
	if c.Param("state") != officialPortraitAssetPreviewState(job) {
		c.String(http.StatusUnauthorized, "invalid portrait asset preview state")
		return
	}

	previewURL, err := resolveOfficialPortraitPreviewURL(job)
	if err != nil {
		c.String(http.StatusNotFound, err.Error())
		return
	}

	c.Header("Cache-Control", "no-store, no-cache, max-age=0, must-revalidate")
	c.Header("Pragma", "no-cache")
	c.Redirect(http.StatusFound, previewURL)
}

func OfficialPortraitAssetCallback(c *gin.Context) {
	idValue := firstNonEmpty(c.Param("id"), c.Query("id"))
	id, err := strconv.Atoi(idValue)
	if err != nil {
		c.String(http.StatusBadRequest, "invalid portrait asset job id")
		return
	}
	job, err := model.GetOfficialPortraitAssetJobByID(id)
	if err != nil {
		c.String(http.StatusNotFound, "portrait asset job not found")
		return
	}
	if firstNonEmpty(c.Param("state"), c.Query("state")) != officialPortraitAssetState(job) {
		c.String(http.StatusUnauthorized, "invalid portrait asset callback state")
		return
	}
	bytedToken := firstNonEmpty(c.Query("bytedToken"), c.Query("byted_token"), job.ValidateToken)
	resultCode := c.Query("resultCode")
	if resultCode == "" {
		resultCode = c.Query("result_code")
	}
	if job.ValidateToken != "" && bytedToken != "" && bytedToken != job.ValidateToken {
		c.String(http.StatusUnauthorized, "invalid portrait asset callback token")
		return
	}
	if err := model.SetOfficialPortraitValidationCallback(job, bytedToken, resultCode); err != nil {
		c.String(http.StatusInternalServerError, err.Error())
		return
	}
	if resultCode == "10000" {
		groupID, err := service.GetVolcPortraitValidateResult(bytedToken, job.ProjectName)
		if err != nil {
			job.ErrorMessage = "真人认证已完成，获取素材组失败：" + err.Error()
			job.UpdatedTime = common.GetTimestamp()
			_ = model.DB.Save(job).Error
		} else {
			_ = model.SetOfficialPortraitAssetGroup(job, groupID)
		}
	}
	redirectURL := "/portrait-assets-official?job_id=" + strconv.Itoa(job.Id)
	if resultCode != "" {
		redirectURL += "&result_code=" + url.QueryEscape(resultCode)
	}
	c.Redirect(http.StatusFound, redirectURL)
}

func startOfficialPortraitValidationSession(c *gin.Context, job *model.PortraitAssetJob) error {
	callbackURL := buildOfficialPortraitCallbackURL(c, job)
	session, err := service.CreateVolcPortraitValidateSession(callbackURL, job.ProjectName)
	if err != nil {
		return err
	}
	return model.SetOfficialPortraitValidateSession(job, session.BytedToken, session.H5Link)
}

func syncOfficialPortraitAssetJob(job *model.PortraitAssetJob) error {
	if job == nil {
		return errors.New("portrait asset job is nil")
	}
	now := common.GetTimestamp()
	if job.Status == model.PortraitAssetStatusValidateReady &&
		job.ValidateResultCode == "" &&
		job.QRExpireTime > 0 &&
		job.QRExpireTime <= now {
		job.Status = model.PortraitAssetStatusExpired
		job.ErrorMessage = "真人认证链接已过期，请重新生成"
		job.UpdatedTime = now
		return model.DB.Save(job).Error
	}
	if job.VolcGroupID == "" && job.ValidateResultCode == "10000" && job.ValidateToken != "" {
		groupID, err := service.GetVolcPortraitValidateResult(job.ValidateToken, job.ProjectName)
		if err != nil {
			return err
		}
		if err := model.SetOfficialPortraitAssetGroup(job, groupID); err != nil {
			return err
		}
	}
	if model.NormalizePortraitAssetID(job.AssetID) != "" &&
		(job.Status == model.PortraitAssetStatusAssetProcessing || job.Status == model.PortraitAssetStatusPendingConfirm) {
		asset, err := service.GetVolcPortraitAsset(job.AssetID, job.ProjectName)
		if err != nil {
			return err
		}
		if err := model.UpdateOfficialPortraitAssetStatus(job, asset.Status, asset.URL); err != nil {
			return err
		}
	}
	return nil
}

func decorateOfficialPortraitAssetJobsResponse(c *gin.Context, jobs []*model.PortraitAssetJob) {
	for _, job := range jobs {
		decorateOfficialPortraitAssetJobResponse(c, job)
	}
}

func decorateOfficialPortraitAssetJobResponse(c *gin.Context, job *model.PortraitAssetJob) {
	if job == nil {
		return
	}
	if previewURL := buildOfficialPortraitPreviewURL(c, job); previewURL != "" {
		job.AssetPreview = previewURL
	}
}

func resolveOfficialPortraitPreviewURL(job *model.PortraitAssetJob) (string, error) {
	if job == nil {
		return "", errors.New("portrait asset job is nil")
	}

	previewURL := ""
	if model.NormalizePortraitAssetID(job.AssetID) != "" {
		asset, err := service.GetVolcPortraitAsset(job.AssetID, job.ProjectName)
		if err != nil {
			common.SysError(fmt.Sprintf("failed to refresh official portrait asset preview for job %d: %v", job.Id, err))
		} else {
			previewURL = strings.TrimSpace(asset.URL)
		}
	}
	previewURL = firstNonEmpty(previewURL, job.AssetPreview, job.AssetURL)
	if previewURL == "" {
		return "", errors.New("portrait asset preview not found")
	}
	if !isPublicHTTPURL(previewURL) {
		return "", errors.New("portrait asset preview url is invalid")
	}
	return previewURL, nil
}

func buildOfficialPortraitCallbackURL(c *gin.Context, job *model.PortraitAssetJob) string {
	baseURL := getPortraitAssetPublicBaseURL(c)
	state := officialPortraitAssetState(job)
	return fmt.Sprintf(
		"%s/api/portrait_assets/official/callback/%s/%s",
		baseURL,
		url.PathEscape(strconv.Itoa(job.Id)),
		url.PathEscape(state),
	)
}

func buildOfficialPortraitPreviewURL(c *gin.Context, job *model.PortraitAssetJob) string {
	if job == nil {
		return ""
	}
	if model.NormalizePortraitAssetID(job.AssetID) == "" &&
		strings.TrimSpace(job.AssetPreview) == "" &&
		strings.TrimSpace(job.AssetURL) == "" {
		return ""
	}
	return buildPublicURL(c, fmt.Sprintf(
		"/api/portrait_assets/official/jobs/%d/preview/%s",
		job.Id,
		officialPortraitAssetPreviewState(job),
	))
}

func officialPortraitAssetState(job *model.PortraitAssetJob) string {
	return common.GenerateHMAC(fmt.Sprintf("portrait-official:%d:%d:%d", job.Id, job.UserId, job.CreatedTime))
}

func officialPortraitAssetPreviewState(job *model.PortraitAssetJob) string {
	return common.GenerateHMAC(fmt.Sprintf("portrait-official-preview:%d:%d:%d", job.Id, job.UserId, job.CreatedTime))
}

func getPortraitAssetPublicBaseURL(c *gin.Context) string {
	baseURL := strings.TrimRight(strings.TrimSpace(service.GetVolcPortraitCallbackBaseURL()), "/")
	if baseURL != "" {
		return baseURL
	}
	proto := c.GetHeader("X-Forwarded-Proto")
	if proto == "" {
		if c.Request.TLS != nil {
			proto = "https"
		} else {
			proto = "http"
		}
	}
	host := c.GetHeader("X-Forwarded-Host")
	if host == "" {
		host = c.Request.Host
	}
	return proto + "://" + host
}

func buildPublicURL(c *gin.Context, routePath string) string {
	return strings.TrimRight(getPortraitAssetPublicBaseURL(c), "/") + "/" + strings.TrimLeft(routePath, "/")
}

func resolveOfficialPortraitUploadFormat(filename string, detectedContentType string, declaredContentType string) (officialPortraitUploadFormat, error) {
	ext := strings.ToLower(filepath.Ext(strings.TrimSpace(filename)))
	detectedContentType = strings.ToLower(strings.TrimSpace(detectedContentType))
	declaredContentType = strings.ToLower(strings.TrimSpace(declaredContentType))

	if format, ok := officialPortraitUploadFormatsByContentType[detectedContentType]; ok {
		if extFormat, extOk := officialPortraitUploadFormatsByExt[ext]; extOk && extFormat.AssetType == format.AssetType {
			format.Ext = extFormat.Ext
		}
		return format, nil
	}
	if format, ok := officialPortraitUploadFormatsByContentType[declaredContentType]; ok {
		if extFormat, extOk := officialPortraitUploadFormatsByExt[ext]; extOk && extFormat.AssetType == format.AssetType {
			format.Ext = extFormat.Ext
		}
		return format, nil
	}
	if format, ok := officialPortraitUploadFormatsByExt[ext]; ok {
		if declaredContentType != "" && declaredContentType != "application/octet-stream" {
			mediaType, _, err := mime.ParseMediaType(declaredContentType)
			if err == nil {
				declaredContentType = strings.ToLower(strings.TrimSpace(mediaType))
			}
		}
		if detectedContentType == "" || detectedContentType == "application/octet-stream" || strings.HasPrefix(detectedContentType, strings.ToLower(strings.Split(format.ContentType, "/")[0])+"/") {
			return format, nil
		}
		if declaredContentType == "" || declaredContentType == "application/octet-stream" || strings.HasPrefix(declaredContentType, strings.ToLower(strings.Split(format.ContentType, "/")[0])+"/") {
			return format, nil
		}
	}
	return officialPortraitUploadFormat{}, errors.New("unsupported official portrait upload type")
}

func isPublicHTTPURL(value string) bool {
	parsed, err := url.ParseRequestURI(value)
	if err != nil || parsed.Host == "" {
		return false
	}
	return parsed.Scheme == "http" || parsed.Scheme == "https"
}

func firstNonEmpty(values ...string) string {
	for _, value := range values {
		if strings.TrimSpace(value) != "" {
			return strings.TrimSpace(value)
		}
	}
	return ""
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
