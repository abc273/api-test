package controller

import (
	"fmt"
	"strconv"
	"strings"

	"github.com/QuantumNous/new-api/common"
	"github.com/QuantumNous/new-api/model"
	"github.com/QuantumNous/new-api/service"
	"github.com/gin-gonic/gin"
)

type portraitAssetOwnerInfo struct {
	UserID   int    `json:"user_id"`
	Username string `json:"username,omitempty"`
	Email    string `json:"email,omitempty"`
}

type adminOfficialPortraitAssetItem struct {
	*model.PortraitAssetJob
	Owner portraitAssetOwnerInfo `json:"owner"`
}

type adminVirtualPortraitAssetItem struct {
	*model.VirtualPortraitAsset
	Owner portraitAssetOwnerInfo `json:"owner"`
}

func AdminListOfficialPortraitAssetJobs(c *gin.Context) {
	pageInfo := common.GetPageQuery(c)
	queryParams, err := getAdminPortraitAssetQuery(c)
	if err != nil {
		common.ApiErrorMsg(c, err.Error())
		return
	}
	jobs, err := model.GetAdminOfficialPortraitAssetJobs(queryParams.official, pageInfo.GetStartIdx(), pageInfo.GetPageSize())
	if err != nil {
		common.ApiError(c, err)
		return
	}
	decorateOfficialPortraitAssetJobsResponse(c, jobs)
	total, err := model.CountAdminOfficialPortraitAssetJobs(queryParams.official)
	if err != nil {
		common.ApiError(c, err)
		return
	}
	pageInfo.SetTotal(int(total))
	pageInfo.SetItems(buildAdminOfficialPortraitAssetItems(jobs))
	common.ApiSuccess(c, pageInfo)
}

func AdminSyncOfficialPortraitAssetJob(c *gin.Context) {
	id, err := strconv.Atoi(c.Param("id"))
	if err != nil {
		common.ApiError(c, err)
		return
	}
	job, err := model.GetOfficialPortraitAssetJobByID(id)
	if err != nil {
		common.ApiError(c, err)
		return
	}
	if err := syncOfficialPortraitAssetJob(job); err != nil {
		common.ApiError(c, err)
		return
	}
	decorateOfficialPortraitAssetJobResponse(c, job)
	common.ApiSuccess(c, adminOfficialPortraitAssetItem{
		PortraitAssetJob: job,
		Owner:            buildPortraitAssetOwnerInfo(job.UserId),
	})
}

func AdminDeleteOfficialPortraitAssetJob(c *gin.Context) {
	id, err := strconv.Atoi(c.Param("id"))
	if err != nil {
		common.ApiError(c, err)
		return
	}
	if err := model.DeleteOfficialPortraitAssetJobByID(id); err != nil {
		common.ApiError(c, err)
		return
	}
	common.ApiSuccess(c, nil)
}

func AdminListVirtualPortraitAssets(c *gin.Context) {
	pageInfo := common.GetPageQuery(c)
	queryParams, err := getAdminPortraitAssetQuery(c)
	if err != nil {
		common.ApiErrorMsg(c, err.Error())
		return
	}
	assets, err := model.GetAdminVirtualPortraitAssets(queryParams.virtual, pageInfo.GetStartIdx(), pageInfo.GetPageSize())
	if err != nil {
		common.ApiError(c, err)
		return
	}
	service.SyncUserVirtualPortraitAssetsForDisplay(assets)
	decorateUserVirtualPortraitAssetsResponse(c, assets)
	total, err := model.CountAdminVirtualPortraitAssets(queryParams.virtual)
	if err != nil {
		common.ApiError(c, err)
		return
	}
	pageInfo.SetTotal(int(total))
	pageInfo.SetItems(buildAdminVirtualPortraitAssetItems(assets))
	common.ApiSuccess(c, pageInfo)
}

func AdminSyncVirtualPortraitAsset(c *gin.Context) {
	id, err := strconv.Atoi(c.Param("id"))
	if err != nil {
		common.ApiError(c, err)
		return
	}
	asset, err := model.GetVirtualPortraitAssetByID(id)
	if err != nil {
		common.ApiError(c, err)
		return
	}
	asset, err = service.SyncUserVirtualPortraitAssetForExternalUser(
		asset.UserId,
		asset.Id,
		asset.ExternalUserID,
		true,
	)
	if err != nil {
		common.ApiError(c, err)
		return
	}
	decorateUserVirtualPortraitAssetResponse(c, asset)
	common.ApiSuccess(c, adminVirtualPortraitAssetItem{
		VirtualPortraitAsset: asset,
		Owner:                buildPortraitAssetOwnerInfo(asset.UserId),
	})
}

func AdminDeleteVirtualPortraitAsset(c *gin.Context) {
	id, err := strconv.Atoi(c.Param("id"))
	if err != nil {
		common.ApiError(c, err)
		return
	}
	if err := model.DeleteVirtualPortraitAssetByID(id); err != nil {
		common.ApiError(c, err)
		return
	}
	common.ApiSuccess(c, nil)
}

type adminPortraitAssetQuery struct {
	official model.AdminOfficialPortraitAssetQuery
	virtual  model.AdminVirtualPortraitAssetQuery
}

func getAdminPortraitAssetQuery(c *gin.Context) (*adminPortraitAssetQuery, error) {
	userID, err := getOptionalPositiveIntQuery(c, "user_id")
	if err != nil {
		return nil, err
	}
	folderID, filterFolderID, err := getFolderIDFilterFromQuery(c)
	if err != nil {
		return nil, err
	}
	createdFrom, err := getOptionalInt64Query(c, "created_from")
	if err != nil {
		return nil, err
	}
	createdTo, err := getOptionalInt64Query(c, "created_to")
	if err != nil {
		return nil, err
	}
	externalUserID, err := normalizeExternalUserIDInput(c.Query("external_user_id"))
	if err != nil {
		return nil, err
	}
	query := &adminPortraitAssetQuery{
		official: model.AdminOfficialPortraitAssetQuery{
			UserID:         userID,
			ExternalUserID: externalUserID,
			Status:         strings.TrimSpace(c.Query("status")),
			FolderID:       folderID,
			FilterFolderID: filterFolderID,
			Keyword:        strings.TrimSpace(c.Query("keyword")),
			AssetID:        strings.TrimSpace(c.Query("asset_id")),
			CreatedFrom:    createdFrom,
			CreatedTo:      createdTo,
		},
		virtual: model.AdminVirtualPortraitAssetQuery{
			UserID:         userID,
			ExternalUserID: externalUserID,
			Status:         strings.TrimSpace(c.Query("status")),
			FolderID:       folderID,
			FilterFolderID: filterFolderID,
			Keyword:        strings.TrimSpace(c.Query("keyword")),
			AssetID:        strings.TrimSpace(c.Query("asset_id")),
			CreatedFrom:    createdFrom,
			CreatedTo:      createdTo,
		},
	}
	return query, nil
}

func getOptionalPositiveIntQuery(c *gin.Context, key string) (int, error) {
	raw := strings.TrimSpace(c.Query(key))
	if raw == "" {
		return 0, nil
	}
	value, err := strconv.Atoi(raw)
	if err != nil || value < 0 {
		return 0, fmt.Errorf("%s must be a non-negative integer", key)
	}
	return value, nil
}

func getOptionalInt64Query(c *gin.Context, key string) (int64, error) {
	raw := strings.TrimSpace(c.Query(key))
	if raw == "" {
		return 0, nil
	}
	value, err := strconv.ParseInt(raw, 10, 64)
	if err != nil || value < 0 {
		return 0, fmt.Errorf("%s must be a non-negative integer", key)
	}
	return value, nil
}

func buildAdminOfficialPortraitAssetItems(jobs []*model.PortraitAssetJob) []*adminOfficialPortraitAssetItem {
	items := make([]*adminOfficialPortraitAssetItem, 0, len(jobs))
	owners := make(map[int]portraitAssetOwnerInfo)
	for _, job := range jobs {
		owner := owners[job.UserId]
		if owner.UserID == 0 {
			owner = buildPortraitAssetOwnerInfo(job.UserId)
			owners[job.UserId] = owner
		}
		items = append(items, &adminOfficialPortraitAssetItem{
			PortraitAssetJob: job,
			Owner:            owner,
		})
	}
	return items
}

func buildAdminVirtualPortraitAssetItems(assets []*model.VirtualPortraitAsset) []*adminVirtualPortraitAssetItem {
	items := make([]*adminVirtualPortraitAssetItem, 0, len(assets))
	owners := make(map[int]portraitAssetOwnerInfo)
	for _, asset := range assets {
		owner := owners[asset.UserId]
		if owner.UserID == 0 {
			owner = buildPortraitAssetOwnerInfo(asset.UserId)
			owners[asset.UserId] = owner
		}
		items = append(items, &adminVirtualPortraitAssetItem{
			VirtualPortraitAsset: asset,
			Owner:                owner,
		})
	}
	return items
}

func buildPortraitAssetOwnerInfo(userID int) portraitAssetOwnerInfo {
	info := portraitAssetOwnerInfo{UserID: userID}
	user, err := model.GetUserCache(userID)
	if err != nil || user == nil {
		return info
	}
	info.Username = user.Username
	info.Email = user.Email
	return info
}
