package video_super_resolution_setting

import (
	"strings"

	"github.com/QuantumNous/new-api/common"
	"github.com/QuantumNous/new-api/setting/config"
)

type VideoSuperResolutionSetting struct {
	Enabled           bool   `json:"enabled"`
	BaseURL           string `json:"base_url"`
	APIKey            string `json:"api_key"`
	OutputTOSPath     string `json:"output_tos_path"`
	OperatorID        string `json:"operator_id"`
	OperatorVersion   string `json:"operator_version"`
	PreserveAudio     bool   `json:"preserve_audio"`
	OutputQualityMode string `json:"output_quality_mode"`
	TOSPublicBaseURL  string `json:"tos_public_base_url"`
	TOSEndpoint       string `json:"tos_endpoint"`
	TOSRegion         string `json:"tos_region"`
	TOSAccessKey      string `json:"tos_access_key"`
	TOSSecretKey      string `json:"tos_secret_key"`
	TOSSessionToken   string `json:"tos_session_token"`
	TOSPresignExpires int    `json:"tos_presign_expires"`
}

var videoSuperResolutionSetting = VideoSuperResolutionSetting{
	Enabled:           common.GetEnvOrDefaultBool("VIDEO_SUPER_RESOLUTION_ENABLED", false),
	BaseURL:           strings.TrimRight(common.GetEnvOrDefaultString("VIDEO_SUPER_RESOLUTION_BASE_URL", "https://operator.las.cn-beijing.volces.com"), "/"),
	APIKey:            strings.TrimSpace(common.GetEnvOrDefaultString("VIDEO_SUPER_RESOLUTION_API_KEY", "")),
	OutputTOSPath:     strings.TrimSpace(common.GetEnvOrDefaultString("VIDEO_SUPER_RESOLUTION_OUTPUT_TOS_PATH", "")),
	OperatorID:        strings.TrimSpace(common.GetEnvOrDefaultString("VIDEO_SUPER_RESOLUTION_OPERATOR_ID", "las_video_super_resolution")),
	OperatorVersion:   strings.TrimSpace(common.GetEnvOrDefaultString("VIDEO_SUPER_RESOLUTION_OPERATOR_VERSION", "v1")),
	PreserveAudio:     common.GetEnvOrDefaultBool("VIDEO_SUPER_RESOLUTION_PRESERVE_AUDIO", true),
	OutputQualityMode: strings.TrimSpace(common.GetEnvOrDefaultString("VIDEO_SUPER_RESOLUTION_OUTPUT_QUALITY_MODE", "balanced")),
	TOSPublicBaseURL:  strings.TrimSpace(common.GetEnvOrDefaultString("VIDEO_SUPER_RESOLUTION_TOS_PUBLIC_BASE_URL", "")),
	TOSEndpoint:       strings.TrimSpace(common.GetEnvOrDefaultString("VIDEO_SUPER_RESOLUTION_TOS_ENDPOINT", "")),
	TOSRegion:         strings.TrimSpace(common.GetEnvOrDefaultString("VIDEO_SUPER_RESOLUTION_TOS_REGION", "")),
	TOSAccessKey:      strings.TrimSpace(common.GetEnvOrDefaultString("VIDEO_SUPER_RESOLUTION_TOS_ACCESS_KEY", "")),
	TOSSecretKey:      strings.TrimSpace(common.GetEnvOrDefaultString("VIDEO_SUPER_RESOLUTION_TOS_SECRET_KEY", "")),
	TOSSessionToken:   strings.TrimSpace(common.GetEnvOrDefaultString("VIDEO_SUPER_RESOLUTION_TOS_SESSION_TOKEN", "")),
	TOSPresignExpires: common.GetEnvOrDefault("VIDEO_SUPER_RESOLUTION_TOS_PRESIGN_EXPIRES", 3600),
}

func init() {
	config.GlobalConfig.Register("video_super_resolution", &videoSuperResolutionSetting)
}

func GetVideoSuperResolutionSetting() *VideoSuperResolutionSetting {
	return &videoSuperResolutionSetting
}
