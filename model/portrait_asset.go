package model

import (
	"errors"
	"strings"

	"github.com/QuantumNous/new-api/common"
	"gorm.io/gorm"
)

const (
	PortraitAssetStatusPending       = "pending"
	PortraitAssetStatusQRReady       = "qr_ready"
	PortraitAssetStatusWaitingUpload = "waiting_upload"
	PortraitAssetStatusWaitingAccept = "waiting_accept"
	PortraitAssetStatusReady         = "ready"
	PortraitAssetStatusFailed        = "failed"
	PortraitAssetStatusDisabled      = "disabled"
)

var ErrPortraitAssetNotReady = errors.New("portrait asset is not ready")

type PortraitAssetJob struct {
	Id           int            `json:"id"`
	UserId       int            `json:"user_id" gorm:"index"`
	Name         string         `json:"name" gorm:"type:varchar(128)"`
	Status       string         `json:"status" gorm:"type:varchar(32);index;default:pending"`
	InviteURL    string         `json:"invite_url" gorm:"type:text"`
	QRImage      string         `json:"qr_image" gorm:"type:text"`
	VolcGroupID  string         `json:"volc_group_id" gorm:"type:varchar(128);index"`
	AssetID      string         `json:"asset_id" gorm:"type:varchar(128);index"`
	ErrorMessage string         `json:"error_message" gorm:"type:text"`
	CreatedTime  int64          `json:"created_time" gorm:"bigint"`
	UpdatedTime  int64          `json:"updated_time" gorm:"bigint"`
	ReadyTime    int64          `json:"ready_time" gorm:"bigint"`
	DeletedAt    gorm.DeletedAt `gorm:"index"`
}

func NormalizePortraitAssetID(assetID string) string {
	assetID = strings.TrimSpace(assetID)
	assetID = strings.TrimPrefix(assetID, "asset://")
	return strings.TrimSpace(assetID)
}

func PortraitAssetURI(assetID string) string {
	assetID = NormalizePortraitAssetID(assetID)
	if assetID == "" {
		return ""
	}
	return "asset://" + assetID
}

func CreatePortraitAssetJob(userId int, name string) (*PortraitAssetJob, error) {
	name = strings.TrimSpace(name)
	if name == "" {
		name = "真人资产"
	}
	job := &PortraitAssetJob{
		UserId:      userId,
		Name:        name,
		Status:      PortraitAssetStatusPending,
		CreatedTime: common.GetTimestamp(),
		UpdatedTime: common.GetTimestamp(),
	}
	return job, DB.Create(job).Error
}

func GetUserPortraitAssetJobs(userId int, startIdx int, num int) ([]*PortraitAssetJob, error) {
	var jobs []*PortraitAssetJob
	err := DB.Where("user_id = ?", userId).Order("id desc").Limit(num).Offset(startIdx).Find(&jobs).Error
	return jobs, err
}

func CountUserPortraitAssetJobs(userId int) (int64, error) {
	var count int64
	err := DB.Model(&PortraitAssetJob{}).Where("user_id = ?", userId).Count(&count).Error
	return count, err
}

func GetUserPortraitAssetJobByID(userId int, id int) (*PortraitAssetJob, error) {
	var job PortraitAssetJob
	err := DB.Where("user_id = ? and id = ?", userId, id).First(&job).Error
	return &job, err
}

func GetReadyUserPortraitAssetJob(userId int, id int) (*PortraitAssetJob, error) {
	job, err := GetUserPortraitAssetJobByID(userId, id)
	if err != nil {
		return nil, err
	}
	if job.Status != PortraitAssetStatusReady || NormalizePortraitAssetID(job.AssetID) == "" {
		return nil, ErrPortraitAssetNotReady
	}
	return job, nil
}

func GetReadyUserPortraitAssetByAssetID(userId int, assetID string) (*PortraitAssetJob, error) {
	assetID = NormalizePortraitAssetID(assetID)
	if assetID == "" {
		return nil, gorm.ErrRecordNotFound
	}
	var job PortraitAssetJob
	err := DB.Where("user_id = ? and asset_id = ? and status = ?", userId, assetID, PortraitAssetStatusReady).First(&job).Error
	return &job, err
}

func ListPortraitAssetJobsForRPA(status string, limit int) ([]*PortraitAssetJob, error) {
	if limit <= 0 || limit > 50 {
		limit = 10
	}
	query := DB.Model(&PortraitAssetJob{})
	if strings.TrimSpace(status) != "" {
		query = query.Where("status = ?", strings.TrimSpace(status))
	} else {
		query = query.Where("status in ?", []string{
			PortraitAssetStatusPending,
			PortraitAssetStatusQRReady,
			PortraitAssetStatusWaitingUpload,
			PortraitAssetStatusWaitingAccept,
		})
	}
	var jobs []*PortraitAssetJob
	err := query.Order("updated_time asc, id asc").Limit(limit).Find(&jobs).Error
	return jobs, err
}

func UpdatePortraitAssetJobFromRPA(job *PortraitAssetJob) error {
	if job == nil {
		return errors.New("portrait asset job is nil")
	}
	job.AssetID = NormalizePortraitAssetID(job.AssetID)
	job.UpdatedTime = common.GetTimestamp()
	if job.Status == PortraitAssetStatusReady && job.ReadyTime == 0 {
		job.ReadyTime = common.GetTimestamp()
	}
	return DB.Save(job).Error
}
