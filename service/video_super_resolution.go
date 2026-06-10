package service

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	neturl "net/url"
	"path"
	"strings"

	"github.com/QuantumNous/new-api/common"
	"github.com/QuantumNous/new-api/model"
	relaycommon "github.com/QuantumNous/new-api/relay/common"
	videoSuperResolutionSetting "github.com/QuantumNous/new-api/setting/video_super_resolution_setting"
	"github.com/volcengine/ve-tos-golang-sdk/v2/tos"
	"github.com/volcengine/ve-tos-golang-sdk/v2/tos/enum"
)

const (
	videoSuperResolutionTaskDataKind = "video_super_resolution_chain"
	videoSuperResolutionTypeLAS      = "volc_las_video_super_resolution"
	videoSuperResolutionOptInKey     = "enable_video_super_resolution"

	Seedance15ModelAlias       = "seedance1.5"
	Seedance15SRModelAlias     = "seedance1.5-sr"
	Seedance2ModelAlias        = "seedance2"
	Seedance20ModelAlias       = "seedance2.0"
	Seedance2SRModelAlias      = "seedance2-sr"
	Seedance20SRModelAlias     = "seedance2.0-sr"
	Seedance20FastModelAlias   = "seedance2.0fast"
	Seedance20FastSRModelAlias = "seedance2.0fast-sr"
	SD20FastModelAlias         = "sd2.0fast"
	SD20FastSRModelAlias       = "sd2.0fast-sr"
	seedance2UpstreamModel     = "doubao-seedance-2-0-260128"
	sd20FastUpstreamModel      = "doubao-seedance-2-0-fast-260128"
)

type VideoSuperResolutionTaskData struct {
	Kind        string                     `json:"kind,omitempty"`
	Original    json.RawMessage            `json:"original,omitempty"`
	PostProcess *VideoSuperResolutionState `json:"post_process,omitempty"`
}

type VideoSuperResolutionState struct {
	Type             string `json:"type,omitempty"`
	TaskID           string `json:"task_id,omitempty"`
	Status           string `json:"status,omitempty"`
	SourceURL        string `json:"source_url,omitempty"`
	SourceResolution string `json:"source_resolution,omitempty"`
	TargetResolution string `json:"target_resolution,omitempty"`
	TargetRatio      string `json:"target_ratio,omitempty"`
	TargetWidth      int    `json:"target_width,omitempty"`
	TargetHeight     int    `json:"target_height,omitempty"`
	ResultURL        string `json:"result_url,omitempty"`
	OutputTOSURL     string `json:"output_tos_url,omitempty"`
	LastError        string `json:"last_error,omitempty"`

	NormalizeTaskID       string `json:"normalize_task_id,omitempty"`
	NormalizeStatus       string `json:"normalize_status,omitempty"`
	NormalizeResultURL    string `json:"normalize_result_url,omitempty"`
	NormalizeOutputTOSURL string `json:"normalize_output_tos_url,omitempty"`
	NormalizeLastError    string `json:"normalize_last_error,omitempty"`
}

type videoSuperResolutionConfig struct {
	Enabled           bool
	BaseURL           string
	APIKey            string
	OutputTOSPath     string
	OperatorID        string
	OperatorVersion   string
	PreserveAudio     bool
	OutputQualityMode string
	TOSPublicBaseURL  string
	TOSEndpoint       string
	TOSRegion         string
	TOSAccessKey      string
	TOSSecretKey      string
	TOSSessionToken   string
	TOSPresignExpires int64
}

type videoSuperResolutionSubmitRequest struct {
	OperatorID      string `json:"operator_id"`
	OperatorVersion string `json:"operator_version"`
	Data            struct {
		VideoURL          string `json:"video_url"`
		OutputTOSPath     string `json:"output_tos_path"`
		TargetWidth       int    `json:"target_width"`
		TargetHeight      int    `json:"target_height,omitempty"`
		PreserveAudio     bool   `json:"preserve_audio,omitempty"`
		OutputBaseName    string `json:"output_base_name,omitempty"`
		OutputQualityMode string `json:"output_quality_mode,omitempty"`
	} `json:"data"`
}

type videoSuperResolutionPollRequest struct {
	OperatorID      string `json:"operator_id"`
	OperatorVersion string `json:"operator_version"`
	TaskID          string `json:"task_id"`
}

type videoResizeSubmitRequest struct {
	OperatorID      string `json:"operator_id"`
	OperatorVersion string `json:"operator_version"`
	Data            struct {
		VideoPath                    string `json:"video_path"`
		OutputTOSDir                 string `json:"output_tos_dir"`
		OutputFileName               string `json:"output_file_name,omitempty"`
		MinWidth                     int    `json:"min_width"`
		MaxWidth                     int    `json:"max_width"`
		MinHeight                    int    `json:"min_height"`
		MaxHeight                    int    `json:"max_height"`
		ForceOriginalAspectRatioType string `json:"force_original_aspect_ratio_type,omitempty"`
		ForceDivisibleBy             int    `json:"force_divisible_by,omitempty"`
		CQ                           int    `json:"cq,omitempty"`
		RC                           string `json:"rc,omitempty"`
	} `json:"data"`
}

type videoSuperResolutionResponse struct {
	Metadata struct {
		TaskID       string `json:"task_id"`
		TaskStatus   string `json:"task_status"`
		BusinessCode string `json:"business_code"`
		ErrorMsg     string `json:"error_msg"`
	} `json:"metadata"`
	Data struct {
		OutputVideoTOSURL string `json:"output_video_tos_url"`
		OutputVideoURL    string `json:"output_video_url"`
		OutputPath        string `json:"output_path"`
		Width             any    `json:"width"`
		Height            any    `json:"height"`
	} `json:"data"`
}

type videoResolutionDimensions struct {
	Width  int
	Height int
}

var videoResolutionDimensionsByPreset = map[string]map[string]videoResolutionDimensions{
	"480p": {
		"16:9": {Width: 864, Height: 496},
		"4:3":  {Width: 752, Height: 560},
		"1:1":  {Width: 640, Height: 640},
		"3:4":  {Width: 560, Height: 752},
		"9:16": {Width: 496, Height: 864},
		"21:9": {Width: 992, Height: 432},
	},
	"720p": {
		"16:9": {Width: 1280, Height: 720},
		"4:3":  {Width: 1112, Height: 834},
		"1:1":  {Width: 960, Height: 960},
		"3:4":  {Width: 834, Height: 1112},
		"9:16": {Width: 720, Height: 1280},
		"21:9": {Width: 1470, Height: 630},
	},
	"1080p": {
		"16:9": {Width: 1920, Height: 1080},
		"4:3":  {Width: 1664, Height: 1248},
		"1:1":  {Width: 1440, Height: 1440},
		"3:4":  {Width: 1248, Height: 1664},
		"9:16": {Width: 1080, Height: 1920},
		"21:9": {Width: 2206, Height: 946},
	},
}

func getVideoSuperResolutionConfig() videoSuperResolutionConfig {
	setting := videoSuperResolutionSetting.GetVideoSuperResolutionSetting()
	return videoSuperResolutionConfig{
		Enabled:           setting.Enabled,
		BaseURL:           strings.TrimRight(setting.BaseURL, "/"),
		APIKey:            strings.TrimSpace(setting.APIKey),
		OutputTOSPath:     strings.TrimSpace(setting.OutputTOSPath),
		OperatorID:        strings.TrimSpace(setting.OperatorID),
		OperatorVersion:   strings.TrimSpace(setting.OperatorVersion),
		PreserveAudio:     setting.PreserveAudio,
		OutputQualityMode: strings.TrimSpace(setting.OutputQualityMode),
		TOSPublicBaseURL:  strings.TrimSpace(setting.TOSPublicBaseURL),
		TOSEndpoint:       strings.TrimSpace(setting.TOSEndpoint),
		TOSRegion:         strings.TrimSpace(setting.TOSRegion),
		TOSAccessKey:      strings.TrimSpace(setting.TOSAccessKey),
		TOSSecretKey:      strings.TrimSpace(setting.TOSSecretKey),
		TOSSessionToken:   strings.TrimSpace(setting.TOSSessionToken),
		TOSPresignExpires: int64(setting.TOSPresignExpires),
	}
}

func (c videoSuperResolutionConfig) ready() bool {
	return c.Enabled && c.APIKey != "" && c.OutputTOSPath != "" && c.OperatorID != "" && c.OperatorVersion != ""
}

func (c videoSuperResolutionConfig) missingRequiredFields() []string {
	missing := make([]string, 0, 5)
	if !c.Enabled {
		missing = append(missing, "enabled")
	}
	if c.APIKey == "" {
		missing = append(missing, "api_key")
	}
	if c.OutputTOSPath == "" {
		missing = append(missing, "output_tos_path")
	}
	if c.OperatorID == "" {
		missing = append(missing, "operator_id")
	}
	if c.OperatorVersion == "" {
		missing = append(missing, "operator_version")
	}
	return missing
}

func (c videoSuperResolutionConfig) shouldPresignTOS() bool {
	return c.TOSEndpoint != "" && c.TOSRegion != "" && c.TOSAccessKey != "" && c.TOSSecretKey != ""
}

func ParseVideoSuperResolutionTaskData(data []byte) (*VideoSuperResolutionTaskData, bool, error) {
	if len(data) == 0 {
		return nil, false, nil
	}
	var wrapper VideoSuperResolutionTaskData
	if err := common.Unmarshal(data, &wrapper); err != nil {
		return nil, false, nil
	}
	if wrapper.Kind != videoSuperResolutionTaskDataKind {
		return nil, false, nil
	}
	return &wrapper, true, nil
}

func ExtractOriginalVideoTaskData(data []byte) []byte {
	wrapper, ok, err := ParseVideoSuperResolutionTaskData(data)
	if err != nil || !ok || len(wrapper.Original) == 0 {
		return data
	}
	return wrapper.Original
}

func ReadVideoSuperResolutionState(data []byte) (*VideoSuperResolutionState, bool, error) {
	wrapper, ok, err := ParseVideoSuperResolutionTaskData(data)
	if err != nil || !ok || wrapper.PostProcess == nil {
		return nil, false, err
	}
	return wrapper.PostProcess, true, nil
}

func ResolveVideoSuperResolutionFetchURL(task *model.Task) (string, error) {
	if task == nil {
		return "", fmt.Errorf("task is nil")
	}
	state, ok, err := ReadVideoSuperResolutionState(task.Data)
	if err != nil {
		return "", err
	}
	if !ok || state == nil {
		return "", nil
	}
	if url := strings.TrimSpace(state.ResultURL); url != "" && !strings.HasPrefix(url, "tos://") {
		return url, nil
	}
	if strings.TrimSpace(state.OutputTOSURL) == "" {
		return "", nil
	}
	return resolveVideoSuperResolutionObjectURL(state.OutputTOSURL, getVideoSuperResolutionConfig())
}

func IsVideoSuperResolutionRequested(metadata map[string]any) bool {
	if metadata == nil {
		return false
	}
	raw, ok := metadata[videoSuperResolutionOptInKey]
	if !ok {
		return false
	}
	switch value := raw.(type) {
	case bool:
		return value
	case string:
		normalized := strings.TrimSpace(strings.ToLower(value))
		return normalized == "true" || normalized == "1" || normalized == "yes" || normalized == "on"
	case float64:
		return value == 1
	case int:
		return value == 1
	case int64:
		return value == 1
	default:
		return false
	}
}

func NormalizeVideoSuperResolutionModelAlias(modelName string) string {
	switch strings.ToLower(strings.TrimSpace(modelName)) {
	case Seedance15SRModelAlias:
		return Seedance15ModelAlias
	case Seedance2SRModelAlias:
		return Seedance2ModelAlias
	case Seedance20SRModelAlias:
		return Seedance20ModelAlias
	case Seedance20FastSRModelAlias:
		return Seedance20FastModelAlias
	case SD20FastSRModelAlias:
		return SD20FastModelAlias
	default:
		return strings.TrimSpace(modelName)
	}
}

// ResolveDedicatedVideoSuperResolutionSourceResolution converts the user-facing
// requested resolution for dedicated SR aliases into the lower source
// resolution that should be generated upstream before super resolution runs.
func ResolveDedicatedVideoSuperResolutionSourceResolution(modelName, resolution string) string {
	if !IsDedicatedVideoSuperResolutionModel(modelName) {
		return strings.TrimSpace(resolution)
	}
	switch strings.ToLower(strings.TrimSpace(resolution)) {
	case "720", "720p":
		return "480p"
	case "1080", "1080p":
		return "720p"
	default:
		return strings.TrimSpace(resolution)
	}
}

func ShouldHideVideoSuperResolutionMetadata(modelName string) bool {
	return IsDedicatedVideoSuperResolutionModel(modelName)
}

func IsDedicatedVideoSuperResolutionModel(modelName string) bool {
	switch strings.ToLower(strings.TrimSpace(modelName)) {
	case Seedance15SRModelAlias, Seedance2SRModelAlias, Seedance20SRModelAlias, Seedance20FastSRModelAlias, SD20FastSRModelAlias:
		return true
	default:
		return false
	}
}

func ShouldAutoEnableVideoSuperResolution(modelNames ...string) bool {
	for _, modelName := range modelNames {
		if IsDedicatedVideoSuperResolutionModel(modelName) {
			return true
		}
	}
	return false
}

func ShouldEnableVideoSuperResolutionForModel(modelNames ...string) bool {
	for _, modelName := range modelNames {
		enabled, err := model.IsVideoSuperResolutionModel(modelName)
		if err != nil {
			common.SysError(fmt.Sprintf("check video super resolution model metadata failed for %q: %v", modelName, err))
			continue
		}
		if enabled {
			return true
		}
	}
	return false
}

func MaybeStartVideoSuperResolution(ctx context.Context, task *model.Task, taskResult *relaycommon.TaskInfo) (bool, error) {
	cfg := getVideoSuperResolutionConfig()
	if task == nil || taskResult == nil || taskResult.Status != string(model.TaskStatusSuccess) {
		return false, nil
	}
	if !task.Properties.VideoSuperResolutionRequested {
		return false, nil
	}
	if !cfg.ready() {
		common.SysLog(fmt.Sprintf("skip video super resolution for task %s: config is incomplete (%s)", task.TaskID, strings.Join(cfg.missingRequiredFields(), ", ")))
		return false, nil
	}
	if strings.TrimSpace(taskResult.Url) == "" {
		return false, nil
	}
	if wrapper, ok, err := ParseVideoSuperResolutionTaskData(task.Data); err != nil {
		return false, err
	} else if ok && wrapper.PostProcess != nil && strings.TrimSpace(wrapper.PostProcess.TaskID) != "" {
		return false, nil
	}

	sourceResolution, targetResolution, targetRatio, targetWidth, targetHeight, ok := resolveVideoSuperResolutionTarget(task)
	if !ok {
		common.SysLog(fmt.Sprintf("skip video super resolution for task %s: unsupported or unknown source resolution", task.TaskID))
		return false, nil
	}

	state, err := submitVideoSuperResolution(ctx, cfg, task, taskResult.Url, sourceResolution, targetResolution, targetRatio, targetWidth, targetHeight)
	if err != nil {
		return false, err
	}

	if err := writeVideoSuperResolutionTaskData(task, task.Data, state); err != nil {
		return false, err
	}
	task.PrivateData.ResultURL = ""
	return true, nil
}

func PollVideoSuperResolution(ctx context.Context, task *model.Task) (*relaycommon.TaskInfo, bool, error) {
	if task == nil {
		return nil, false, nil
	}
	wrapper, ok, err := ParseVideoSuperResolutionTaskData(task.Data)
	if err != nil || !ok || wrapper.PostProcess == nil || strings.TrimSpace(wrapper.PostProcess.TaskID) == "" {
		return nil, false, err
	}

	cfg := getVideoSuperResolutionConfig()
	if !cfg.ready() {
		return nil, true, fmt.Errorf("video super resolution is enabled for task %s but config is incomplete", task.TaskID)
	}

	state := wrapper.PostProcess
	if shouldNormalizeVideoSuperResolutionOutput(state) && strings.TrimSpace(state.NormalizeTaskID) != "" {
		taskInfo, err := pollVideoResolutionNormalization(ctx, cfg, task, wrapper, state)
		if err != nil {
			return nil, true, err
		}
		if writeErr := writeVideoSuperResolutionTaskData(task, wrapper.Original, state); writeErr != nil {
			return nil, true, writeErr
		}
		return taskInfo, true, nil
	}

	respBody, err := doVideoSuperResolutionRequest(ctx, cfg, "/api/v1/poll", videoSuperResolutionPollRequest{
		OperatorID:      cfg.OperatorID,
		OperatorVersion: cfg.OperatorVersion,
		TaskID:          wrapper.PostProcess.TaskID,
	})
	if err != nil {
		return nil, true, err
	}

	var resp videoSuperResolutionResponse
	if err := common.Unmarshal(respBody, &resp); err != nil {
		return nil, true, err
	}

	if resp.Metadata.TaskID != "" {
		state.TaskID = resp.Metadata.TaskID
	}
	if resp.Metadata.TaskStatus != "" {
		state.Status = resp.Metadata.TaskStatus
	}
	state.OutputTOSURL = strings.TrimSpace(resp.Data.OutputVideoTOSURL)
	state.ResultURL = strings.TrimSpace(resp.Data.OutputVideoURL)
	state.LastError = strings.TrimSpace(resp.Metadata.ErrorMsg)

	taskInfo := &relaycommon.TaskInfo{
		TaskID:   task.TaskID,
		Progress: "85%",
		Status:   string(model.TaskStatusInProgress),
	}
	taskInfo.TotalTokens = extractVideoSuperResolutionOriginalTotalTokens(wrapper.Original)

	switch strings.ToUpper(strings.TrimSpace(resp.Metadata.TaskStatus)) {
	case "COMPLETED":
		if state.ResultURL == "" && state.OutputTOSURL != "" {
			if resolvedURL, resolveErr := resolveVideoSuperResolutionObjectURL(state.OutputTOSURL, cfg); resolveErr == nil {
				state.ResultURL = resolvedURL
			}
		}
		if state.ResultURL == "" && state.OutputTOSURL == "" {
			fallbackTaskInfoToOriginalVideo(taskInfo, state, "video super resolution completed without output url")
			break
		}
		if shouldNormalizeVideoSuperResolutionOutput(state) {
			if err := submitVideoResolutionNormalization(ctx, cfg, task, state); err != nil {
				fallbackTaskInfoToOriginalVideo(taskInfo, state, err.Error())
				break
			}
			taskInfo.Status = string(model.TaskStatusInProgress)
			taskInfo.Progress = "90%"
			taskInfo.Url = ""
			break
		}
		taskInfo.Status = string(model.TaskStatusSuccess)
		taskInfo.Progress = "100%"
		taskInfo.Url = state.ResultURL
	case "FAILED", "FAILURE":
		fallbackTaskInfoToOriginalVideo(taskInfo, state, fallbackString(state.LastError, "video super resolution failed"))
	default:
		if state.LastError != "" && resp.Metadata.BusinessCode != "" && resp.Metadata.BusinessCode != "0" {
			fallbackTaskInfoToOriginalVideo(taskInfo, state, state.LastError)
		}
	}

	if err := writeVideoSuperResolutionTaskData(task, wrapper.Original, state); err != nil {
		return nil, true, err
	}
	return taskInfo, true, nil
}

func extractVideoSuperResolutionOriginalTotalTokens(data []byte) int {
	if len(data) == 0 {
		return 0
	}
	var carrier struct {
		Usage struct {
			TotalTokens int `json:"total_tokens"`
		} `json:"usage"`
	}
	if err := common.Unmarshal(data, &carrier); err != nil {
		return 0
	}
	return carrier.Usage.TotalTokens
}

func fallbackTaskInfoToOriginalVideo(taskInfo *relaycommon.TaskInfo, state *VideoSuperResolutionState, reason string) {
	if taskInfo == nil || state == nil {
		return
	}
	state.LastError = fallbackString(reason, state.LastError)
	state.ResultURL = ""
	taskInfo.Status = string(model.TaskStatusFailure)
	taskInfo.Progress = "100%"
	taskInfo.Url = ""
	taskInfo.Reason = state.LastError
}

func shouldNormalizeVideoSuperResolutionOutput(state *VideoSuperResolutionState) bool {
	if state == nil {
		return false
	}
	if state.TargetWidth <= 0 || state.TargetHeight <= 0 {
		return false
	}
	sourceWidth, sourceHeight, ok := lookupVideoResolutionDimensions(state.SourceResolution, state.TargetRatio)
	if !ok {
		return false
	}
	return sourceWidth*state.TargetHeight != sourceHeight*state.TargetWidth
}

func submitVideoResolutionNormalization(ctx context.Context, cfg videoSuperResolutionConfig, task *model.Task, state *VideoSuperResolutionState) error {
	if strings.TrimSpace(state.NormalizeTaskID) != "" {
		return nil
	}
	inputPath := strings.TrimSpace(state.OutputTOSURL)
	if inputPath == "" {
		return fmt.Errorf("video resolution normalization requires output_tos_url")
	}
	outputDir := normalizeVideoResolutionOutputTOSDir(cfg.OutputTOSPath)
	if outputDir == "" {
		return fmt.Errorf("video resolution normalization output_tos_dir is empty")
	}

	reqBody := videoResizeSubmitRequest{
		OperatorID:      "las_video_resize",
		OperatorVersion: "v1",
	}
	reqBody.Data.VideoPath = inputPath
	reqBody.Data.OutputTOSDir = outputDir
	reqBody.Data.OutputFileName = fmt.Sprintf("%s_%dx%d.mp4", task.TaskID, state.TargetWidth, state.TargetHeight)
	reqBody.Data.MinWidth = state.TargetWidth
	reqBody.Data.MaxWidth = state.TargetWidth
	reqBody.Data.MinHeight = state.TargetHeight
	reqBody.Data.MaxHeight = state.TargetHeight
	reqBody.Data.ForceOriginalAspectRatioType = "disable"
	reqBody.Data.ForceDivisibleBy = 2
	reqBody.Data.CQ = 18
	reqBody.Data.RC = "vbr"

	respBody, err := doVideoSuperResolutionRequest(ctx, cfg, "/api/v1/submit", reqBody)
	if err != nil {
		return err
	}
	var resp videoSuperResolutionResponse
	if err := common.Unmarshal(respBody, &resp); err != nil {
		return err
	}
	if strings.TrimSpace(resp.Metadata.TaskID) == "" {
		return fmt.Errorf("video resolution normalization task_id is empty")
	}
	if msg := strings.TrimSpace(resp.Metadata.ErrorMsg); msg != "" && resp.Metadata.BusinessCode != "0" {
		return fmt.Errorf("video resolution normalization submit failed: %s", msg)
	}
	state.NormalizeTaskID = resp.Metadata.TaskID
	state.NormalizeStatus = fallbackString(resp.Metadata.TaskStatus, "PENDING")
	state.NormalizeLastError = strings.TrimSpace(resp.Metadata.ErrorMsg)
	return nil
}

func pollVideoResolutionNormalization(ctx context.Context, cfg videoSuperResolutionConfig, task *model.Task, wrapper *VideoSuperResolutionTaskData, state *VideoSuperResolutionState) (*relaycommon.TaskInfo, error) {
	respBody, err := doVideoSuperResolutionRequest(ctx, cfg, "/api/v1/poll", videoSuperResolutionPollRequest{
		OperatorID:      "las_video_resize",
		OperatorVersion: "v1",
		TaskID:          state.NormalizeTaskID,
	})
	if err != nil {
		return nil, err
	}

	var resp videoSuperResolutionResponse
	if err := common.Unmarshal(respBody, &resp); err != nil {
		return nil, err
	}

	if resp.Metadata.TaskID != "" {
		state.NormalizeTaskID = resp.Metadata.TaskID
	}
	if resp.Metadata.TaskStatus != "" {
		state.NormalizeStatus = resp.Metadata.TaskStatus
	}
	state.NormalizeLastError = strings.TrimSpace(resp.Metadata.ErrorMsg)

	taskInfo := &relaycommon.TaskInfo{
		TaskID:   task.TaskID,
		Progress: "90%",
		Status:   string(model.TaskStatusInProgress),
	}
	taskInfo.TotalTokens = extractVideoSuperResolutionOriginalTotalTokens(wrapper.Original)

	switch strings.ToUpper(strings.TrimSpace(resp.Metadata.TaskStatus)) {
	case "COMPLETED":
		outputPath := strings.TrimSpace(resp.Data.OutputPath)
		if outputPath == "" {
			outputPath = strings.TrimSpace(resp.Data.OutputVideoTOSURL)
		}
		if outputPath == "" {
			fallbackTaskInfoToOriginalVideo(taskInfo, state, "video resolution normalization completed without output path")
			break
		}
		state.NormalizeOutputTOSURL = outputPath
		resolvedURL, resolveErr := resolveVideoSuperResolutionObjectURL(outputPath, cfg)
		if resolveErr != nil {
			fallbackTaskInfoToOriginalVideo(taskInfo, state, resolveErr.Error())
			break
		}
		state.NormalizeResultURL = resolvedURL
		state.ResultURL = resolvedURL
		state.OutputTOSURL = outputPath
		taskInfo.Status = string(model.TaskStatusSuccess)
		taskInfo.Progress = "100%"
		taskInfo.Url = resolvedURL
	case "FAILED", "FAILURE":
		fallbackTaskInfoToOriginalVideo(taskInfo, state, fallbackString(state.NormalizeLastError, "video resolution normalization failed"))
	default:
		if state.NormalizeLastError != "" && resp.Metadata.BusinessCode != "" && resp.Metadata.BusinessCode != "0" {
			fallbackTaskInfoToOriginalVideo(taskInfo, state, state.NormalizeLastError)
		}
	}
	return taskInfo, nil
}

func normalizeVideoResolutionOutputTOSDir(outputTOSPath string) string {
	outputTOSPath = strings.TrimSpace(outputTOSPath)
	if outputTOSPath == "" {
		return ""
	}
	if strings.HasPrefix(outputTOSPath, "tos://") {
		return strings.TrimRight(outputTOSPath, "/") + "/normalized"
	}
	return strings.TrimRight(outputTOSPath, "/") + "/normalized"
}

func writeVideoSuperResolutionTaskData(task *model.Task, original []byte, state *VideoSuperResolutionState) error {
	wrapper := VideoSuperResolutionTaskData{
		Kind:        videoSuperResolutionTaskDataKind,
		Original:    json.RawMessage(original),
		PostProcess: state,
	}
	data, err := common.Marshal(wrapper)
	if err != nil {
		return err
	}
	task.Data = data
	return nil
}

func submitVideoSuperResolution(ctx context.Context, cfg videoSuperResolutionConfig, task *model.Task, sourceURL, sourceResolution, targetResolution, targetRatio string, targetWidth, targetHeight int) (*VideoSuperResolutionState, error) {
	reqBody := videoSuperResolutionSubmitRequest{
		OperatorID:      cfg.OperatorID,
		OperatorVersion: cfg.OperatorVersion,
	}
	reqBody.Data.VideoURL = sourceURL
	reqBody.Data.OutputTOSPath = cfg.OutputTOSPath
	reqBody.Data.TargetWidth = targetWidth
	reqBody.Data.TargetHeight = targetHeight
	reqBody.Data.PreserveAudio = cfg.PreserveAudio
	reqBody.Data.OutputBaseName = task.TaskID
	reqBody.Data.OutputQualityMode = fallbackString(cfg.OutputQualityMode, "balanced")

	respBody, err := doVideoSuperResolutionRequest(ctx, cfg, "/api/v1/submit", reqBody)
	if err != nil {
		return nil, err
	}

	var resp videoSuperResolutionResponse
	if err := common.Unmarshal(respBody, &resp); err != nil {
		return nil, err
	}
	if strings.TrimSpace(resp.Metadata.TaskID) == "" {
		return nil, fmt.Errorf("video super resolution task_id is empty")
	}
	if msg := strings.TrimSpace(resp.Metadata.ErrorMsg); msg != "" && resp.Metadata.BusinessCode != "0" {
		return nil, fmt.Errorf("video super resolution submit failed: %s", msg)
	}

	return &VideoSuperResolutionState{
		Type:             videoSuperResolutionTypeLAS,
		TaskID:           resp.Metadata.TaskID,
		Status:           fallbackString(resp.Metadata.TaskStatus, "PENDING"),
		SourceURL:        sourceURL,
		SourceResolution: sourceResolution,
		TargetResolution: targetResolution,
		TargetRatio:      targetRatio,
		TargetWidth:      targetWidth,
		TargetHeight:     targetHeight,
	}, nil
}

func doVideoSuperResolutionRequest(ctx context.Context, cfg videoSuperResolutionConfig, path string, payload any) ([]byte, error) {
	data, err := common.Marshal(payload)
	if err != nil {
		return nil, err
	}
	req, err := http.NewRequestWithContext(ctx, http.MethodPost, cfg.BaseURL+path, bytes.NewReader(data))
	if err != nil {
		return nil, err
	}
	req.Header.Set("Authorization", "Bearer "+cfg.APIKey)
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Accept", "application/json")

	client, err := GetHttpClientWithProxy("")
	if err != nil {
		return nil, err
	}
	resp, err := client.Do(req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, err
	}
	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("video super resolution upstream status %d: %s", resp.StatusCode, string(body))
	}
	return body, nil
}

type tosObjectRef struct {
	Bucket string
	Key    string
}

func resolveVideoSuperResolutionObjectURL(raw string, cfg videoSuperResolutionConfig) (string, error) {
	raw = strings.TrimSpace(raw)
	if raw == "" {
		return "", fmt.Errorf("tos url is empty")
	}
	if strings.HasPrefix(raw, "http://") || strings.HasPrefix(raw, "https://") {
		return raw, nil
	}
	ref, err := parseTOSObjectURL(raw)
	if err != nil {
		return "", err
	}
	if cfg.shouldPresignTOS() {
		return presignTOSObjectURL(ref, cfg)
	}
	if cfg.TOSPublicBaseURL != "" {
		return buildPublicTOSObjectURL(ref, cfg.TOSPublicBaseURL)
	}
	if cfg.TOSEndpoint != "" {
		return buildTOSEndpointURL(ref, cfg.TOSEndpoint), nil
	}
	return "", fmt.Errorf("unable to resolve TOS url %s: missing TOS endpoint or credentials", raw)
}

func parseTOSObjectURL(raw string) (*tosObjectRef, error) {
	parsed, err := neturl.Parse(raw)
	if err != nil {
		return nil, err
	}
	if !strings.EqualFold(parsed.Scheme, "tos") {
		return nil, fmt.Errorf("unsupported tos scheme: %s", parsed.Scheme)
	}
	bucket := strings.TrimSpace(parsed.Host)
	key := strings.TrimPrefix(parsed.Path, "/")
	key = strings.TrimSpace(key)
	if bucket == "" || key == "" {
		return nil, fmt.Errorf("invalid tos url: %s", raw)
	}
	return &tosObjectRef{
		Bucket: bucket,
		Key:    key,
	}, nil
}

func presignTOSObjectURL(ref *tosObjectRef, cfg videoSuperResolutionConfig) (string, error) {
	creds := tos.NewStaticCredentials(cfg.TOSAccessKey, cfg.TOSSecretKey)
	if cfg.TOSSessionToken != "" {
		creds.WithSecurityToken(cfg.TOSSessionToken)
	}
	client, err := tos.NewClientV2(
		cfg.TOSEndpoint,
		tos.WithRegion(cfg.TOSRegion),
		tos.WithCredentials(creds),
	)
	if err != nil {
		return "", err
	}
	expires := cfg.TOSPresignExpires
	if expires <= 0 {
		expires = 3600
	}
	if expires > 604800 {
		expires = 604800
	}
	output, err := client.PreSignedURL(&tos.PreSignedURLInput{
		HTTPMethod: enum.HttpMethodGet,
		Bucket:     ref.Bucket,
		Key:        ref.Key,
		Expires:    expires,
	})
	if err != nil {
		return "", err
	}
	return output.SignedUrl, nil
}

func buildPublicTOSObjectURL(ref *tosObjectRef, base string) (string, error) {
	resolved := strings.TrimSpace(base)
	if resolved == "" {
		return "", fmt.Errorf("public TOS base url is empty")
	}
	resolved = strings.ReplaceAll(resolved, "{bucket}", ref.Bucket)
	if strings.Contains(resolved, "{key}") {
		return strings.ReplaceAll(resolved, "{key}", ref.Key), nil
	}
	parsed, err := neturl.Parse(resolved)
	if err != nil {
		return "", err
	}
	parsed.Path = path.Join(parsed.Path, ref.Key)
	return parsed.String(), nil
}

func buildTOSEndpointURL(ref *tosObjectRef, endpoint string) string {
	endpoint = strings.TrimSpace(endpoint)
	if endpoint == "" {
		return ""
	}
	if !strings.Contains(endpoint, "://") {
		endpoint = "https://" + endpoint
	}
	parsed, err := neturl.Parse(endpoint)
	if err != nil {
		return endpoint
	}
	host := strings.TrimPrefix(parsed.Host, ".")
	parsed.Host = ref.Bucket + "." + host
	parsed.Path = path.Join(parsed.Path, ref.Key)
	return parsed.String()
}

func resolveVideoSuperResolutionTarget(task *model.Task) (string, string, string, int, int, bool) {
	type resolutionCarrier struct {
		Resolution string `json:"resolution"`
		Ratio      string `json:"ratio"`
	}

	original := ExtractOriginalVideoTaskData(task.Data)
	carrier := resolutionCarrier{}
	_ = common.Unmarshal(original, &carrier)
	resolution := strings.ToLower(strings.TrimSpace(carrier.Resolution))
	if resolution == "" {
		resolution = inferVideoResolutionFromModel(task.Properties.OriginModelName, task.Properties.UpstreamModelName)
	}
	ratio := normalizeVideoAspectRatio(carrier.Ratio)

	var targetResolution string
	switch resolution {
	case "480p", "480":
		targetResolution = "720p"
	case "720p", "720":
		targetResolution = "1080p"
	default:
		return "", "", "", 0, 0, false
	}

	targetWidth, targetHeight, ok := lookupVideoResolutionDimensions(targetResolution, ratio)
	if !ok {
		return "", "", "", 0, 0, false
	}
	return normalizeVideoResolutionPreset(resolution), targetResolution, ratio, targetWidth, targetHeight, true
}

func inferVideoResolutionFromModel(modelNames ...string) string {
	for _, modelName := range modelNames {
		switch strings.ToLower(strings.TrimSpace(modelName)) {
		case Seedance15SRModelAlias, Seedance2SRModelAlias, Seedance20SRModelAlias, Seedance20FastSRModelAlias, SD20FastSRModelAlias:
			return "480p"
		case Seedance15ModelAlias,
			Seedance2ModelAlias, Seedance20ModelAlias,
			Seedance20FastModelAlias, SD20FastModelAlias,
			seedance2UpstreamModel, sd20FastUpstreamModel:
			return "720p"
		}
	}
	return ""
}

func normalizeVideoResolutionPreset(resolution string) string {
	switch strings.ToLower(strings.TrimSpace(resolution)) {
	case "480", "480p":
		return "480p"
	case "720", "720p":
		return "720p"
	case "1080", "1080p":
		return "1080p"
	default:
		return strings.ToLower(strings.TrimSpace(resolution))
	}
}

func normalizeVideoAspectRatio(ratio string) string {
	ratio = strings.TrimSpace(ratio)
	if ratio == "" {
		return "16:9"
	}
	return ratio
}

func lookupVideoResolutionDimensions(resolution, ratio string) (int, int, bool) {
	resolution = normalizeVideoResolutionPreset(resolution)
	ratio = normalizeVideoAspectRatio(ratio)
	dimensionsByRatio, ok := videoResolutionDimensionsByPreset[resolution]
	if !ok {
		return 0, 0, false
	}
	dimensions, ok := dimensionsByRatio[ratio]
	if !ok {
		return 0, 0, false
	}
	return dimensions.Width, dimensions.Height, true
}

func fallbackString(value, fallback string) string {
	value = strings.TrimSpace(value)
	if value == "" {
		return fallback
	}
	return value
}
