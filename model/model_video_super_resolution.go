package model

import (
	"errors"
	"strings"

	"gorm.io/gorm"
)

func IsVideoSuperResolutionModel(modelName string) (bool, error) {
	modelName = strings.TrimSpace(modelName)
	if modelName == "" {
		return false, nil
	}

	var modelMeta Model
	err := DB.
		Where("model_name = ?", modelName).
		Where("video_super_resolution_enabled = ?", true).
		First(&modelMeta).Error
	if err == nil {
		return true, nil
	}
	if errors.Is(err, gorm.ErrRecordNotFound) {
		return false, nil
	}
	return false, err
}
