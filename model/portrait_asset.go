package model

import (
	"errors"
	"strings"

	"github.com/QuantumNous/new-api/common"
	"gorm.io/gorm"
)

const (
	PortraitAssetStatusPending        = "pending"
	PortraitAssetStatusQRReady        = "qr_ready"
	PortraitAssetStatusWaitingUpload  = "waiting_upload"
	PortraitAssetStatusWaitingAccept  = "waiting_accept"
	PortraitAssetStatusPendingConfirm = "pending_confirm"
	PortraitAssetStatusReady          = "ready"
	PortraitAssetStatusFailed         = "failed"
	PortraitAssetStatusDisabled       = "disabled"
	PortraitAssetStatusExpired        = "expired"
)

const PortraitAssetQRCodeLeaseSeconds int64 = 180

var (
	ErrPortraitAssetNotReady      = errors.New("portrait asset is not ready")
	ErrPortraitAssetActiveJob     = errors.New("portrait asset job already active")
	ErrPortraitAssetQRCodeExpired = errors.New("portrait asset qr code expired")
	ErrPortraitAssetNoPublicQR    = errors.New("portrait asset public qr code is not configured")
)

type PortraitAssetJob struct {
	Id            int            `json:"id"`
	UserId        int            `json:"user_id" gorm:"index"`
	Name          string         `json:"name" gorm:"type:varchar(128)"`
	Status        string         `json:"status" gorm:"type:varchar(32);index;default:pending"`
	InviteURL     string         `json:"invite_url" gorm:"type:text"`
	QRImage       string         `json:"qr_image" gorm:"type:text"`
	VolcGroupID   string         `json:"volc_group_id" gorm:"type:varchar(128);index"`
	AssetID       string         `json:"asset_id" gorm:"type:varchar(128);index"`
	AssetStatus   string         `json:"asset_status" gorm:"type:varchar(64)"`
	AssetPreview  string         `json:"asset_preview" gorm:"type:text"`
	ErrorMessage  string         `json:"error_message" gorm:"type:text"`
	CreatedTime   int64          `json:"created_time" gorm:"bigint"`
	UpdatedTime   int64          `json:"updated_time" gorm:"bigint"`
	AcceptTime    int64          `json:"accept_time" gorm:"bigint;index"`
	QRExpireTime  int64          `json:"qr_expires_time" gorm:"bigint;index"`
	ReadyTime     int64          `json:"ready_time" gorm:"bigint"`
	QueuePosition int64          `json:"queue_position" gorm:"-"`
	DeletedAt     gorm.DeletedAt `gorm:"index"`
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
	if err := AdvancePortraitAssetQueue(); err != nil {
		return nil, err
	}
	active, err := HasActivePortraitAssetJob(userId)
	if err != nil {
		return nil, err
	}
	if active {
		return nil, ErrPortraitAssetActiveJob
	}
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
	if err := DB.Create(job).Error; err != nil {
		return job, err
	}
	return job, AdvancePortraitAssetQueue()
}

func GetUserPortraitAssetJobs(userId int, startIdx int, num int) ([]*PortraitAssetJob, error) {
	if err := AdvancePortraitAssetQueue(); err != nil {
		return nil, err
	}
	var jobs []*PortraitAssetJob
	err := DB.Where("user_id = ?", userId).Order("id desc").Limit(num).Offset(startIdx).Find(&jobs).Error
	if err == nil {
		fillPortraitAssetQueuePositions(jobs)
	}
	return jobs, err
}

func CountUserPortraitAssetJobs(userId int) (int64, error) {
	var count int64
	err := DB.Model(&PortraitAssetJob{}).Where("user_id = ?", userId).Count(&count).Error
	return count, err
}

func GetUserPortraitAssetJobByID(userId int, id int) (*PortraitAssetJob, error) {
	if err := AdvancePortraitAssetQueue(); err != nil {
		return nil, err
	}
	var job PortraitAssetJob
	err := DB.Where("user_id = ? and id = ?", userId, id).First(&job).Error
	if err == nil {
		fillPortraitAssetQueuePositions([]*PortraitAssetJob{&job})
	}
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

func RequestUserPortraitAssetAccept(userId int, id int) (*PortraitAssetJob, error) {
	job, err := GetUserPortraitAssetJobByID(userId, id)
	if err != nil {
		return nil, err
	}
	now := common.GetTimestamp()
	switch job.Status {
	case PortraitAssetStatusReady:
		return job, nil
	case PortraitAssetStatusQRReady:
		if job.QRExpireTime > 0 && job.QRExpireTime <= now {
			job.Status = PortraitAssetStatusExpired
			job.ErrorMessage = "二维码已超时，请重新排队"
			job.UpdatedTime = now
			_ = DB.Save(job).Error
			_ = AdvancePortraitAssetQueue()
			return nil, ErrPortraitAssetQRCodeExpired
		}
		job.Status = PortraitAssetStatusWaitingAccept
		job.AssetID = ""
		job.AssetStatus = ""
		job.AssetPreview = ""
		job.ErrorMessage = ""
		job.AcceptTime = now
		job.UpdatedTime = now
		if err := DB.Save(job).Error; err != nil {
			return nil, err
		}
		return job, AdvancePortraitAssetQueue()
	case PortraitAssetStatusWaitingUpload, PortraitAssetStatusWaitingAccept:
		job.Status = PortraitAssetStatusWaitingAccept
		job.ErrorMessage = ""
		job.AcceptTime = now
		job.UpdatedTime = now
		if err := DB.Save(job).Error; err != nil {
			return nil, err
		}
		return job, AdvancePortraitAssetQueue()
	default:
		return nil, ErrPortraitAssetNotReady
	}
}

func ConfirmUserPortraitAsset(userId int, id int) (*PortraitAssetJob, error) {
	job, err := GetUserPortraitAssetJobByID(userId, id)
	if err != nil {
		return nil, err
	}
	if job.Status == PortraitAssetStatusReady {
		return job, nil
	}
	if job.Status != PortraitAssetStatusPendingConfirm || NormalizePortraitAssetID(job.AssetID) == "" {
		return nil, ErrPortraitAssetNotReady
	}
	now := common.GetTimestamp()
	job.Status = PortraitAssetStatusReady
	job.ErrorMessage = ""
	job.UpdatedTime = now
	if job.ReadyTime == 0 {
		job.ReadyTime = now
	}
	return job, DB.Save(job).Error
}

func RejectUserPortraitAsset(userId int, id int) (*PortraitAssetJob, error) {
	job, err := GetUserPortraitAssetJobByID(userId, id)
	if err != nil {
		return nil, err
	}
	if job.Status != PortraitAssetStatusPendingConfirm {
		return nil, ErrPortraitAssetNotReady
	}
	now := common.GetTimestamp()
	job.Status = PortraitAssetStatusFailed
	job.AssetID = ""
	job.ErrorMessage = "用户确认缩略图不匹配"
	job.UpdatedTime = now
	return job, DB.Save(job).Error
}

func HasActivePortraitAssetJob(userId int) (bool, error) {
	var count int64
	err := DB.Model(&PortraitAssetJob{}).Where("user_id = ? and status in ?", userId, []string{
		PortraitAssetStatusPending,
		PortraitAssetStatusQRReady,
		PortraitAssetStatusWaitingUpload,
		PortraitAssetStatusWaitingAccept,
		PortraitAssetStatusPendingConfirm,
	}).Count(&count).Error
	return count > 0, err
}

func AdvancePortraitAssetQueue() error {
	now := common.GetTimestamp()
	err := DB.Model(&PortraitAssetJob{}).Where("status = ? and qr_expire_time > 0 and qr_expire_time <= ?", PortraitAssetStatusQRReady, now).Updates(map[string]any{
		"status":        PortraitAssetStatusExpired,
		"error_message": "二维码已超时，请重新排队",
		"updated_time":  now,
	}).Error
	if err != nil {
		return err
	}

	var activeCount int64
	err = DB.Model(&PortraitAssetJob{}).Where("status = ? and qr_expire_time > ?", PortraitAssetStatusQRReady, now).Count(&activeCount).Error
	if err != nil || activeCount > 0 {
		return err
	}

	var next PortraitAssetJob
	err = DB.Where("status = ?", PortraitAssetStatusPending).Order("created_time asc, id asc").First(&next).Error
	if errors.Is(err, gorm.ErrRecordNotFound) {
		return nil
	}
	if err != nil {
		return err
	}

	invite, err := GetPortraitAssetPublicInvite()
	if errors.Is(err, gorm.ErrRecordNotFound) {
		next.ErrorMessage = "等待管理员配置真人资产二维码"
		next.UpdatedTime = now
		return DB.Save(&next).Error
	}
	if err != nil {
		return err
	}

	next.Status = PortraitAssetStatusQRReady
	next.InviteURL = invite.InviteURL
	next.QRImage = invite.QRImage
	next.ErrorMessage = ""
	next.QRExpireTime = now + PortraitAssetQRCodeLeaseSeconds
	next.UpdatedTime = now
	return DB.Save(&next).Error
}

func GetPortraitAssetPublicInvite() (*PortraitAssetJob, error) {
	var invite PortraitAssetJob
	err := DB.Where("invite_url <> '' and qr_image <> ''").Order("updated_time desc, id desc").First(&invite).Error
	return &invite, err
}

func fillPortraitAssetQueuePositions(jobs []*PortraitAssetJob) {
	for _, job := range jobs {
		if job == nil || job.Status != PortraitAssetStatusPending {
			continue
		}
		var position int64
		err := DB.Model(&PortraitAssetJob{}).Where(
			"status = ? and (created_time < ? or (created_time = ? and id <= ?))",
			PortraitAssetStatusPending,
			job.CreatedTime,
			job.CreatedTime,
			job.Id,
		).Count(&position).Error
		if err == nil {
			job.QueuePosition = position
		}
	}
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
