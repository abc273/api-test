package model

import (
	"crypto/sha256"
	"errors"
	"fmt"
	"strings"
	"time"

	"github.com/QuantumNous/new-api/common"
	"gorm.io/gorm"
)

const ApiDocChangeTypeAdded = "added"
const ApiDocChangeTypeChanged = "changed"
const ApiDocChangeTypeDeprecated = "deprecated"
const ApiDocChangeTypeRemoved = "removed"
const ApiDocChangeTypeFixed = "fixed"

var ErrApiDocRevisionNotFound = errors.New("api doc revision not found")

var forbiddenApiDocTerms = []string{"\u5ba2\u6237", "\u7528\u6237"}

type ApiDocRevision struct {
	Id              int    `json:"id" gorm:"primaryKey"`
	Version         string `json:"version" gorm:"uniqueIndex;size:64;not null"`
	Title           string `json:"title" gorm:"size:255"`
	Summary         string `json:"summary" gorm:"type:text"`
	ChangedSections string `json:"changed_sections" gorm:"type:text"`
	Content         string `json:"content" gorm:"type:text"`
	ContentSHA256   string `json:"content_sha256" gorm:"size:64;index"`
	SourceCommit    string `json:"source_commit" gorm:"size:128"`
	PublishedBy     int    `json:"published_by" gorm:"index"`
	PublishedAt     int64  `json:"published_at" gorm:"index"`
	CreatedTime     int64  `json:"created_time"`
}

type ApiDocChangeItem struct {
	Id          int    `json:"id" gorm:"primaryKey"`
	RevisionID  int    `json:"revision_id" gorm:"index;not null"`
	ChangeType  string `json:"change_type" gorm:"size:32;index;not null"`
	Endpoint    string `json:"endpoint" gorm:"size:255"`
	Method      string `json:"method" gorm:"size:16"`
	Section     string `json:"section" gorm:"size:255"`
	Description string `json:"description" gorm:"type:text"`
	Impact      string `json:"impact" gorm:"type:text"`
	CreatedTime int64  `json:"created_time"`
}

type ApiDocChangeItemInput struct {
	ChangeType  string `json:"change_type"`
	Endpoint    string `json:"endpoint"`
	Method      string `json:"method"`
	Section     string `json:"section"`
	Description string `json:"description"`
	Impact      string `json:"impact"`
}

type ApiDocPublishInput struct {
	Version         string
	Title           string
	Summary         string
	ChangedSections []string
	Content         string
	SourceCommit    string
	PublishedBy     int
	ChangeItems     []ApiDocChangeItemInput
}

type ApiDocRevisionView struct {
	Id              int                    `json:"id"`
	Version         string                 `json:"version"`
	Title           string                 `json:"title"`
	Summary         string                 `json:"summary"`
	ChangedSections []string               `json:"changed_sections"`
	Content         string                 `json:"content,omitempty"`
	ContentSHA256   string                 `json:"content_sha256"`
	SourceCommit    string                 `json:"source_commit"`
	PublishedBy     int                    `json:"published_by"`
	PublishedAt     int64                  `json:"published_at"`
	CreatedTime     int64                  `json:"created_time"`
	ChangeItems     []ApiDocChangeItemView `json:"change_items,omitempty"`
}

type ApiDocChangeItemView struct {
	Id          int    `json:"id"`
	RevisionID  int    `json:"revision_id"`
	ChangeType  string `json:"change_type"`
	Endpoint    string `json:"endpoint"`
	Method      string `json:"method"`
	Section     string `json:"section"`
	Description string `json:"description"`
	Impact      string `json:"impact"`
	CreatedTime int64  `json:"created_time"`
}

type ApiDocDiffLine struct {
	Type string `json:"type"`
	Text string `json:"text"`
}

type ApiDocDiffResult struct {
	FromVersion  string           `json:"from_version"`
	ToVersion    string           `json:"to_version"`
	Changed      bool             `json:"changed"`
	AddedLines   int              `json:"added_lines"`
	RemovedLines int              `json:"removed_lines"`
	Lines        []ApiDocDiffLine `json:"lines"`
}

func apiDocSHA256(content string) string {
	sum := sha256.Sum256([]byte(content))
	return fmt.Sprintf("%x", sum[:])
}

func normalizeApiDocString(value string) string {
	return strings.TrimSpace(value)
}

func normalizeApiDocChangeType(value string) string {
	switch strings.ToLower(strings.TrimSpace(value)) {
	case ApiDocChangeTypeAdded:
		return ApiDocChangeTypeAdded
	case ApiDocChangeTypeChanged:
		return ApiDocChangeTypeChanged
	case ApiDocChangeTypeDeprecated:
		return ApiDocChangeTypeDeprecated
	case ApiDocChangeTypeRemoved:
		return ApiDocChangeTypeRemoved
	case ApiDocChangeTypeFixed:
		return ApiDocChangeTypeFixed
	default:
		return ""
	}
}

func validateApiDocPublishInput(input ApiDocPublishInput) error {
	if normalizeApiDocString(input.Version) == "" {
		return errors.New("version is required")
	}
	if normalizeApiDocString(input.Title) == "" {
		return errors.New("title is required")
	}
	if normalizeApiDocString(input.Summary) == "" {
		return errors.New("summary is required")
	}
	if len(input.ChangedSections) == 0 {
		return errors.New("changed_sections is required")
	}
	if normalizeApiDocString(input.Content) == "" {
		return errors.New("content is required")
	}
	for _, term := range forbiddenApiDocTerms {
		if strings.Contains(input.Content, term) {
			return errors.New("content contains forbidden wording")
		}
	}
	if len(input.ChangeItems) == 0 {
		return errors.New("change_items is required")
	}
	for _, item := range input.ChangeItems {
		if normalizeApiDocChangeType(item.ChangeType) == "" {
			return errors.New("change_items contains invalid change_type")
		}
		if normalizeApiDocString(item.Description) == "" {
			return errors.New("change_items contains empty description")
		}
	}
	return nil
}

func marshalApiDocChangedSections(sections []string) (string, error) {
	normalized := make([]string, 0, len(sections))
	for _, section := range sections {
		section = normalizeApiDocString(section)
		if section != "" {
			normalized = append(normalized, section)
		}
	}
	if len(normalized) == 0 {
		return "", errors.New("changed_sections is required")
	}
	bytes, err := common.Marshal(normalized)
	if err != nil {
		return "", err
	}
	return string(bytes), nil
}

func parseApiDocChangedSections(raw string) []string {
	if normalizeApiDocString(raw) == "" {
		return []string{}
	}
	var sections []string
	if err := common.UnmarshalJsonStr(raw, &sections); err != nil {
		return []string{}
	}
	return sections
}

func buildApiDocRevisionView(revision ApiDocRevision, items []ApiDocChangeItem, includeContent bool) ApiDocRevisionView {
	itemViews := make([]ApiDocChangeItemView, 0, len(items))
	for _, item := range items {
		itemViews = append(itemViews, ApiDocChangeItemView{
			Id:          item.Id,
			RevisionID:  item.RevisionID,
			ChangeType:  item.ChangeType,
			Endpoint:    item.Endpoint,
			Method:      item.Method,
			Section:     item.Section,
			Description: item.Description,
			Impact:      item.Impact,
			CreatedTime: item.CreatedTime,
		})
	}
	view := ApiDocRevisionView{
		Id:              revision.Id,
		Version:         revision.Version,
		Title:           revision.Title,
		Summary:         revision.Summary,
		ChangedSections: parseApiDocChangedSections(revision.ChangedSections),
		ContentSHA256:   revision.ContentSHA256,
		SourceCommit:    revision.SourceCommit,
		PublishedBy:     revision.PublishedBy,
		PublishedAt:     revision.PublishedAt,
		CreatedTime:     revision.CreatedTime,
		ChangeItems:     itemViews,
	}
	if includeContent {
		view.Content = revision.Content
	}
	return view
}

func PublishApiDocRevision(input ApiDocPublishInput) (*ApiDocRevisionView, error) {
	if err := validateApiDocPublishInput(input); err != nil {
		return nil, err
	}
	changedSections, err := marshalApiDocChangedSections(input.ChangedSections)
	if err != nil {
		return nil, err
	}
	now := time.Now().Unix()
	revision := ApiDocRevision{
		Version:         normalizeApiDocString(input.Version),
		Title:           normalizeApiDocString(input.Title),
		Summary:         normalizeApiDocString(input.Summary),
		ChangedSections: changedSections,
		Content:         input.Content,
		ContentSHA256:   apiDocSHA256(input.Content),
		SourceCommit:    normalizeApiDocString(input.SourceCommit),
		PublishedBy:     input.PublishedBy,
		PublishedAt:     now,
		CreatedTime:     now,
	}

	var view ApiDocRevisionView
	err = DB.Transaction(func(tx *gorm.DB) error {
		var existing int64
		if err := tx.Model(&ApiDocRevision{}).Where("version = ?", revision.Version).Count(&existing).Error; err != nil {
			return err
		}
		if existing > 0 {
			return errors.New("version already exists")
		}
		if err := tx.Create(&revision).Error; err != nil {
			return err
		}
		items := make([]ApiDocChangeItem, 0, len(input.ChangeItems))
		for _, item := range input.ChangeItems {
			items = append(items, ApiDocChangeItem{
				RevisionID:  revision.Id,
				ChangeType:  normalizeApiDocChangeType(item.ChangeType),
				Endpoint:    normalizeApiDocString(item.Endpoint),
				Method:      strings.ToUpper(normalizeApiDocString(item.Method)),
				Section:     normalizeApiDocString(item.Section),
				Description: normalizeApiDocString(item.Description),
				Impact:      normalizeApiDocString(item.Impact),
				CreatedTime: now,
			})
		}
		if err := tx.Create(&items).Error; err != nil {
			return err
		}
		view = buildApiDocRevisionView(revision, items, true)
		return nil
	})
	if err != nil {
		return nil, err
	}
	if err := UpdateOption("ApiDocs", input.Content); err != nil {
		return nil, err
	}
	return &view, nil
}

func GetLatestApiDocRevision(includeContent bool) (*ApiDocRevisionView, error) {
	var revision ApiDocRevision
	err := DB.Order("published_at desc, id desc").First(&revision).Error
	if errors.Is(err, gorm.ErrRecordNotFound) {
		return nil, ErrApiDocRevisionNotFound
	}
	if err != nil {
		return nil, err
	}
	items, err := GetApiDocChangeItems(revision.Id)
	if err != nil {
		return nil, err
	}
	view := buildApiDocRevisionView(revision, items, includeContent)
	return &view, nil
}

func GetApiDocRevisionByVersion(version string, includeContent bool) (*ApiDocRevisionView, error) {
	var revision ApiDocRevision
	err := DB.Where("version = ?", version).First(&revision).Error
	if errors.Is(err, gorm.ErrRecordNotFound) {
		return nil, ErrApiDocRevisionNotFound
	}
	if err != nil {
		return nil, err
	}
	items, err := GetApiDocChangeItems(revision.Id)
	if err != nil {
		return nil, err
	}
	view := buildApiDocRevisionView(revision, items, includeContent)
	return &view, nil
}

func GetPreviousApiDocRevision(version string) (*ApiDocRevision, error) {
	var current ApiDocRevision
	if err := DB.Where("version = ?", version).First(&current).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, ErrApiDocRevisionNotFound
		}
		return nil, err
	}
	var previous ApiDocRevision
	err := DB.Where("published_at < ? OR (published_at = ? AND id < ?)", current.PublishedAt, current.PublishedAt, current.Id).
		Order("published_at desc, id desc").
		First(&previous).Error
	if errors.Is(err, gorm.ErrRecordNotFound) {
		return nil, ErrApiDocRevisionNotFound
	}
	if err != nil {
		return nil, err
	}
	return &previous, nil
}

func GetApiDocChangeItems(revisionID int) ([]ApiDocChangeItem, error) {
	var items []ApiDocChangeItem
	err := DB.Where("revision_id = ?", revisionID).Order("id asc").Find(&items).Error
	return items, err
}

func ListApiDocRevisions(startIdx int, pageSize int, includeContent bool) ([]ApiDocRevisionView, int64, error) {
	var total int64
	if err := DB.Model(&ApiDocRevision{}).Count(&total).Error; err != nil {
		return nil, 0, err
	}
	var revisions []ApiDocRevision
	if err := DB.Order("published_at desc, id desc").Offset(startIdx).Limit(pageSize).Find(&revisions).Error; err != nil {
		return nil, 0, err
	}
	views := make([]ApiDocRevisionView, 0, len(revisions))
	for _, revision := range revisions {
		items, err := GetApiDocChangeItems(revision.Id)
		if err != nil {
			return nil, 0, err
		}
		views = append(views, buildApiDocRevisionView(revision, items, includeContent))
	}
	return views, total, nil
}

func BuildApiDocDiff(fromVersion string, fromContent string, toVersion string, toContent string) ApiDocDiffResult {
	fromLines := strings.Split(fromContent, "\n")
	toLines := strings.Split(toContent, "\n")
	maxLen := len(fromLines)
	if len(toLines) > maxLen {
		maxLen = len(toLines)
	}
	result := ApiDocDiffResult{
		FromVersion: fromVersion,
		ToVersion:   toVersion,
		Lines:       make([]ApiDocDiffLine, 0),
	}
	for i := 0; i < maxLen; i++ {
		var fromLine string
		var toLine string
		hasFrom := i < len(fromLines)
		hasTo := i < len(toLines)
		if hasFrom {
			fromLine = fromLines[i]
		}
		if hasTo {
			toLine = toLines[i]
		}
		switch {
		case hasFrom && hasTo && fromLine == toLine:
			if strings.TrimSpace(fromLine) != "" {
				result.Lines = append(result.Lines, ApiDocDiffLine{Type: "context", Text: fromLine})
			}
		case hasFrom && hasTo:
			result.RemovedLines++
			result.AddedLines++
			result.Lines = append(result.Lines, ApiDocDiffLine{Type: "removed", Text: fromLine})
			result.Lines = append(result.Lines, ApiDocDiffLine{Type: "added", Text: toLine})
		case hasFrom:
			result.RemovedLines++
			result.Lines = append(result.Lines, ApiDocDiffLine{Type: "removed", Text: fromLine})
		case hasTo:
			result.AddedLines++
			result.Lines = append(result.Lines, ApiDocDiffLine{Type: "added", Text: toLine})
		}
	}
	result.Changed = result.AddedLines > 0 || result.RemovedLines > 0
	return result
}
