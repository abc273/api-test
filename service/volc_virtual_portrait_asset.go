package service

import (
	"errors"
	"fmt"
	"strings"

	"github.com/QuantumNous/new-api/common"
	"github.com/QuantumNous/new-api/model"
	"gorm.io/gorm"
)

type VolcPortraitAssetGroupInfo struct {
	Id          string `json:"Id"`
	Name        string `json:"Name"`
	Title       string `json:"Title"`
	Description string `json:"Description"`
	GroupType   string `json:"GroupType"`
	ProjectName string `json:"ProjectName"`
	CreateTime  string `json:"CreateTime"`
	UpdateTime  string `json:"UpdateTime"`
}

func CreateVolcPortraitAssetGroup(name string, description string, projectName string) (string, error) {
	payload := map[string]any{
		"Name":        strings.TrimSpace(name),
		"Description": strings.TrimSpace(description),
		"GroupType":   "AIGC",
		"ProjectName": resolveVolcPortraitProjectName(projectName),
	}
	var result struct {
		Id string `json:"Id"`
	}
	if err := callVolcPortraitOpenAPI("CreateAssetGroup", payload, &result); err != nil {
		return "", err
	}
	if strings.TrimSpace(result.Id) == "" {
		return "", fmt.Errorf("volc portrait create asset group response has no group id")
	}
	return strings.TrimSpace(result.Id), nil
}

func EnsureUserVirtualPortraitAssetGroup(userId int) (*model.VirtualPortraitAssetGroup, error) {
	projectName := GetVolcPortraitProjectName()
	groupName := buildUserVirtualPortraitAssetGroupName(userId)
	description := buildUserVirtualPortraitAssetGroupDescription(userId)

	group, err := model.GetUserVirtualPortraitAssetGroup(userId)
	switch {
	case err == nil:
		if strings.TrimSpace(group.VolcGroupID) != "" && group.Status == model.VirtualPortraitAssetGroupStatusActive {
			return group, nil
		}
		// Another request is likely provisioning the same group. Let it finish instead of
		// creating duplicate remote groups.
		if strings.TrimSpace(group.VolcGroupID) == "" &&
			group.Status == model.VirtualPortraitAssetGroupStatusCreating &&
			common.GetTimestamp()-group.UpdatedTime < 30 {
			return nil, errors.New("素材组正在初始化，请稍后再试")
		}
	case errors.Is(err, gorm.ErrRecordNotFound):
		group = &model.VirtualPortraitAssetGroup{
			UserId:      userId,
			Name:        groupName,
			Description: description,
			ProjectName: projectName,
			Status:      model.VirtualPortraitAssetGroupStatusCreating,
		}
	default:
		return nil, err
	}

	group.Name = groupName
	group.Description = description
	group.ProjectName = projectName
	group.Status = model.VirtualPortraitAssetGroupStatusCreating
	group.ErrorMessage = ""
	group.VolcGroupID = strings.TrimSpace(group.VolcGroupID)
	if err := model.SaveVirtualPortraitAssetGroup(group); err != nil {
		return nil, err
	}

	groupID, err := CreateVolcPortraitAssetGroup(groupName, description, projectName)
	if err != nil {
		group.Status = model.VirtualPortraitAssetGroupStatusFailed
		group.ErrorMessage = err.Error()
		_ = model.SaveVirtualPortraitAssetGroup(group)
		return nil, err
	}

	group.VolcGroupID = groupID
	group.Status = model.VirtualPortraitAssetGroupStatusActive
	group.ErrorMessage = ""
	if err := model.SaveVirtualPortraitAssetGroup(group); err != nil {
		return nil, err
	}
	return group, nil
}

func CreateUserVirtualPortraitAsset(userId int, name string, assetURL string, assetType string) (*model.VirtualPortraitAsset, error) {
	group, err := EnsureUserVirtualPortraitAssetGroup(userId)
	if err != nil {
		return nil, err
	}

	assetID, err := CreateVolcPortraitAsset(group.VolcGroupID, assetURL, assetType, name, group.ProjectName)
	if err != nil {
		return nil, err
	}

	asset, err := model.CreateUserVirtualPortraitAsset(userId, group, name, normalizeVolcPortraitAssetType(assetType), assetURL, assetID)
	if err != nil {
		return nil, err
	}

	info, err := GetVolcPortraitAsset(assetID, group.ProjectName)
	if err == nil {
		_ = applyVolcVirtualPortraitAssetInfo(asset, info)
	}
	return asset, nil
}

func SyncUserVirtualPortraitAsset(userId int, id int) (*model.VirtualPortraitAsset, error) {
	asset, err := model.GetUserVirtualPortraitAssetByID(userId, id)
	if err != nil {
		return nil, err
	}
	if strings.TrimSpace(asset.VolcAssetID) == "" {
		return nil, errors.New("素材缺少火山资产 ID")
	}
	info, err := GetVolcPortraitAsset(asset.VolcAssetID, asset.ProjectName)
	if err != nil {
		return nil, err
	}
	if err := applyVolcVirtualPortraitAssetInfo(asset, info); err != nil {
		return nil, err
	}
	return asset, nil
}

func applyVolcVirtualPortraitAssetInfo(asset *model.VirtualPortraitAsset, info *VolcPortraitAssetInfo) error {
	if asset == nil {
		return gorm.ErrInvalidData
	}
	if info == nil {
		return nil
	}
	asset.VolcAssetID = model.NormalizePortraitAssetID(firstNonEmpty(info.Id, asset.VolcAssetID))
	asset.VolcGroupID = firstNonEmpty(info.GroupId, asset.VolcGroupID)
	if strings.TrimSpace(info.AssetType) != "" {
		asset.AssetType = strings.TrimSpace(info.AssetType)
	}
	if strings.TrimSpace(info.ProjectName) != "" {
		asset.ProjectName = strings.TrimSpace(info.ProjectName)
	}
	if strings.TrimSpace(info.URL) != "" {
		asset.PreviewURL = strings.TrimSpace(info.URL)
	}
	asset.VolcStatus = strings.TrimSpace(info.Status)
	asset.Status = model.NormalizeVirtualPortraitAssetStatus(info.Status)
	switch asset.Status {
	case model.VirtualPortraitAssetStatusActive:
		asset.ErrorMessage = ""
	case model.VirtualPortraitAssetStatusFailed:
		if asset.ErrorMessage == "" {
			asset.ErrorMessage = "素材入库失败"
		}
	default:
		asset.ErrorMessage = ""
	}
	return model.SaveVirtualPortraitAsset(asset)
}

func buildUserVirtualPortraitAssetGroupName(userId int) string {
	return fmt.Sprintf("user-%d-virtual-portraits", userId)
}

func buildUserVirtualPortraitAssetGroupDescription(userId int) string {
	return fmt.Sprintf("Managed virtual portrait asset group for user %d", userId)
}

func firstNonEmpty(values ...string) string {
	for _, value := range values {
		if strings.TrimSpace(value) != "" {
			return strings.TrimSpace(value)
		}
	}
	return ""
}
