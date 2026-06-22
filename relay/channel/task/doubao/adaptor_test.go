package doubao

import (
	"bytes"
	"fmt"
	"io"
	"net/http"
	"net/http/httptest"
	"os"
	"strings"
	"testing"

	"github.com/QuantumNous/new-api/common"
	"github.com/QuantumNous/new-api/dto"
	"github.com/QuantumNous/new-api/model"
	relaycommon "github.com/QuantumNous/new-api/relay/common"
	"github.com/QuantumNous/new-api/service"
	"github.com/QuantumNous/new-api/setting/config"
	"github.com/gin-gonic/gin"
	"github.com/stretchr/testify/require"
)

func TestEstimateBillingSkipsLegacyVideoRatioForOutputTierPrice(t *testing.T) {
	gin.SetMode(gin.TestMode)

	saved := map[string]string{}
	require.NoError(t, config.GlobalConfig.SaveToDB(func(key, value string) error {
		saved[key] = value
		return nil
	}))
	t.Cleanup(func() {
		require.NoError(t, config.GlobalConfig.LoadFromDB(saved))
	})

	require.NoError(t, config.GlobalConfig.LoadFromDB(map[string]string{
		"billing_setting.billing_mode": `{"doubao-seedance-2-0-260128":"output_tier_price"}`,
	}))

	recorder := httptest.NewRecorder()
	ctx, _ := gin.CreateTestContext(recorder)
	ctx.Set("task_request", relaycommon.TaskSubmitReq{
		Metadata: map[string]interface{}{
			"content": []interface{}{
				map[string]interface{}{
					"type": "video_url",
					"video_url": map[string]interface{}{
						"url": "https://example.com/video.mp4",
					},
				},
			},
		},
	})

	adaptor := &TaskAdaptor{}
	ratios := adaptor.EstimateBilling(ctx, &relaycommon.RelayInfo{
		OriginModelName: "doubao-seedance-2-0-260128",
	})
	require.Nil(t, ratios)
}

func TestEstimateBillingKeepsLegacyVideoRatioForRatioBilling(t *testing.T) {
	gin.SetMode(gin.TestMode)

	saved := map[string]string{}
	require.NoError(t, config.GlobalConfig.SaveToDB(func(key, value string) error {
		saved[key] = value
		return nil
	}))
	t.Cleanup(func() {
		require.NoError(t, config.GlobalConfig.LoadFromDB(saved))
	})

	require.NoError(t, config.GlobalConfig.LoadFromDB(map[string]string{
		"billing_setting.billing_mode": `{"doubao-seedance-2-0-260128":"per-token"}`,
	}))

	recorder := httptest.NewRecorder()
	ctx, _ := gin.CreateTestContext(recorder)
	ctx.Set("task_request", relaycommon.TaskSubmitReq{
		Metadata: map[string]interface{}{
			"content": []interface{}{
				map[string]interface{}{
					"type": "video_url",
					"video_url": map[string]interface{}{
						"url": "https://example.com/video.mp4",
					},
				},
			},
		},
	})

	adaptor := &TaskAdaptor{}
	ratios := adaptor.EstimateBilling(ctx, &relaycommon.RelayInfo{
		OriginModelName: "doubao-seedance-2-0-260128",
	})
	require.Equal(t, map[string]float64{
		"video_input": 28.0 / 46.0,
	}, ratios)
}

func TestConvertToRequestPayloadUsesDurationField(t *testing.T) {
	adaptor := &TaskAdaptor{}
	req := relaycommon.TaskSubmitReq{
		Model:    "doubao-seedance-2-0-260128",
		Prompt:   "make a video",
		Duration: 15,
	}

	payload, err := adaptor.convertToRequestPayload(&req, &relaycommon.RelayInfo{})
	require.NoError(t, err)
	require.NotNil(t, payload.Duration)
	require.Equal(t, dto.IntValue(15), *payload.Duration)
}

func TestConvertToRequestPayloadUsesDefaultResolutionForDedicatedSRModelAlias(t *testing.T) {
	adaptor := &TaskAdaptor{}
	req := relaycommon.TaskSubmitReq{
		Model:  "seedance2-sr",
		Prompt: "make a video",
	}

	payload, err := adaptor.convertToRequestPayload(&req, &relaycommon.RelayInfo{
		ChannelMeta: &relaycommon.ChannelMeta{
			UpstreamModelName: "ep-seedance2",
		},
	})
	require.NoError(t, err)
	require.Equal(t, "ep-seedance2", payload.Model)
	require.Equal(t, "480p", payload.Resolution)
}

func TestConvertToRequestPayloadUsesDefaultResolutionForSeedance15SRModelAlias(t *testing.T) {
	adaptor := &TaskAdaptor{}
	req := relaycommon.TaskSubmitReq{
		Model:  "seedance1.5-sr",
		Prompt: "make a video",
	}

	payload, err := adaptor.convertToRequestPayload(&req, &relaycommon.RelayInfo{
		ChannelMeta: &relaycommon.ChannelMeta{
			UpstreamModelName: "ep-seedance15",
		},
	})
	require.NoError(t, err)
	require.Equal(t, "ep-seedance15", payload.Model)
	require.Equal(t, "480p", payload.Resolution)
}

func TestConvertToRequestPayloadUsesDefaultResolutionForProductionFastSRModelAlias(t *testing.T) {
	adaptor := &TaskAdaptor{}
	req := relaycommon.TaskSubmitReq{
		Model:  "seedance2.0fast-sr",
		Prompt: "make a video",
	}

	payload, err := adaptor.convertToRequestPayload(&req, &relaycommon.RelayInfo{
		ChannelMeta: &relaycommon.ChannelMeta{
			UpstreamModelName: "ep-seedance2-fast",
		},
	})
	require.NoError(t, err)
	require.Equal(t, "ep-seedance2-fast", payload.Model)
	require.Equal(t, "480p", payload.Resolution)
}

func TestConvertToRequestPayloadUsesTopLevelResolutionAndRatio(t *testing.T) {
	adaptor := &TaskAdaptor{}
	req := relaycommon.TaskSubmitReq{
		Model:      "seedance2.0fast-sr",
		Prompt:     "make a video",
		Resolution: "480p",
		Ratio:      "16:9",
	}

	payload, err := adaptor.convertToRequestPayload(&req, &relaycommon.RelayInfo{
		ChannelMeta: &relaycommon.ChannelMeta{
			UpstreamModelName: "ep-seedance2-fast",
		},
	})
	require.NoError(t, err)
	require.Equal(t, "ep-seedance2-fast", payload.Model)
	require.Equal(t, "480p", payload.Resolution)
	require.Equal(t, "16:9", payload.Ratio)
}

func TestConvertToRequestPayloadDownshiftsDedicatedSRResolution(t *testing.T) {
	adaptor := &TaskAdaptor{}
	req := relaycommon.TaskSubmitReq{
		Model:      "seedance2.0fast-sr",
		Prompt:     "make a video",
		Resolution: "720p",
	}

	payload, err := adaptor.convertToRequestPayload(&req, &relaycommon.RelayInfo{
		ChannelMeta: &relaycommon.ChannelMeta{
			UpstreamModelName: "ep-seedance2-fast",
		},
	})
	require.NoError(t, err)
	require.Equal(t, "480p", payload.Resolution)
}

func TestConvertToRequestPayloadMarksImagesAsReferenceImages(t *testing.T) {
	adaptor := &TaskAdaptor{}
	req := relaycommon.TaskSubmitReq{
		Model:  "doubao-seedance-2-0-fast-260128",
		Prompt: "make a video",
		Images: []string{
			"https://example.com/ref-1.png",
			"https://example.com/ref-2.png",
		},
	}

	payload, err := adaptor.convertToRequestPayload(&req, &relaycommon.RelayInfo{})
	require.NoError(t, err)
	require.Len(t, payload.Content, 3)

	require.Equal(t, "image_url", payload.Content[0].Type)
	require.Equal(t, "reference_image", payload.Content[0].Role)
	require.NotNil(t, payload.Content[0].ImageURL)
	require.Equal(t, "https://example.com/ref-1.png", payload.Content[0].ImageURL.URL)

	require.Equal(t, "image_url", payload.Content[1].Type)
	require.Equal(t, "reference_image", payload.Content[1].Role)
	require.NotNil(t, payload.Content[1].ImageURL)
	require.Equal(t, "https://example.com/ref-2.png", payload.Content[1].ImageURL.URL)

	require.Equal(t, "text", payload.Content[2].Type)
	require.Equal(t, "make a video", payload.Content[2].Text)
}

func TestBuildRequestBodyMarksJSONImagesAsReferenceImages(t *testing.T) {
	gin.SetMode(gin.TestMode)

	requestBody := `{
		"model": "doubao-seedance-2-0-fast-260128",
		"prompt": "use the clothes and colors from the reference images",
		"duration": 5,
		"images": [
			"https://example.com/ref-1.png",
			"https://example.com/ref-2.png"
		]
	}`
	recorder := httptest.NewRecorder()
	ctx, _ := gin.CreateTestContext(recorder)
	ctx.Request = httptest.NewRequest(http.MethodPost, "/v1/video/generations", bytes.NewBufferString(requestBody))
	ctx.Request.Header.Set("Content-Type", "application/json")

	adaptor := &TaskAdaptor{}
	info := &relaycommon.RelayInfo{
		ChannelMeta:   &relaycommon.ChannelMeta{},
		TaskRelayInfo: &relaycommon.TaskRelayInfo{},
	}

	taskErr := adaptor.ValidateRequestAndSetAction(ctx, info)
	require.Nil(t, taskErr)

	upstreamBody, err := adaptor.BuildRequestBody(ctx, info)
	require.NoError(t, err)

	data, err := io.ReadAll(upstreamBody)
	require.NoError(t, err)

	var payload requestPayload
	require.NoError(t, common.Unmarshal(data, &payload))

	require.Equal(t, "doubao-seedance-2-0-fast-260128", payload.Model)
	require.NotNil(t, payload.Duration)
	require.Equal(t, dto.IntValue(5), *payload.Duration)
	require.Len(t, payload.Content, 3)

	for i, expectedURL := range []string{"https://example.com/ref-1.png", "https://example.com/ref-2.png"} {
		require.Equal(t, "image_url", payload.Content[i].Type)
		require.Equal(t, "reference_image", payload.Content[i].Role)
		require.NotNil(t, payload.Content[i].ImageURL)
		require.Equal(t, expectedURL, payload.Content[i].ImageURL.URL)
	}

	require.Equal(t, "text", payload.Content[2].Type)
	require.Equal(t, "use the clothes and colors from the reference images", payload.Content[2].Text)
}

func TestConvertToOpenAIVideoHidesDedicatedSuperResolutionMetadata(t *testing.T) {
	original, err := common.Marshal(map[string]any{
		"id":     "upstream-task-1",
		"status": "succeeded",
		"content": map[string]any{
			"video_url": "https://example.com/source.mp4",
		},
	})
	require.NoError(t, err)

	wrapped, err := common.Marshal(service.VideoSuperResolutionTaskData{
		Kind:     "video_super_resolution_chain",
		Original: original,
		PostProcess: &service.VideoSuperResolutionState{
			Type:             "volc_las_video_super_resolution",
			TaskID:           "sr-task-1",
			Status:           "SUCCESS",
			SourceURL:        "https://example.com/source.mp4",
			SourceResolution: "480p",
			TargetResolution: "720p",
			TargetWidth:      1280,
			ResultURL:        "https://example.com/final.mp4",
		},
	})
	require.NoError(t, err)

	task := &model.Task{
		TaskID:    "task_public_1",
		Status:    model.TaskStatusSuccess,
		Progress:  "100%",
		CreatedAt: 100,
		UpdatedAt: 200,
		Data:      wrapped,
		PrivateData: model.TaskPrivateData{
			ResultURL: "https://example.com/final.mp4",
		},
		Properties: model.Properties{
			OriginModelName: service.Seedance20FastSRModelAlias,
		},
	}

	adaptor := &TaskAdaptor{}
	body, err := adaptor.ConvertToOpenAIVideo(task)
	require.NoError(t, err)

	var video dto.OpenAIVideo
	require.NoError(t, common.Unmarshal(body, &video))
	require.Equal(t, "task_public_1", video.ID)
	require.Equal(t, dto.VideoStatusCompleted, video.Status)
	require.Equal(t, service.Seedance20FastSRModelAlias, video.Model)
	require.Equal(t, "https://example.com/final.mp4", video.Metadata["url"])
	_, hasSuperResolution := video.Metadata["super_resolution"]
	require.False(t, hasSuperResolution)
}

func TestParseTaskResultMapsCancelledStatus(t *testing.T) {
	adaptor := &TaskAdaptor{}
	body := []byte(`{
		"id": "cgt-test-cancelled",
		"model": "doubao-seedance-2-0-fast-260128",
		"status": "cancelled",
		"content": {
			"video_url": ""
		}
	}`)

	result, err := adaptor.ParseTaskResult(body)
	require.NoError(t, err)
	require.Equal(t, model.TaskStatusCancelled, result.Status)
	require.Equal(t, "100%", result.Progress)
}

func TestDeleteTaskUsesDoubaoDeleteEndpoint(t *testing.T) {
	gin.SetMode(gin.TestMode)
	service.InitHttpClient()

	var (
		gotMethod string
		gotPath   string
		gotAuth   string
	)
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		gotMethod = r.Method
		gotPath = r.URL.Path
		gotAuth = r.Header.Get("Authorization")
		w.WriteHeader(http.StatusOK)
	}))
	defer server.Close()

	recorder := httptest.NewRecorder()
	ctx, _ := gin.CreateTestContext(recorder)
	ctx.Request = httptest.NewRequest(http.MethodDelete, "/v1/video/generations/task_public_1", nil)

	adaptor := &TaskAdaptor{}
	info := &relaycommon.RelayInfo{
		ChannelMeta: &relaycommon.ChannelMeta{
			ChannelBaseUrl: server.URL,
			ApiKey:         "sk-test-delete",
		},
	}
	adaptor.Init(info)

	task := &model.Task{
		TaskID: "task_public_1",
		PrivateData: model.TaskPrivateData{
			UpstreamTaskID: "cgt-delete-me",
		},
	}

	taskErr := adaptor.DeleteTask(ctx, info, task)
	require.Nil(t, taskErr)
	require.Equal(t, http.MethodDelete, gotMethod)
	require.Equal(t, "/api/v3/contents/generations/tasks/cgt-delete-me", gotPath)
	require.Equal(t, "Bearer sk-test-delete", gotAuth)
}

func TestConvertToRequestPayloadScopesPortraitAssetReferencesByExternalUserID(t *testing.T) {
	setupDoubaoPortraitAssetTestDB(t)
	require.NoError(t, model.DB.Create(&model.VirtualPortraitAsset{
		UserId:         42,
		ExternalUserID: "customer-user-a",
		Name:           "virtual-a",
		VolcAssetID:    "asset-virtual-a",
		Status:         model.VirtualPortraitAssetStatusActive,
		CreatedTime:    1,
		UpdatedTime:    1,
	}).Error)

	adaptor := &TaskAdaptor{}
	req := relaycommon.TaskSubmitReq{
		Model:          "doubao-seedance-2-0-fast-260128",
		Prompt:         "make a video",
		ExternalUserID: "customer-user-a",
		Metadata: map[string]any{
			"external_user_id": "customer-user-b",
			"content": []map[string]any{
				{
					"type": "image_url",
					"role": "reference_image",
					"image_url": map[string]any{
						"url": "asset://asset-virtual-a",
					},
				},
			},
		},
	}
	_, err := adaptor.convertToRequestPayload(&req, &relaycommon.RelayInfo{UserId: 42})
	require.NoError(t, err)

	req.ExternalUserID = ""
	_, err = adaptor.convertToRequestPayload(&req, &relaycommon.RelayInfo{UserId: 42})
	require.Error(t, err)

	req.Metadata["external_user_id"] = "customer-user-a"
	_, err = adaptor.convertToRequestPayload(&req, &relaycommon.RelayInfo{UserId: 42})
	require.NoError(t, err)
}

func TestConvertToRequestPayloadAppendsTopLevelImagesAfterMetadataContent(t *testing.T) {
	setupDoubaoPortraitAssetTestDB(t)
	require.NoError(t, model.DB.Create(&model.VirtualPortraitAsset{
		UserId:      42,
		Name:        "virtual-a",
		VolcAssetID: "asset-virtual-a",
		Status:      model.VirtualPortraitAssetStatusActive,
		CreatedTime: 1,
		UpdatedTime: 1,
	}).Error)
	require.NoError(t, model.DB.Create(&model.PortraitAssetJob{
		UserId:      42,
		Name:        "official-image",
		Source:      model.PortraitAssetSourceOfficial,
		Status:      model.PortraitAssetStatusReady,
		AssetID:     "asset-official-image-a",
		AssetType:   "Image",
		CreatedTime: 1,
		UpdatedTime: 1,
		ReadyTime:   1,
	}).Error)

	adaptor := &TaskAdaptor{}
	req := relaycommon.TaskSubmitReq{
		Model:  "doubao-seedance-2-0-fast-260128",
		Prompt: "make a video",
		Images: []string{"data:image/jpeg;base64,cat-image"},
		Metadata: map[string]any{
			"content": []map[string]any{
				{
					"type": "image_url",
					"role": "reference_image",
					"image_url": map[string]any{
						"url": "asset://asset-virtual-a",
					},
				},
				{
					"type": "image_url",
					"role": "reference_image",
					"image_url": map[string]any{
						"url": "asset://asset-official-image-a",
					},
				},
			},
		},
	}

	payload, err := adaptor.convertToRequestPayload(&req, &relaycommon.RelayInfo{UserId: 42})
	require.NoError(t, err)
	require.Len(t, payload.Content, 4)
	require.Equal(t, "asset://asset-virtual-a", payload.Content[0].ImageURL.URL)
	require.Equal(t, "asset://asset-official-image-a", payload.Content[1].ImageURL.URL)
	require.Equal(t, "data:image/jpeg;base64,cat-image", payload.Content[2].ImageURL.URL)
	require.Equal(t, "text", payload.Content[3].Type)
}

func TestConvertToRequestPayloadMapsOfficialPortraitAssetByAssetType(t *testing.T) {
	setupDoubaoPortraitAssetTestDB(t)
	require.NoError(t, model.DB.Create(&model.PortraitAssetJob{
		UserId:         42,
		ExternalUserID: "customer-user-a",
		Name:           "official-video",
		Source:         model.PortraitAssetSourceOfficial,
		Status:         model.PortraitAssetStatusReady,
		AssetID:        "asset-official-video-a",
		AssetType:      "Video",
		CreatedTime:    1,
		UpdatedTime:    1,
		ReadyTime:      1,
	}).Error)

	adaptor := &TaskAdaptor{}
	req := relaycommon.TaskSubmitReq{
		Model:          "doubao-seedance-2-0-fast-260128",
		Prompt:         "make a video",
		ExternalUserID: "customer-user-a",
		Images:         []string{"https://example.com/ref.png"},
		Metadata: map[string]any{
			"asset_id": "asset://asset-official-video-a",
		},
	}

	payload, err := adaptor.convertToRequestPayload(&req, &relaycommon.RelayInfo{UserId: 42})
	require.NoError(t, err)
	require.Len(t, payload.Content, 3)

	require.Equal(t, "image_url", payload.Content[0].Type)
	require.Equal(t, "reference_image", payload.Content[0].Role)
	require.NotNil(t, payload.Content[0].ImageURL)
	require.Equal(t, "https://example.com/ref.png", payload.Content[0].ImageURL.URL)

	require.Equal(t, "video_url", payload.Content[1].Type)
	require.Equal(t, "reference_video", payload.Content[1].Role)
	require.NotNil(t, payload.Content[1].VideoURL)
	require.Equal(t, "asset://asset-official-video-a", payload.Content[1].VideoURL.URL)

	require.Equal(t, "text", payload.Content[2].Type)
	require.Equal(t, "make a video", payload.Content[2].Text)
}

func TestConvertToRequestPayloadMapsOfficialPortraitAssetJobIDByAssetType(t *testing.T) {
	setupDoubaoPortraitAssetTestDB(t)
	require.NoError(t, model.DB.Create(&model.PortraitAssetJob{
		Id:             7,
		UserId:         42,
		ExternalUserID: "customer-user-a",
		Name:           "official-audio",
		Source:         model.PortraitAssetSourceOfficial,
		Status:         model.PortraitAssetStatusReady,
		AssetID:        "asset-official-audio-a",
		AssetType:      "Audio",
		CreatedTime:    1,
		UpdatedTime:    1,
		ReadyTime:      1,
	}).Error)

	adaptor := &TaskAdaptor{}
	req := relaycommon.TaskSubmitReq{
		Model:          "doubao-seedance-2-0-fast-260128",
		Prompt:         "make a video",
		ExternalUserID: "customer-user-a",
		Metadata: map[string]any{
			"portrait_asset_id": 7,
		},
	}

	payload, err := adaptor.convertToRequestPayload(&req, &relaycommon.RelayInfo{UserId: 42})
	require.NoError(t, err)
	require.Len(t, payload.Content, 2)

	require.Equal(t, "audio_url", payload.Content[0].Type)
	require.Empty(t, payload.Content[0].Role)
	require.NotNil(t, payload.Content[0].AudioURL)
	require.Equal(t, "asset://asset-official-audio-a", payload.Content[0].AudioURL.URL)

	require.Equal(t, "text", payload.Content[1].Type)
	require.Equal(t, "make a video", payload.Content[1].Text)
}

func TestConvertToRequestPayloadMapsOfficialPortraitAssetByAssetTypeWithoutExternalUserID(t *testing.T) {
	setupDoubaoPortraitAssetTestDB(t)
	require.NoError(t, model.DB.Create(&model.PortraitAssetJob{
		Id:             18,
		UserId:         42,
		ExternalUserID: "customer-user-a",
		Name:           "official-image",
		Source:         model.PortraitAssetSourceOfficial,
		Status:         model.PortraitAssetStatusReady,
		AssetID:        "asset-official-image-a",
		AssetType:      "Image",
		CreatedTime:    1,
		UpdatedTime:    1,
		ReadyTime:      1,
	}).Error)

	adaptor := &TaskAdaptor{}
	req := relaycommon.TaskSubmitReq{
		Model:  "doubao-seedance-2-0-fast-260128",
		Prompt: "make a video",
		Metadata: map[string]any{
			"asset_id": "asset://asset-official-image-a",
		},
	}

	payload, err := adaptor.convertToRequestPayload(&req, &relaycommon.RelayInfo{UserId: 42})
	require.NoError(t, err)
	require.Len(t, payload.Content, 2)
	require.Equal(t, "image_url", payload.Content[0].Type)
	require.Equal(t, "reference_image", payload.Content[0].Role)
	require.NotNil(t, payload.Content[0].ImageURL)
	require.Equal(t, "asset://asset-official-image-a", payload.Content[0].ImageURL.URL)
	require.Equal(t, "text", payload.Content[1].Type)
}

func TestConvertToRequestPayloadMapsOfficialPortraitAssetJobIDWithoutExternalUserID(t *testing.T) {
	setupDoubaoPortraitAssetTestDB(t)
	require.NoError(t, model.DB.Create(&model.PortraitAssetJob{
		Id:             19,
		UserId:         42,
		ExternalUserID: "customer-user-a",
		Name:           "official-image-by-job",
		Source:         model.PortraitAssetSourceOfficial,
		Status:         model.PortraitAssetStatusReady,
		AssetID:        "asset-official-image-b",
		AssetType:      "Image",
		CreatedTime:    1,
		UpdatedTime:    1,
		ReadyTime:      1,
	}).Error)

	adaptor := &TaskAdaptor{}
	req := relaycommon.TaskSubmitReq{
		Model:  "doubao-seedance-2-0-fast-260128",
		Prompt: "make a video",
		Metadata: map[string]any{
			"portrait_asset_id": 19,
		},
	}

	payload, err := adaptor.convertToRequestPayload(&req, &relaycommon.RelayInfo{UserId: 42})
	require.NoError(t, err)
	require.Len(t, payload.Content, 2)
	require.Equal(t, "image_url", payload.Content[0].Type)
	require.Equal(t, "reference_image", payload.Content[0].Role)
	require.NotNil(t, payload.Content[0].ImageURL)
	require.Equal(t, "asset://asset-official-image-b", payload.Content[0].ImageURL.URL)
	require.Equal(t, "text", payload.Content[1].Type)
}

func setupDoubaoPortraitAssetTestDB(t *testing.T) {
	originalDB := model.DB
	originalLogDB := model.LOG_DB
	originalIsMasterNode := common.IsMasterNode
	originalSQLitePath := common.SQLitePath
	originalUsingSQLite := common.UsingSQLite
	originalUsingMySQL := common.UsingMySQL
	originalUsingPostgreSQL := common.UsingPostgreSQL
	originalDSN := os.Getenv("SQL_DSN")
	testDB := model.DB
	t.Cleanup(func() {
		model.DB = originalDB
		model.LOG_DB = originalLogDB
		common.IsMasterNode = originalIsMasterNode
		common.SQLitePath = originalSQLitePath
		common.UsingSQLite = originalUsingSQLite
		common.UsingMySQL = originalUsingMySQL
		common.UsingPostgreSQL = originalUsingPostgreSQL
		if testDB != nil {
			sqlDB, err := testDB.DB()
			if err == nil {
				_ = sqlDB.Close()
			}
		}
		if originalDSN == "" {
			require.NoError(t, os.Unsetenv("SQL_DSN"))
		} else {
			require.NoError(t, os.Setenv("SQL_DSN", originalDSN))
		}
	})

	common.IsMasterNode = false
	common.SQLitePath = fmt.Sprintf("file:%s_doubao_portrait_asset?mode=memory&cache=shared", strings.ReplaceAll(t.Name(), "/", "_"))
	common.UsingSQLite = false
	common.UsingMySQL = false
	common.UsingPostgreSQL = false
	require.NoError(t, os.Setenv("SQL_DSN", "local"))
	require.NoError(t, model.InitDB())
	testDB = model.DB
	require.NoError(t, model.DB.AutoMigrate(&model.PortraitAssetJob{}, &model.VirtualPortraitAsset{}))
}
