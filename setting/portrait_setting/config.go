package portrait_setting

import (
	"os"
	"strings"

	"github.com/QuantumNous/new-api/common"
	"github.com/QuantumNous/new-api/setting/config"
)

type OfficialPortraitAssetSetting struct {
	AccessKey       string `json:"access_key"`
	SecretKey       string `json:"secret_key"`
	ProjectName     string `json:"project_name"`
	Region          string `json:"region"`
	CallbackBaseURL string `json:"callback_base_url"`
}

var officialPortraitAssetSetting = OfficialPortraitAssetSetting{
	ProjectName:     strings.TrimSpace(common.GetEnvOrDefaultString("VOLC_PORTRAIT_PROJECT_NAME", "default")),
	Region:          strings.TrimSpace(common.GetEnvOrDefaultString("VOLC_PORTRAIT_REGION", "cn-beijing")),
	CallbackBaseURL: strings.TrimSpace(os.Getenv("PORTRAIT_OFFICIAL_CALLBACK_BASE_URL")),
}

func init() {
	config.GlobalConfig.Register("portrait_asset", &officialPortraitAssetSetting)
}

func GetOfficialPortraitAssetSetting() *OfficialPortraitAssetSetting {
	return &officialPortraitAssetSetting
}

func GetAccessKey() string {
	if value := strings.TrimSpace(officialPortraitAssetSetting.AccessKey); value != "" {
		return value
	}
	if value := strings.TrimSpace(os.Getenv("VOLC_PORTRAIT_AK")); value != "" {
		return value
	}
	return strings.TrimSpace(os.Getenv("VOLC_ACCESS_KEY_ID"))
}

func GetSecretKey() string {
	if value := strings.TrimSpace(officialPortraitAssetSetting.SecretKey); value != "" {
		return value
	}
	if value := strings.TrimSpace(os.Getenv("VOLC_PORTRAIT_SK")); value != "" {
		return value
	}
	return strings.TrimSpace(os.Getenv("VOLC_SECRET_ACCESS_KEY"))
}

func GetProjectName() string {
	if value := strings.TrimSpace(officialPortraitAssetSetting.ProjectName); value != "" {
		return value
	}
	return strings.TrimSpace(common.GetEnvOrDefaultString("VOLC_PORTRAIT_PROJECT_NAME", "default"))
}

func GetRegion() string {
	if value := strings.TrimSpace(officialPortraitAssetSetting.Region); value != "" {
		return value
	}
	return strings.TrimSpace(common.GetEnvOrDefaultString("VOLC_PORTRAIT_REGION", "cn-beijing"))
}

func GetCallbackBaseURL() string {
	if value := strings.TrimSpace(officialPortraitAssetSetting.CallbackBaseURL); value != "" {
		return value
	}
	return strings.TrimSpace(os.Getenv("PORTRAIT_OFFICIAL_CALLBACK_BASE_URL"))
}
