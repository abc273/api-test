package common

import (
	"os"
	"path/filepath"
)

const PortraitAssetUploadRoutePrefix = "/portrait-asset-uploads"

func GetPortraitAssetUploadRoot() string {
	if info, err := os.Stat("/data"); err == nil && info.IsDir() {
		return filepath.Join("/data", "portrait-asset-uploads")
	}
	return filepath.Join("data", "portrait-asset-uploads")
}
