package main

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"os"
	"strings"
)

type changeItem struct {
	ChangeType  string `json:"change_type"`
	Endpoint    string `json:"endpoint"`
	Method      string `json:"method"`
	Section     string `json:"section"`
	Description string `json:"description"`
	Impact      string `json:"impact"`
}

type publishRequest struct {
	Version         string       `json:"version"`
	Title           string       `json:"title"`
	Summary         string       `json:"summary"`
	ChangedSections []string     `json:"changed_sections"`
	ChangeItems     []changeItem `json:"change_items"`
	Content         string       `json:"content"`
	SourceCommit    string       `json:"source_commit,omitempty"`
}

func requiredEnv(key string) string {
	value := strings.TrimSpace(os.Getenv(key))
	if value == "" {
		fmt.Fprintf(os.Stderr, "missing env: %s\n", key)
		os.Exit(1)
	}
	return value
}

func main() {
	baseURL := strings.TrimRight(requiredEnv("API_DOCS_BASE_URL"), "/")
	token := requiredEnv("API_DOCS_TOKEN")
	actorID := requiredEnv("API_DOCS_ACTOR_ID")
	version := requiredEnv("API_DOCS_VERSION")
	summary := requiredEnv("API_DOCS_SUMMARY")
	sections := strings.Split(requiredEnv("API_DOCS_CHANGED_SECTIONS"), ",")
	contentBytes, err := os.ReadFile("docs/api/current.md")
	if err != nil {
		fmt.Fprintln(os.Stderr, err)
		os.Exit(1)
	}
	requestBody := publishRequest{
		Version:         version,
		Title:           "8liangai.com API Docs",
		Summary:         summary,
		ChangedSections: sections,
		ChangeItems: []changeItem{
			{
				ChangeType:  "changed",
				Section:     "接口文档",
				Description: summary,
				Impact:      "兼容旧调用",
			},
		},
		Content:      string(contentBytes),
		SourceCommit: os.Getenv("API_DOCS_SOURCE_COMMIT"),
	}
	body, err := json.Marshal(requestBody)
	if err != nil {
		fmt.Fprintln(os.Stderr, err)
		os.Exit(1)
	}
	req, err := http.NewRequest(http.MethodPost, baseURL+"/api/docs/publish", bytes.NewReader(body))
	if err != nil {
		fmt.Fprintln(os.Stderr, err)
		os.Exit(1)
	}
	req.Header.Set("Authorization", "Bearer "+token)
	req.Header.Set("New-Api-User", actorID)
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Accept", "application/json")
	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		fmt.Fprintln(os.Stderr, err)
		os.Exit(1)
	}
	defer resp.Body.Close()
	respBody, _ := io.ReadAll(resp.Body)
	fmt.Println(string(respBody))
	if resp.StatusCode >= 400 {
		os.Exit(1)
	}
}
