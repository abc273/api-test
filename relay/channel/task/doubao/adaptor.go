package doubao

import (
	"bytes"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"strconv"
	"strings"
	"time"

	"github.com/QuantumNous/new-api/common"

	"github.com/QuantumNous/new-api/constant"
	"github.com/QuantumNous/new-api/dto"
	"github.com/QuantumNous/new-api/model"
	"github.com/QuantumNous/new-api/relay/channel"
	"github.com/QuantumNous/new-api/relay/channel/task/taskcommon"
	relaycommon "github.com/QuantumNous/new-api/relay/common"
	"github.com/QuantumNous/new-api/service"
	"github.com/QuantumNous/new-api/setting/billing_setting"

	"github.com/gin-gonic/gin"
	"github.com/pkg/errors"
	"github.com/samber/lo"
)

// ============================
// Request / Response structures
// ============================

type ContentItem struct {
	Type     string    `json:"type,omitempty"`
	Text     string    `json:"text,omitempty"`
	ImageURL *MediaURL `json:"image_url,omitempty"`
	VideoURL *MediaURL `json:"video_url,omitempty"`
	AudioURL *MediaURL `json:"audio_url,omitempty"`
	Role     string    `json:"role,omitempty"`
}

type MediaURL struct {
	URL string `json:"url,omitempty"`
}

type requestPayload struct {
	Model                 string         `json:"model"`
	Content               []ContentItem  `json:"content,omitempty"`
	CallbackURL           string         `json:"callback_url,omitempty"`
	ReturnLastFrame       *dto.BoolValue `json:"return_last_frame,omitempty"`
	ServiceTier           string         `json:"service_tier,omitempty"`
	ExecutionExpiresAfter *dto.IntValue  `json:"execution_expires_after,omitempty"`
	GenerateAudio         *dto.BoolValue `json:"generate_audio,omitempty"`
	Draft                 *dto.BoolValue `json:"draft,omitempty"`
	Tools                 []struct {
		Type string `json:"type,omitempty"`
	} `json:"tools,omitempty"`
	Resolution  string         `json:"resolution,omitempty"`
	Ratio       string         `json:"ratio,omitempty"`
	Duration    *dto.IntValue  `json:"duration,omitempty"`
	Frames      *dto.IntValue  `json:"frames,omitempty"`
	Seed        *dto.IntValue  `json:"seed,omitempty"`
	CameraFixed *dto.BoolValue `json:"camera_fixed,omitempty"`
	Watermark   *dto.BoolValue `json:"watermark,omitempty"`
}

type responsePayload struct {
	ID string `json:"id"` // task_id
}

type responseTask struct {
	ID      string `json:"id"`
	Model   string `json:"model"`
	Status  string `json:"status"`
	Content struct {
		VideoURL string `json:"video_url"`
	} `json:"content"`
	Seed            int    `json:"seed"`
	Resolution      string `json:"resolution"`
	Duration        int    `json:"duration"`
	Ratio           string `json:"ratio"`
	FramesPerSecond int    `json:"framespersecond"`
	ServiceTier     string `json:"service_tier"`
	Tools           []struct {
		Type string `json:"type"`
	} `json:"tools"`
	Usage struct {
		CompletionTokens int `json:"completion_tokens"`
		TotalTokens      int `json:"total_tokens"`
		ToolUsage        struct {
			WebSearch int `json:"web_search"`
		} `json:"tool_usage"`
	} `json:"usage"`
	Error struct {
		Code    string `json:"code"`
		Message string `json:"message"`
	} `json:"error"`
	CreatedAt int64 `json:"created_at"`
	UpdatedAt int64 `json:"updated_at"`
}

// ============================
// Adaptor implementation
// ============================

type TaskAdaptor struct {
	taskcommon.BaseBilling
	ChannelType int
	apiKey      string
	baseURL     string
}

func (a *TaskAdaptor) Init(info *relaycommon.RelayInfo) {
	a.ChannelType = info.ChannelType
	a.baseURL = info.ChannelBaseUrl
	a.apiKey = info.ApiKey
}

// ValidateRequestAndSetAction parses body, validates fields and sets default action.
func (a *TaskAdaptor) ValidateRequestAndSetAction(c *gin.Context, info *relaycommon.RelayInfo) (taskErr *dto.TaskError) {
	// Accept only POST /v1/video/generations as "generate" action.
	return relaycommon.ValidateBasicTaskRequest(c, info, constant.TaskActionGenerate)
}

// BuildRequestURL constructs the upstream URL.
func (a *TaskAdaptor) BuildRequestURL(_ *relaycommon.RelayInfo) (string, error) {
	return fmt.Sprintf("%s/api/v3/contents/generations/tasks", a.baseURL), nil
}

// BuildRequestHeader sets required headers.
func (a *TaskAdaptor) BuildRequestHeader(_ *gin.Context, req *http.Request, _ *relaycommon.RelayInfo) error {
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Accept", "application/json")
	req.Header.Set("Authorization", "Bearer "+a.apiKey)
	return nil
}

// EstimateBilling 检测请求 metadata 中是否包含视频输入，返回视频折扣 OtherRatio。
func (a *TaskAdaptor) EstimateBilling(c *gin.Context, info *relaycommon.RelayInfo) map[string]float64 {
	// output_tier_price 已经把视频输入维度编码进分档单价里，
	// 这里跳过旧的 video_input 折扣，避免重复叠乘。
	if billing_setting.GetBillingMode(info.OriginModelName) == billing_setting.BillingModeOutputTierPrice {
		return nil
	}
	req, err := relaycommon.GetTaskRequest(c)
	if err != nil {
		return nil
	}
	if hasVideoInMetadata(req.Metadata) {
		if ratio, ok := GetVideoInputRatio(info.OriginModelName); ok {
			return map[string]float64{"video_input": ratio}
		}
	}
	return nil
}

// hasVideoInMetadata 直接检查 metadata 的 content 数组是否包含 video_url 条目，
// 避免构建完整的上游 requestPayload。
func hasVideoInMetadata(metadata map[string]interface{}) bool {
	if metadata == nil {
		return false
	}
	contentRaw, ok := metadata["content"]
	if !ok {
		return false
	}
	contentSlice, ok := contentRaw.([]interface{})
	if !ok {
		return false
	}
	for _, item := range contentSlice {
		itemMap, ok := item.(map[string]interface{})
		if !ok {
			continue
		}
		if itemMap["type"] == "video_url" {
			return true
		}
		if _, has := itemMap["video_url"]; has {
			return true
		}
	}
	return false
}

// BuildRequestBody converts request into Doubao specific format.
func (a *TaskAdaptor) BuildRequestBody(c *gin.Context, info *relaycommon.RelayInfo) (io.Reader, error) {
	req, err := relaycommon.GetTaskRequest(c)
	if err != nil {
		return nil, err
	}

	body, err := a.convertToRequestPayload(&req, info)
	if err != nil {
		return nil, errors.Wrap(err, "convert request payload failed")
	}
	if info.IsModelMapped {
		body.Model = info.UpstreamModelName
	} else {
		info.UpstreamModelName = body.Model
	}
	data, err := common.Marshal(body)
	if err != nil {
		return nil, err
	}
	return bytes.NewReader(data), nil
}

// DoRequest delegates to common helper.
func (a *TaskAdaptor) DoRequest(c *gin.Context, info *relaycommon.RelayInfo, requestBody io.Reader) (*http.Response, error) {
	return channel.DoTaskApiRequest(a, c, info, requestBody)
}

func (a *TaskAdaptor) DeleteTask(c *gin.Context, info *relaycommon.RelayInfo, originTask *model.Task) *dto.TaskError {
	if originTask == nil {
		return service.TaskErrorWrapperLocal(fmt.Errorf("task is required"), "invalid_task", http.StatusBadRequest)
	}
	upstreamTaskID := strings.TrimSpace(originTask.GetUpstreamTaskID())
	if upstreamTaskID == "" {
		return service.TaskErrorWrapperLocal(fmt.Errorf("upstream task id is empty"), "invalid_upstream_task_id", http.StatusBadRequest)
	}

	deleteURL := fmt.Sprintf("%s/api/v3/contents/generations/tasks/%s", a.baseURL, url.PathEscape(upstreamTaskID))
	req, err := http.NewRequest(http.MethodDelete, deleteURL, nil)
	if err != nil {
		return service.TaskErrorWrapper(err, "build_delete_request_failed", http.StatusInternalServerError)
	}
	if err = a.BuildRequestHeader(c, req, info); err != nil {
		return service.TaskErrorWrapper(err, "build_delete_request_header_failed", http.StatusInternalServerError)
	}
	client, err := service.GetHttpClientWithProxy(info.ChannelSetting.Proxy)
	if err != nil {
		return service.TaskErrorWrapper(err, "build_delete_http_client_failed", http.StatusInternalServerError)
	}
	resp, err := client.Do(req)
	if err != nil {
		return service.TaskErrorWrapper(err, "delete_task_request_failed", http.StatusInternalServerError)
	}
	if resp == nil {
		return service.TaskErrorWrapperLocal(fmt.Errorf("delete task response is nil"), "delete_task_response_nil", http.StatusInternalServerError)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK && resp.StatusCode != http.StatusNoContent {
		responseBody, _ := io.ReadAll(resp.Body)
		return service.TaskErrorWrapper(fmt.Errorf("%s", string(responseBody)), "delete_task_failed", resp.StatusCode)
	}
	return nil
}

// DoResponse handles upstream response, returns taskID etc.
func (a *TaskAdaptor) DoResponse(c *gin.Context, resp *http.Response, info *relaycommon.RelayInfo) (taskID string, taskData []byte, taskErr *dto.TaskError) {
	responseBody, err := io.ReadAll(resp.Body)
	if err != nil {
		taskErr = service.TaskErrorWrapper(err, "read_response_body_failed", http.StatusInternalServerError)
		return
	}
	_ = resp.Body.Close()

	// Parse Doubao response
	var dResp responsePayload
	if err := common.Unmarshal(responseBody, &dResp); err != nil {
		taskErr = service.TaskErrorWrapper(errors.Wrapf(err, "body: %s", responseBody), "unmarshal_response_body_failed", http.StatusInternalServerError)
		return
	}

	if dResp.ID == "" {
		taskErr = service.TaskErrorWrapper(fmt.Errorf("task_id is empty"), "invalid_response", http.StatusInternalServerError)
		return
	}

	ov := dto.NewOpenAIVideo()
	ov.ID = info.PublicTaskID
	ov.TaskID = info.PublicTaskID
	ov.CreatedAt = time.Now().Unix()
	ov.Model = info.OriginModelName

	c.JSON(http.StatusOK, ov)
	return dResp.ID, responseBody, nil
}

// FetchTask fetch task status
func (a *TaskAdaptor) FetchTask(baseUrl, key string, body map[string]any, proxy string) (*http.Response, error) {
	taskID, ok := body["task_id"].(string)
	if !ok {
		return nil, fmt.Errorf("invalid task_id")
	}

	uri := fmt.Sprintf("%s/api/v3/contents/generations/tasks/%s", baseUrl, taskID)

	req, err := http.NewRequest(http.MethodGet, uri, nil)
	if err != nil {
		return nil, err
	}

	req.Header.Set("Accept", "application/json")
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Authorization", "Bearer "+key)

	client, err := service.GetHttpClientWithProxy(proxy)
	if err != nil {
		return nil, fmt.Errorf("new proxy http client failed: %w", err)
	}
	return client.Do(req)
}

func (a *TaskAdaptor) GetModelList() []string {
	return ModelList
}

func (a *TaskAdaptor) GetChannelName() string {
	return ChannelName
}

func (a *TaskAdaptor) convertToRequestPayload(req *relaycommon.TaskSubmitReq, info *relaycommon.RelayInfo) (*requestPayload, error) {
	modelName := strings.TrimSpace(req.Model)
	if info != nil && info.ChannelMeta != nil && strings.TrimSpace(info.UpstreamModelName) != "" {
		modelName = strings.TrimSpace(info.UpstreamModelName)
	}
	r := requestPayload{
		Model:   modelName,
		Content: []ContentItem{},
	}

	imageItems := buildReferenceImageItems(req.Images)

	metadata := req.Metadata
	if err := taskcommon.UnmarshalMetadata(metadata, &r); err != nil {
		return nil, errors.Wrap(err, "unmarshal metadata failed")
	}
	if len(imageItems) > 0 {
		// Keep explicit metadata.content ordering stable. Legacy top-level images
		// are appended after explicit media so "图1/图2/图3" numbering does not shift.
		r.Content = append(r.Content, imageItems...)
	}
	if r.Resolution == "" && strings.TrimSpace(req.Resolution) != "" {
		r.Resolution = strings.TrimSpace(req.Resolution)
	}
	if r.Ratio == "" && strings.TrimSpace(req.Ratio) != "" {
		r.Ratio = strings.TrimSpace(req.Ratio)
	}
	if r.Resolution == "" {
		switch service.NormalizeVideoSuperResolutionModelAlias(strings.TrimSpace(req.Model)) {
		case service.Seedance15ModelAlias, service.Seedance2ModelAlias, service.Seedance20ModelAlias, service.Seedance20FastModelAlias, service.SD20FastModelAlias:
			r.Resolution = "720p"
		}
	}
	r.Resolution = service.ResolveDedicatedVideoSuperResolutionSourceResolution(req.Model, r.Resolution)

	externalUserID, err := resolveRequestExternalUserID(req)
	if err != nil {
		return nil, err
	}
	if portraitAssetItem, err := resolvePortraitAssetContentItem(req, info, externalUserID); err != nil {
		return nil, err
	} else if portraitAssetItem != nil {
		r.Content = append(r.Content, *portraitAssetItem)
	}

	userID := 0
	if info != nil {
		userID = info.UserId
	}
	if err := validatePortraitAssetReferences(userID, externalUserID, r.Content); err != nil {
		return nil, err
	}

	if req.Duration > 0 {
		r.Duration = lo.ToPtr(dto.IntValue(req.Duration))
	} else if sec, _ := strconv.Atoi(req.Seconds); sec > 0 {
		r.Duration = lo.ToPtr(dto.IntValue(sec))
	}

	r.Content = lo.Reject(r.Content, func(c ContentItem, _ int) bool { return c.Type == "text" })
	r.Content = append(r.Content, ContentItem{
		Type: "text",
		Text: req.Prompt,
	})

	return &r, nil
}

func buildReferenceImageItems(images []string) []ContentItem {
	items := make([]ContentItem, 0, len(images))
	for _, imgURL := range images {
		items = append(items, ContentItem{
			Type: "image_url",
			Role: "reference_image",
			ImageURL: &MediaURL{
				URL: imgURL,
			},
		})
	}
	return items
}

func resolveRequestExternalUserID(req *relaycommon.TaskSubmitReq) (string, error) {
	if req == nil {
		return "", nil
	}
	if externalUserID := model.NormalizeExternalUserID(req.ExternalUserID); externalUserID != "" {
		return validateRequestExternalUserID(externalUserID)
	}
	if req.Metadata == nil {
		return "", nil
	}
	if raw, ok := req.Metadata["external_user_id"]; ok {
		if externalUserID, ok := raw.(string); ok {
			return validateRequestExternalUserID(externalUserID)
		}
	}
	return "", nil
}

func validateRequestExternalUserID(externalUserID string) (string, error) {
	externalUserID = model.NormalizeExternalUserID(externalUserID)
	if len([]rune(externalUserID)) > model.ExternalUserIDMaxRunes {
		return "", fmt.Errorf("external_user_id must not exceed %d characters", model.ExternalUserIDMaxRunes)
	}
	return externalUserID, nil
}

func resolvePortraitAssetContentItem(req *relaycommon.TaskSubmitReq, info *relaycommon.RelayInfo, externalUserID string) (*ContentItem, error) {
	if req.Metadata == nil {
		return nil, nil
	}
	userID := 0
	if info != nil {
		userID = info.UserId
	}
	if raw, ok := req.Metadata["portrait_asset_id"]; ok {
		id, ok := parsePortraitAssetJobID(raw)
		if !ok || id <= 0 {
			return nil, fmt.Errorf("metadata.portrait_asset_id is invalid")
		}
		job, err := model.GetReadyUserPortraitAssetJobForExternalUser(userID, id, externalUserID)
		if err != nil && model.NormalizeExternalUserID(externalUserID) == "" {
			job, err = model.GetReadyUserPortraitAssetJobByIDIgnoringExternalUser(userID, id)
		}
		if err != nil {
			return nil, fmt.Errorf("portrait asset %d is not ready or not bound to current user", id)
		}
		item := buildPortraitAssetContentItem(model.PortraitAssetURI(job.AssetID), job.AssetType)
		return &item, nil
	}
	if raw, ok := req.Metadata["asset_id"]; ok {
		assetID, ok := raw.(string)
		if !ok {
			return nil, fmt.Errorf("metadata.asset_id must be a string")
		}
		assetID = model.NormalizePortraitAssetID(assetID)
		if assetID == "" {
			return nil, nil
		}
		job, err := model.GetReadyUserPortraitAssetByAssetIDForExternalUser(userID, assetID, externalUserID)
		if err != nil && model.NormalizeExternalUserID(externalUserID) == "" {
			job, err = model.GetReadyUserPortraitAssetByAssetIDIgnoringExternalUser(userID, assetID)
		}
		if err != nil {
			return nil, fmt.Errorf("portrait asset %s is not ready or not bound to current user", assetID)
		}
		item := buildPortraitAssetContentItem(model.PortraitAssetURI(job.AssetID), job.AssetType)
		return &item, nil
	}
	return nil, nil
}

func buildPortraitAssetContentItem(assetURI string, assetType string) ContentItem {
	normalizedType := strings.TrimSpace(strings.ToLower(assetType))
	switch normalizedType {
	case "video":
		return ContentItem{
			Type: "video_url",
			Role: "reference_video",
			VideoURL: &MediaURL{
				URL: assetURI,
			},
		}
	case "audio":
		return ContentItem{
			Type: "audio_url",
			AudioURL: &MediaURL{
				URL: assetURI,
			},
		}
	default:
		return ContentItem{
			Type: "image_url",
			Role: "reference_image",
			ImageURL: &MediaURL{
				URL: assetURI,
			},
		}
	}
}

func parsePortraitAssetJobID(raw any) (int, bool) {
	switch v := raw.(type) {
	case int:
		return v, true
	case int64:
		return int(v), true
	case float64:
		return int(v), v == float64(int(v))
	case string:
		id, err := strconv.Atoi(strings.TrimSpace(v))
		return id, err == nil
	default:
		return 0, false
	}
}

func validatePortraitAssetReferences(userId int, externalUserID string, content []ContentItem) error {
	for _, item := range content {
		url := strings.TrimSpace(contentItemMediaURL(item))
		if !strings.HasPrefix(url, "asset://") {
			continue
		}
		assetID := model.NormalizePortraitAssetID(url)
		if assetID == "" {
			return fmt.Errorf("portrait asset reference is empty")
		}
		if err := service.ValidateUserOwnedPortraitAssetByAssetIDForExternalUser(userId, assetID, externalUserID); err != nil {
			return fmt.Errorf("portrait asset %s is not bound to current user", assetID)
		}
	}
	return nil
}

func contentItemMediaURL(item ContentItem) string {
	if item.ImageURL != nil {
		return item.ImageURL.URL
	}
	if item.VideoURL != nil {
		return item.VideoURL.URL
	}
	if item.AudioURL != nil {
		return item.AudioURL.URL
	}
	return ""
}

func (a *TaskAdaptor) ParseTaskResult(respBody []byte) (*relaycommon.TaskInfo, error) {
	resTask := responseTask{}
	if err := common.Unmarshal(respBody, &resTask); err != nil {
		return nil, errors.Wrap(err, "unmarshal task result failed")
	}

	taskResult := relaycommon.TaskInfo{
		Code: 0,
	}

	// Map Doubao status to internal status
	switch resTask.Status {
	case "pending", "queued":
		taskResult.Status = model.TaskStatusQueued
		taskResult.Progress = "10%"
	case "processing", "running":
		taskResult.Status = model.TaskStatusInProgress
		taskResult.Progress = "50%"
	case "succeeded":
		taskResult.Status = model.TaskStatusSuccess
		taskResult.Progress = "100%"
		taskResult.Url = resTask.Content.VideoURL
		// 解析 usage 信息用于按倍率计费
		taskResult.CompletionTokens = resTask.Usage.CompletionTokens
		taskResult.TotalTokens = resTask.Usage.TotalTokens
	case "failed":
		taskResult.Status = model.TaskStatusFailure
		taskResult.Progress = "100%"
		taskResult.Reason = resTask.Error.Message
	case "cancelled":
		taskResult.Status = model.TaskStatusCancelled
		taskResult.Progress = "100%"
	case "expired":
		taskResult.Status = model.TaskStatusFailure
		taskResult.Progress = "100%"
		taskResult.Reason = "task expired"
	default:
		// Unknown status, treat as processing
		taskResult.Status = model.TaskStatusInProgress
		taskResult.Progress = "30%"
	}

	return &taskResult, nil
}

func (a *TaskAdaptor) ConvertToOpenAIVideo(originTask *model.Task) ([]byte, error) {
	var dResp responseTask
	originalData := service.ExtractOriginalVideoTaskData(originTask.Data)
	if err := common.Unmarshal(originalData, &dResp); err != nil {
		return nil, errors.Wrap(err, "unmarshal doubao task data failed")
	}

	openAIVideo := dto.NewOpenAIVideo()
	openAIVideo.ID = originTask.TaskID
	openAIVideo.TaskID = originTask.TaskID
	openAIVideo.Status = originTask.Status.ToVideoStatus()
	openAIVideo.SetProgressStr(originTask.Progress)
	srState, hasSRState, srStateErr := service.ReadVideoSuperResolutionState(originTask.Data)
	if originTask.Status == model.TaskStatusSuccess && originTask.GetResultURL() != "" {
		openAIVideo.SetMetadata("url", originTask.GetResultURL())
	}
	if srStateErr == nil && hasSRState && srState != nil && !service.ShouldHideVideoSuperResolutionMetadata(originTask.Properties.OriginModelName) {
		openAIVideo.SetMetadata("super_resolution", map[string]any{
			"type":                     srState.Type,
			"task_id":                  srState.TaskID,
			"status":                   srState.Status,
			"source_url":               srState.SourceURL,
			"source_resolution":        srState.SourceResolution,
			"target_resolution":        srState.TargetResolution,
			"target_width":             srState.TargetWidth,
			"target_height":            srState.TargetHeight,
			"result_url":               srState.ResultURL,
			"output_tos_url":           srState.OutputTOSURL,
			"last_error":               srState.LastError,
			"normalize_task_id":        srState.NormalizeTaskID,
			"normalize_status":         srState.NormalizeStatus,
			"normalize_result_url":     srState.NormalizeResultURL,
			"normalize_output_tos_url": srState.NormalizeOutputTOSURL,
			"normalize_last_error":     srState.NormalizeLastError,
		})
	} else if dResp.Content.VideoURL != "" && !(originTask.Status == model.TaskStatusSuccess && originTask.GetResultURL() != "") {
		openAIVideo.SetMetadata("url", dResp.Content.VideoURL)
	}
	openAIVideo.CreatedAt = originTask.CreatedAt
	if originTask.Status == model.TaskStatusSuccess || originTask.Status == model.TaskStatusFailure || originTask.Status == model.TaskStatusCancelled {
		openAIVideo.CompletedAt = originTask.UpdatedAt
	}
	openAIVideo.Model = originTask.Properties.OriginModelName

	if dResp.Status == "failed" {
		openAIVideo.Error = &dto.OpenAIVideoError{
			Message: dResp.Error.Message,
			Code:    dResp.Error.Code,
		}
	}

	return common.Marshal(openAIVideo)
}
