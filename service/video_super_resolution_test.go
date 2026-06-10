package service

import (
	"context"
	"io"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/QuantumNous/new-api/common"
	"github.com/QuantumNous/new-api/model"
	relaycommon "github.com/QuantumNous/new-api/relay/common"
	"github.com/QuantumNous/new-api/setting/config"
	"github.com/stretchr/testify/require"
)

func TestResolveVideoSuperResolutionTarget(t *testing.T) {
	tests := []struct {
		name          string
		data          map[string]any
		wantSource    string
		wantTarget    string
		wantRatio     string
		wantWidth     int
		wantHeight    int
		originModel   string
		upstreamModel string
	}{
		{
			name:       "defaults to 16:9 when ratio omitted",
			data:       map[string]any{"resolution": "480p"},
			wantSource: "480p",
			wantTarget: "720p",
			wantRatio:  "16:9",
			wantWidth:  1280,
			wantHeight: 720,
		},
		{
			name:       "uses requested portrait ratio",
			data:       map[string]any{"resolution": "480p", "ratio": "9:16"},
			wantSource: "480p",
			wantTarget: "720p",
			wantRatio:  "9:16",
			wantWidth:  720,
			wantHeight: 1280,
		},
		{
			name:       "uses requested square ratio",
			data:       map[string]any{"resolution": "720p", "ratio": "1:1"},
			wantSource: "720p",
			wantTarget: "1080p",
			wantRatio:  "1:1",
			wantWidth:  1440,
			wantHeight: 1440,
		},
		{
			name:          "falls back to model-derived source resolution",
			data:          map[string]any{"ratio": "4:3"},
			wantSource:    "480p",
			wantTarget:    "720p",
			wantRatio:     "4:3",
			wantWidth:     1112,
			wantHeight:    834,
			originModel:   "seedance2-sr",
			upstreamModel: "doubao-seedance-2-0-260128",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			task := &model.Task{
				Properties: model.Properties{
					OriginModelName:   tt.originModel,
					UpstreamModelName: tt.upstreamModel,
				},
			}
			task.SetData(tt.data)

			source, target, ratio, width, height, ok := resolveVideoSuperResolutionTarget(task)
			require.True(t, ok)
			require.Equal(t, tt.wantSource, source)
			require.Equal(t, tt.wantTarget, target)
			require.Equal(t, tt.wantRatio, ratio)
			require.Equal(t, tt.wantWidth, width)
			require.Equal(t, tt.wantHeight, height)
		})
	}
}

func TestShouldNormalizeVideoSuperResolutionOutput(t *testing.T) {
	require.True(t, shouldNormalizeVideoSuperResolutionOutput(&VideoSuperResolutionState{
		SourceResolution: "480p",
		TargetResolution: "720p",
		TargetRatio:      "9:16",
		TargetWidth:      720,
		TargetHeight:     1280,
	}))

	require.False(t, shouldNormalizeVideoSuperResolutionOutput(&VideoSuperResolutionState{
		SourceResolution: "720p",
		TargetResolution: "1080p",
		TargetRatio:      "1:1",
		TargetWidth:      1440,
		TargetHeight:     1440,
	}))
}

func TestExtractOriginalVideoTaskData(t *testing.T) {
	original, err := common.Marshal(map[string]any{
		"content": map[string]any{
			"video_url": "https://example.com/source.mp4",
		},
		"resolution": "720p",
	})
	require.NoError(t, err)

	wrapped, err := common.Marshal(VideoSuperResolutionTaskData{
		Kind:     videoSuperResolutionTaskDataKind,
		Original: original,
		PostProcess: &VideoSuperResolutionState{
			TaskID:           "task-sr-123",
			Status:           "PENDING",
			SourceResolution: "720p",
			TargetResolution: "1080p",
			TargetWidth:      1920,
			TargetHeight:     1080,
		},
	})
	require.NoError(t, err)

	extracted := ExtractOriginalVideoTaskData(wrapped)
	require.JSONEq(t, string(original), string(extracted))

	state, ok, err := ReadVideoSuperResolutionState(wrapped)
	require.NoError(t, err)
	require.True(t, ok)
	require.Equal(t, "task-sr-123", state.TaskID)
	require.Equal(t, "1080p", state.TargetResolution)
	require.Equal(t, 1920, state.TargetWidth)
	require.Equal(t, 1080, state.TargetHeight)
}

func TestInferVideoResolutionFromModel(t *testing.T) {
	require.Equal(t, "720p", inferVideoResolutionFromModel("seedance1.5"))
	require.Equal(t, "480p", inferVideoResolutionFromModel("seedance1.5-sr"))
	require.Equal(t, "720p", inferVideoResolutionFromModel("seedance2"))
	require.Equal(t, "480p", inferVideoResolutionFromModel("seedance2-sr"))
	require.Equal(t, "720p", inferVideoResolutionFromModel("seedance2.0fast"))
	require.Equal(t, "480p", inferVideoResolutionFromModel("seedance2.0fast-sr"))
	require.Equal(t, "720p", inferVideoResolutionFromModel("doubao-seedance-2-0-fast-260128"))
	require.Equal(t, "", inferVideoResolutionFromModel("other-model"))
}

func TestResolveDedicatedVideoSuperResolutionSourceResolution(t *testing.T) {
	require.Equal(t, "480p", ResolveDedicatedVideoSuperResolutionSourceResolution("seedance1.5-sr", "720p"))
	require.Equal(t, "720p", ResolveDedicatedVideoSuperResolutionSourceResolution("seedance1.5-sr", "1080p"))
	require.Equal(t, "480p", ResolveDedicatedVideoSuperResolutionSourceResolution("seedance2-sr", "720p"))
	require.Equal(t, "720p", ResolveDedicatedVideoSuperResolutionSourceResolution("seedance2.0fast-sr", "1080p"))
	require.Equal(t, "480p", ResolveDedicatedVideoSuperResolutionSourceResolution("seedance2", "480p"))
}

func TestResolveVideoSuperResolutionObjectURLWithEndpoint(t *testing.T) {
	url, err := resolveVideoSuperResolutionObjectURL("tos://demo-bucket/results/output.mp4", videoSuperResolutionConfig{
		TOSEndpoint: "tos-cn-beijing.volces.com",
	})
	require.NoError(t, err)
	require.Equal(t, "https://demo-bucket.tos-cn-beijing.volces.com/results/output.mp4", url)
}

func TestResolveVideoSuperResolutionObjectURLWithPresign(t *testing.T) {
	url, err := resolveVideoSuperResolutionObjectURL("tos://demo-bucket/results/output.mp4", videoSuperResolutionConfig{
		TOSEndpoint:       "tos-cn-beijing.volces.com",
		TOSRegion:         "cn-beijing",
		TOSAccessKey:      "ak-test",
		TOSSecretKey:      "sk-test",
		TOSPresignExpires: 120,
	})
	require.NoError(t, err)
	require.Contains(t, url, "https://demo-bucket.tos-cn-beijing.volces.com/results/output.mp4")
	require.True(t, strings.Contains(url, "X-Tos-Algorithm=") || strings.Contains(url, "x-tos-algorithm="))
}

func TestIsVideoSuperResolutionRequested(t *testing.T) {
	require.True(t, IsVideoSuperResolutionRequested(map[string]any{
		"enable_video_super_resolution": true,
	}))
	require.True(t, IsVideoSuperResolutionRequested(map[string]any{
		"enable_video_super_resolution": "true",
	}))
	require.True(t, IsVideoSuperResolutionRequested(map[string]any{
		"enable_video_super_resolution": 1,
	}))
	require.False(t, IsVideoSuperResolutionRequested(map[string]any{
		"enable_video_super_resolution": false,
	}))
	require.False(t, IsVideoSuperResolutionRequested(map[string]any{}))
	require.False(t, IsVideoSuperResolutionRequested(nil))
}

func TestNormalizeVideoSuperResolutionModelAlias(t *testing.T) {
	require.Equal(t, "seedance1.5", NormalizeVideoSuperResolutionModelAlias("seedance1.5-sr"))
	require.Equal(t, "seedance2", NormalizeVideoSuperResolutionModelAlias("seedance2-sr"))
	require.Equal(t, "seedance2.0", NormalizeVideoSuperResolutionModelAlias("seedance2.0-sr"))
	require.Equal(t, "sd2.0fast", NormalizeVideoSuperResolutionModelAlias("sd2.0fast-sr"))
	require.Equal(t, "seedance2.0fast", NormalizeVideoSuperResolutionModelAlias("seedance2.0fast-sr"))
	require.Equal(t, "seedance2", NormalizeVideoSuperResolutionModelAlias("seedance2"))
}

func TestShouldAutoEnableVideoSuperResolution(t *testing.T) {
	require.True(t, ShouldAutoEnableVideoSuperResolution("seedance1.5-sr"))
	require.True(t, ShouldAutoEnableVideoSuperResolution("seedance2-sr"))
	require.True(t, ShouldAutoEnableVideoSuperResolution("seedance2.0-sr"))
	require.True(t, ShouldAutoEnableVideoSuperResolution("seedance2.0fast-sr"))
	require.True(t, ShouldAutoEnableVideoSuperResolution("sd2.0fast-sr"))
	require.False(t, ShouldAutoEnableVideoSuperResolution("seedance2"))
	require.False(t, ShouldAutoEnableVideoSuperResolution("seedance1.5"))
	require.False(t, ShouldAutoEnableVideoSuperResolution("seedance2.0fast"))
	require.False(t, ShouldAutoEnableVideoSuperResolution("sd2.0fast"))
}

func TestMaybeStartVideoSuperResolutionRequiresDedicatedSRModel(t *testing.T) {
	saved := map[string]string{}
	require.NoError(t, config.GlobalConfig.SaveToDB(func(key, value string) error {
		saved[key] = value
		return nil
	}))
	t.Cleanup(func() {
		require.NoError(t, config.GlobalConfig.LoadFromDB(saved))
	})

	require.NoError(t, config.GlobalConfig.LoadFromDB(map[string]string{
		"video_super_resolution.enabled":             "true",
		"video_super_resolution.base_url":            "https://las.example.com",
		"video_super_resolution.api_key":             "sr-api-key",
		"video_super_resolution.output_tos_path":     "tos://demo-bucket/video-sr",
		"video_super_resolution.operator_id":         "las_video_super_resolution",
		"video_super_resolution.operator_version":    "v1",
		"video_super_resolution.preserve_audio":      "true",
		"video_super_resolution.output_quality_mode": "balanced",
	}))

	started, err := MaybeStartVideoSuperResolution(nil, &model.Task{
		Properties: model.Properties{
			OriginModelName:               "seedance2-sr",
			VideoSuperResolutionRequested: false,
		},
	}, &relaycommon.TaskInfo{
		Status: string(model.TaskStatusSuccess),
		Url:    "https://example.com/video.mp4",
	})
	require.NoError(t, err)
	require.False(t, started)
}

func TestMaybeStartVideoSuperResolutionSubmitsTargetHeight(t *testing.T) {
	InitHttpClient()
	var submitted map[string]any
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		require.Equal(t, "/api/v1/submit", r.URL.Path)
		body, err := io.ReadAll(r.Body)
		require.NoError(t, err)
		require.NoError(t, common.Unmarshal(body, &submitted))
		w.Header().Set("Content-Type", "application/json")
		_, _ = w.Write([]byte(`{
			"metadata": {
				"task_id": "sr-task-submit",
				"task_status": "PENDING",
				"business_code": "0",
				"error_msg": ""
			},
			"data": {}
		}`))
	}))
	defer server.Close()

	saved := map[string]string{}
	require.NoError(t, config.GlobalConfig.SaveToDB(func(key, value string) error {
		saved[key] = value
		return nil
	}))
	t.Cleanup(func() {
		require.NoError(t, config.GlobalConfig.LoadFromDB(saved))
	})

	require.NoError(t, config.GlobalConfig.LoadFromDB(map[string]string{
		"video_super_resolution.enabled":             "true",
		"video_super_resolution.base_url":            server.URL,
		"video_super_resolution.api_key":             "sr-api-key",
		"video_super_resolution.output_tos_path":     "tos://demo-bucket/video-sr",
		"video_super_resolution.operator_id":         "las_video_super_resolution",
		"video_super_resolution.operator_version":    "v1",
		"video_super_resolution.preserve_audio":      "true",
		"video_super_resolution.output_quality_mode": "balanced",
	}))

	task := &model.Task{
		TaskID: "task-submit",
		Properties: model.Properties{
			OriginModelName:               "seedance2-sr",
			VideoSuperResolutionRequested: true,
		},
	}
	task.SetData(map[string]any{
		"resolution": "480p",
		"ratio":      "9:16",
	})

	started, err := MaybeStartVideoSuperResolution(context.Background(), task, &relaycommon.TaskInfo{
		Status: string(model.TaskStatusSuccess),
		Url:    "https://example.com/video.mp4",
	})
	require.NoError(t, err)
	require.True(t, started)

	data := submitted["data"].(map[string]any)
	require.Equal(t, float64(720), data["target_width"])
	require.Equal(t, float64(1280), data["target_height"])

	state, ok, err := ReadVideoSuperResolutionState(task.Data)
	require.NoError(t, err)
	require.True(t, ok)
	require.Equal(t, "9:16", state.TargetRatio)
	require.Equal(t, 720, state.TargetWidth)
	require.Equal(t, 1280, state.TargetHeight)
}

func TestPollVideoSuperResolutionFailsTaskWhenPostProcessFails(t *testing.T) {
	InitHttpClient()
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		require.Equal(t, "/api/v1/poll", r.URL.Path)
		w.Header().Set("Content-Type", "application/json")
		_, _ = w.Write([]byte(`{
			"metadata": {
				"task_id": "sr-task-1",
				"task_status": "FAILED",
				"business_code": "1001",
				"error_msg": "super resolution failed"
			},
			"data": {}
		}`))
	}))
	defer server.Close()

	saved := map[string]string{}
	require.NoError(t, config.GlobalConfig.SaveToDB(func(key, value string) error {
		saved[key] = value
		return nil
	}))
	t.Cleanup(func() {
		require.NoError(t, config.GlobalConfig.LoadFromDB(saved))
	})

	require.NoError(t, config.GlobalConfig.LoadFromDB(map[string]string{
		"video_super_resolution.enabled":          "true",
		"video_super_resolution.base_url":         server.URL,
		"video_super_resolution.api_key":          "sr-api-key",
		"video_super_resolution.output_tos_path":  "tos://demo-bucket/video-sr",
		"video_super_resolution.operator_id":      "las_video_super_resolution",
		"video_super_resolution.operator_version": "v1",
	}))

	original, err := common.Marshal(map[string]any{
		"content": map[string]any{
			"video_url": "https://example.com/original.mp4",
		},
		"resolution": "480p",
		"usage": map[string]any{
			"total_tokens": 4321,
		},
	})
	require.NoError(t, err)

	task := &model.Task{TaskID: "task-1"}
	task.Data, err = common.Marshal(VideoSuperResolutionTaskData{
		Kind:     videoSuperResolutionTaskDataKind,
		Original: original,
		PostProcess: &VideoSuperResolutionState{
			TaskID:           "sr-task-1",
			Status:           "RUNNING",
			SourceURL:        "https://example.com/original.mp4",
			SourceResolution: "480p",
			TargetResolution: "720p",
			TargetWidth:      1280,
		},
	})
	require.NoError(t, err)

	taskInfo, postProcessed, err := PollVideoSuperResolution(context.Background(), task)
	require.NoError(t, err)
	require.True(t, postProcessed)
	require.Equal(t, string(model.TaskStatusFailure), taskInfo.Status)
	require.Equal(t, "100%", taskInfo.Progress)
	require.Equal(t, "", taskInfo.Url)
	require.Equal(t, "super resolution failed", taskInfo.Reason)
	require.Equal(t, 4321, taskInfo.TotalTokens)

	state, ok, err := ReadVideoSuperResolutionState(task.Data)
	require.NoError(t, err)
	require.True(t, ok)
	require.Equal(t, "", state.ResultURL)
	require.Equal(t, "super resolution failed", state.LastError)
}

func TestPollVideoSuperResolutionFailsTaskWhenNoOutputURL(t *testing.T) {
	InitHttpClient()
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		require.Equal(t, "/api/v1/poll", r.URL.Path)
		w.Header().Set("Content-Type", "application/json")
		_, _ = w.Write([]byte(`{
			"metadata": {
				"task_id": "sr-task-2",
				"task_status": "COMPLETED",
				"business_code": "0",
				"error_msg": ""
			},
			"data": {}
		}`))
	}))
	defer server.Close()

	saved := map[string]string{}
	require.NoError(t, config.GlobalConfig.SaveToDB(func(key, value string) error {
		saved[key] = value
		return nil
	}))
	t.Cleanup(func() {
		require.NoError(t, config.GlobalConfig.LoadFromDB(saved))
	})

	require.NoError(t, config.GlobalConfig.LoadFromDB(map[string]string{
		"video_super_resolution.enabled":          "true",
		"video_super_resolution.base_url":         server.URL,
		"video_super_resolution.api_key":          "sr-api-key",
		"video_super_resolution.output_tos_path":  "tos://demo-bucket/video-sr",
		"video_super_resolution.operator_id":      "las_video_super_resolution",
		"video_super_resolution.operator_version": "v1",
	}))

	task := &model.Task{TaskID: "task-2"}
	original, err := common.Marshal(map[string]any{
		"usage": map[string]any{
			"total_tokens": 2468,
		},
	})
	require.NoError(t, err)

	task.Data, _ = common.Marshal(VideoSuperResolutionTaskData{
		Kind:     videoSuperResolutionTaskDataKind,
		Original: original,
		PostProcess: &VideoSuperResolutionState{
			TaskID:           "sr-task-2",
			Status:           "RUNNING",
			SourceURL:        "https://example.com/original-2.mp4",
			SourceResolution: "720p",
			TargetResolution: "1080p",
			TargetWidth:      1920,
		},
	})

	taskInfo, postProcessed, err := PollVideoSuperResolution(context.Background(), task)
	require.NoError(t, err)
	require.True(t, postProcessed)
	require.Equal(t, string(model.TaskStatusFailure), taskInfo.Status)
	require.Equal(t, "", taskInfo.Url)
	require.Equal(t, "video super resolution completed without output url", taskInfo.Reason)
	require.Equal(t, 2468, taskInfo.TotalTokens)

	state, ok, err := ReadVideoSuperResolutionState(task.Data)
	require.NoError(t, err)
	require.True(t, ok)
	require.Equal(t, "", state.ResultURL)
	require.Equal(t, "video super resolution completed without output url", state.LastError)
}

func TestPollVideoSuperResolutionStartsNormalizationFor720pOutput(t *testing.T) {
	InitHttpClient()
	var resizeSubmit map[string]any
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		switch r.URL.Path {
		case "/api/v1/poll":
			_, _ = w.Write([]byte(`{
				"metadata": {
					"task_id": "sr-task-3",
					"task_status": "COMPLETED",
					"business_code": "0",
					"error_msg": ""
				},
				"data": {
					"output_video_tos_url": "tos://demo-bucket/video-sr/source_1280x736.mp4"
				}
			}`))
		case "/api/v1/submit":
			body, err := io.ReadAll(r.Body)
			require.NoError(t, err)
			require.NoError(t, common.Unmarshal(body, &resizeSubmit))
			_, _ = w.Write([]byte(`{
				"metadata": {
					"task_id": "resize-task-3",
					"task_status": "PENDING",
					"business_code": "0",
					"error_msg": ""
				},
				"data": {}
			}`))
		default:
			t.Fatalf("unexpected path: %s", r.URL.Path)
		}
	}))
	defer server.Close()

	saved := map[string]string{}
	require.NoError(t, config.GlobalConfig.SaveToDB(func(key, value string) error {
		saved[key] = value
		return nil
	}))
	t.Cleanup(func() {
		require.NoError(t, config.GlobalConfig.LoadFromDB(saved))
	})

	require.NoError(t, config.GlobalConfig.LoadFromDB(map[string]string{
		"video_super_resolution.enabled":          "true",
		"video_super_resolution.base_url":         server.URL,
		"video_super_resolution.api_key":          "sr-api-key",
		"video_super_resolution.output_tos_path":  "tos://demo-bucket/video-sr",
		"video_super_resolution.operator_id":      "las_video_super_resolution",
		"video_super_resolution.operator_version": "v1",
	}))

	task := &model.Task{TaskID: "task-3"}
	original, err := common.Marshal(map[string]any{
		"usage": map[string]any{"total_tokens": 1357},
	})
	require.NoError(t, err)
	task.Data, err = common.Marshal(VideoSuperResolutionTaskData{
		Kind:     videoSuperResolutionTaskDataKind,
		Original: original,
		PostProcess: &VideoSuperResolutionState{
			TaskID:           "sr-task-3",
			Status:           "RUNNING",
			SourceResolution: "480p",
			TargetResolution: "720p",
			TargetWidth:      1280,
			TargetHeight:     720,
		},
	})
	require.NoError(t, err)

	taskInfo, postProcessed, err := PollVideoSuperResolution(context.Background(), task)
	require.NoError(t, err)
	require.True(t, postProcessed)
	require.Equal(t, string(model.TaskStatusInProgress), taskInfo.Status)
	require.Equal(t, "90%", taskInfo.Progress)
	require.Equal(t, "", taskInfo.Url)

	require.Equal(t, "las_video_resize", resizeSubmit["operator_id"])
	data := resizeSubmit["data"].(map[string]any)
	require.Equal(t, "tos://demo-bucket/video-sr/source_1280x736.mp4", data["video_path"])
	require.Equal(t, float64(1280), data["min_width"])
	require.Equal(t, float64(1280), data["max_width"])
	require.Equal(t, float64(720), data["min_height"])
	require.Equal(t, float64(720), data["max_height"])
	require.Equal(t, "disable", data["force_original_aspect_ratio_type"])

	state, ok, err := ReadVideoSuperResolutionState(task.Data)
	require.NoError(t, err)
	require.True(t, ok)
	require.Equal(t, "resize-task-3", state.NormalizeTaskID)
	require.Equal(t, "PENDING", state.NormalizeStatus)
}

func TestPollVideoSuperResolutionCompletesNormalization(t *testing.T) {
	InitHttpClient()
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		require.Equal(t, "/api/v1/poll", r.URL.Path)
		body, err := io.ReadAll(r.Body)
		require.NoError(t, err)
		var payload map[string]any
		require.NoError(t, common.Unmarshal(body, &payload))
		require.Equal(t, "las_video_resize", payload["operator_id"])
		w.Header().Set("Content-Type", "application/json")
		_, _ = w.Write([]byte(`{
			"metadata": {
				"task_id": "resize-task-4",
				"task_status": "COMPLETED",
				"business_code": "0",
				"error_msg": ""
			},
			"data": {
				"output_path": "tos://demo-bucket/video-sr/normalized/task-4_1280x720.mp4"
			}
		}`))
	}))
	defer server.Close()

	saved := map[string]string{}
	require.NoError(t, config.GlobalConfig.SaveToDB(func(key, value string) error {
		saved[key] = value
		return nil
	}))
	t.Cleanup(func() {
		require.NoError(t, config.GlobalConfig.LoadFromDB(saved))
	})

	require.NoError(t, config.GlobalConfig.LoadFromDB(map[string]string{
		"video_super_resolution.enabled":             "true",
		"video_super_resolution.base_url":            server.URL,
		"video_super_resolution.api_key":             "sr-api-key",
		"video_super_resolution.output_tos_path":     "tos://demo-bucket/video-sr",
		"video_super_resolution.operator_id":         "las_video_super_resolution",
		"video_super_resolution.operator_version":    "v1",
		"video_super_resolution.tos_public_base_url": "https://cdn.example.com/{bucket}/{key}",
	}))

	task := &model.Task{TaskID: "task-4"}
	original, err := common.Marshal(map[string]any{
		"usage": map[string]any{"total_tokens": 2468},
	})
	require.NoError(t, err)
	task.Data, err = common.Marshal(VideoSuperResolutionTaskData{
		Kind:     videoSuperResolutionTaskDataKind,
		Original: original,
		PostProcess: &VideoSuperResolutionState{
			TaskID:             "sr-task-4",
			Status:             "COMPLETED",
			SourceResolution:   "480p",
			TargetResolution:   "720p",
			TargetWidth:        1280,
			TargetHeight:       720,
			OutputTOSURL:       "tos://demo-bucket/video-sr/source_1280x736.mp4",
			NormalizeTaskID:    "resize-task-4",
			NormalizeStatus:    "RUNNING",
			NormalizeLastError: "",
		},
	})
	require.NoError(t, err)

	taskInfo, postProcessed, err := PollVideoSuperResolution(context.Background(), task)
	require.NoError(t, err)
	require.True(t, postProcessed)
	require.Equal(t, string(model.TaskStatusSuccess), taskInfo.Status)
	require.Equal(t, "100%", taskInfo.Progress)
	require.Equal(t, "https://cdn.example.com/demo-bucket/video-sr/normalized/task-4_1280x720.mp4", taskInfo.Url)
	require.Equal(t, 2468, taskInfo.TotalTokens)

	state, ok, err := ReadVideoSuperResolutionState(task.Data)
	require.NoError(t, err)
	require.True(t, ok)
	require.Equal(t, "COMPLETED", state.NormalizeStatus)
	require.Equal(t, "tos://demo-bucket/video-sr/normalized/task-4_1280x720.mp4", state.OutputTOSURL)
	require.Equal(t, taskInfo.Url, state.ResultURL)
}
