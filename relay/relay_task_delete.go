package relay

import (
	"errors"
	"fmt"
	"io"
	"net/http"
	"strings"
	"time"

	"github.com/QuantumNous/new-api/constant"
	"github.com/QuantumNous/new-api/dto"
	"github.com/QuantumNous/new-api/model"
	"github.com/QuantumNous/new-api/relay/channel"
	relaycommon "github.com/QuantumNous/new-api/relay/common"
	"github.com/QuantumNous/new-api/service"
	"github.com/gin-gonic/gin"
)

type VideoDeleteResponse struct {
	ID      string `json:"id"`
	TaskID  string `json:"task_id,omitempty"`
	Object  string `json:"object"`
	Deleted bool   `json:"deleted"`
	Status  string `json:"status"`
}

func RelayTaskDelete(c *gin.Context) (*VideoDeleteResponse, *dto.TaskError) {
	taskID := c.Param("task_id")
	if strings.TrimSpace(taskID) == "" {
		return nil, service.TaskErrorWrapperLocal(fmt.Errorf("task_id is required"), "invalid_request", http.StatusBadRequest)
	}
	userID := c.GetInt("id")

	task, exist, err := model.GetByTaskId(userID, taskID)
	if err != nil {
		return nil, service.TaskErrorWrapper(err, "get_task_failed", http.StatusInternalServerError)
	}
	if !exist {
		return nil, service.TaskErrorWrapperLocal(errors.New("task_not_exist"), "task_not_exist", http.StatusNotFound)
	}

	adaptor := GetTaskAdaptor(task.Platform)
	if adaptor == nil {
		return nil, service.TaskErrorWrapperLocal(fmt.Errorf("task platform %s does not support delete", task.Platform), "task_delete_not_supported", http.StatusNotImplemented)
	}
	deleter, ok := adaptor.(channel.TaskDeleteAdaptor)
	if !ok {
		return nil, service.TaskErrorWrapperLocal(fmt.Errorf("task platform %s does not support delete", task.Platform), "task_delete_not_supported", http.StatusNotImplemented)
	}

	ch, err := model.GetChannelById(task.ChannelId, true)
	if err != nil {
		return nil, service.TaskErrorWrapper(err, "get_task_channel_failed", http.StatusInternalServerError)
	}
	apiKey, _, keyErr := ch.GetNextEnabledKey()
	if keyErr != nil {
		return nil, service.TaskErrorWrapper(keyErr, "get_task_channel_key_failed", http.StatusBadRequest)
	}

	info := &relaycommon.RelayInfo{
		UserId: userID,
		OriginModelName: func() string {
			if task.Properties.OriginModelName != "" {
				return task.Properties.OriginModelName
			}
			return task.Properties.UpstreamModelName
		}(),
		ChannelMeta: &relaycommon.ChannelMeta{
			ChannelId:      task.ChannelId,
			ChannelType:    ch.Type,
			ChannelBaseUrl: ch.GetBaseURL(),
			ApiKey:         apiKey,
		},
	}
	adaptor.Init(info)

	latestStatus, fetchErr := fetchLatestTaskStatusForDelete(task, adaptor, ch, apiKey)
	if fetchErr != nil {
		latestStatus = task.Status
	}

	switch latestStatus {
	case model.TaskStatusInProgress:
		return nil, service.TaskErrorWrapperLocal(fmt.Errorf("running task does not support delete"), "task_delete_not_supported_for_running", http.StatusBadRequest)
	case model.TaskStatusCancelled:
		return nil, service.TaskErrorWrapperLocal(fmt.Errorf("task is already cancelled"), "task_already_cancelled", http.StatusBadRequest)
	}

	if taskErr := deleter.DeleteTask(c, info, task); taskErr != nil {
		return nil, taskErr
	}

	resp := &VideoDeleteResponse{
		ID:     task.TaskID,
		TaskID: task.TaskID,
		Object: "video",
	}

	if shouldCancelTask(latestStatus) {
		if err = markTaskCancelled(c, task); err != nil {
			return nil, service.TaskErrorWrapper(err, "mark_task_cancelled_failed", http.StatusInternalServerError)
		}
		resp.Deleted = false
		resp.Status = dto.VideoStatusCancelled
		return resp, nil
	}

	if err = deleteLocalTask(task); err != nil {
		return nil, service.TaskErrorWrapper(err, "delete_local_task_failed", http.StatusInternalServerError)
	}
	resp.Deleted = true
	resp.Status = "deleted"
	return resp, nil
}

func fetchLatestTaskStatusForDelete(task *model.Task, adaptor channel.TaskAdaptor, ch *model.Channel, apiKey string) (model.TaskStatus, error) {
	baseURL := constant.ChannelBaseURLs[ch.Type]
	if channelBaseURL := ch.GetBaseURL(); channelBaseURL != "" {
		baseURL = channelBaseURL
	}
	resp, err := adaptor.FetchTask(baseURL, apiKey, map[string]any{
		"task_id": task.GetUpstreamTaskID(),
		"action":  task.Action,
	}, ch.GetSetting().Proxy)
	if err != nil {
		return "", err
	}
	if resp == nil {
		return "", fmt.Errorf("fetch task response is nil")
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		body, _ := io.ReadAll(resp.Body)
		return "", fmt.Errorf("fetch task failed: %s", strings.TrimSpace(string(body)))
	}

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return "", err
	}
	ti, err := adaptor.ParseTaskResult(body)
	if err != nil || ti == nil || strings.TrimSpace(ti.Status) == "" {
		return "", fmt.Errorf("parse latest task status failed")
	}
	return model.TaskStatus(ti.Status), nil
}

func shouldCancelTask(status model.TaskStatus) bool {
	switch status {
	case model.TaskStatusNotStart, model.TaskStatusSubmitted, model.TaskStatusQueued, model.TaskStatusUnknown:
		return true
	default:
		return false
	}
}

func markTaskCancelled(c *gin.Context, task *model.Task) error {
	preStatus := task.Status
	now := time.Now().Unix()

	task.Status = model.TaskStatusCancelled
	task.Progress = "100%"
	task.FinishTime = now
	task.UpdatedAt = now
	task.FailReason = "task cancelled"
	task.PrivateData.ResultURL = ""
	task.SetData(map[string]any{
		"status": "cancelled",
	})

	if err := task.Update(); err != nil {
		return err
	}

	if task.Quota != 0 && preStatus != model.TaskStatusCancelled && preStatus != model.TaskStatusFailure && preStatus != model.TaskStatusSuccess {
		service.RefundTaskQuota(c.Request.Context(), task, task.FailReason)
	}
	return nil
}

func deleteLocalTask(task *model.Task) error {
	return model.DB.Delete(task).Error
}
