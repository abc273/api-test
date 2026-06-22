package model

import (
	"testing"

	"github.com/QuantumNous/new-api/common"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func resetApiDocTestState(t *testing.T) {
	t.Helper()
	require.NoError(t, DB.Exec("DELETE FROM api_doc_change_items").Error)
	require.NoError(t, DB.Exec("DELETE FROM api_doc_revisions").Error)
	require.NoError(t, DB.Exec("DELETE FROM options").Error)
	common.OptionMapRWMutex.Lock()
	common.OptionMap = map[string]string{}
	common.OptionMapRWMutex.Unlock()
}

func validApiDocPublishInput() ApiDocPublishInput {
	return ApiDocPublishInput{
		Version:         "2026.06.16.1",
		Title:           "8liangai.com API Docs",
		Summary:         "发布文档版本记录",
		ChangedSections: []string{"文档发布接口"},
		Content:         "# 八两 API 接口文档\n\n完整请求示例\n",
		PublishedBy:     1,
		ChangeItems: []ApiDocChangeItemInput{
			{
				ChangeType:  ApiDocChangeTypeAdded,
				Endpoint:    "/api/docs/meta",
				Method:      "GET",
				Section:     "文档版本信息",
				Description: "新增文档版本信息接口",
				Impact:      "兼容旧调用",
			},
		},
	}
}

func TestPublishApiDocRevisionCreatesRevisionAndUpdatesOption(t *testing.T) {
	resetApiDocTestState(t)

	view, err := PublishApiDocRevision(validApiDocPublishInput())
	require.NoError(t, err)
	require.NotNil(t, view)
	assert.Equal(t, "2026.06.16.1", view.Version)
	assert.Equal(t, []string{"文档发布接口"}, view.ChangedSections)
	require.Len(t, view.ChangeItems, 1)
	assert.Equal(t, "GET", view.ChangeItems[0].Method)

	latest, err := GetLatestApiDocRevision(true)
	require.NoError(t, err)
	assert.Equal(t, view.Version, latest.Version)
	assert.Contains(t, latest.Content, "八两 API")

	common.OptionMapRWMutex.RLock()
	assert.Equal(t, validApiDocPublishInput().Content, common.OptionMap["ApiDocs"])
	common.OptionMapRWMutex.RUnlock()
}

func TestPublishApiDocRevisionRejectsDuplicateVersion(t *testing.T) {
	resetApiDocTestState(t)

	_, err := PublishApiDocRevision(validApiDocPublishInput())
	require.NoError(t, err)
	_, err = PublishApiDocRevision(validApiDocPublishInput())
	require.Error(t, err)
	assert.Contains(t, err.Error(), "version already exists")
}

func TestPublishApiDocRevisionRejectsForbiddenWording(t *testing.T) {
	resetApiDocTestState(t)

	input := validApiDocPublishInput()
	input.Content = "这里包含" + string([]rune{'\u7528', '\u6237'}) + "称呼"
	_, err := PublishApiDocRevision(input)
	require.Error(t, err)
	assert.Contains(t, err.Error(), "forbidden")
}

func TestBuildApiDocDiffCountsChangedLines(t *testing.T) {
	diff := BuildApiDocDiff("old", "a\nb\nc", "new", "a\nbb\nc\nd")

	assert.True(t, diff.Changed)
	assert.Equal(t, 2, diff.AddedLines)
	assert.Equal(t, 1, diff.RemovedLines)
	assert.Equal(t, "old", diff.FromVersion)
	assert.Equal(t, "new", diff.ToVersion)
}
