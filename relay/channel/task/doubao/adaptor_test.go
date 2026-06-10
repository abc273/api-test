package doubao

import (
	"bytes"
	"io"
	"net/http"
	"net/http/httptest"
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
