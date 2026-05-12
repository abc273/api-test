package service

import (
	"bytes"
	"crypto/hmac"
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"strings"
	"time"

	"github.com/QuantumNous/new-api/common"
	"github.com/QuantumNous/new-api/setting/portrait_setting"
)

const (
	volcPortraitService = "ark"
	volcPortraitVersion = "2024-01-01"
)

type VolcPortraitValidateSession struct {
	BytedToken string `json:"BytedToken"`
	H5Link     string `json:"H5Link"`
}

type VolcPortraitAssetInfo struct {
	Id          string `json:"Id"`
	GroupId     string `json:"GroupId"`
	Status      string `json:"Status"`
	CreateTime  string `json:"CreateTime"`
	UpdateTime  string `json:"UpdateTime"`
	AssetType   string `json:"AssetType"`
	ProjectName string `json:"ProjectName"`
	Name        string `json:"Name"`
	URL         string `json:"URL"`
}

type volcPortraitConfig struct {
	AccessKey   string
	SecretKey   string
	Region      string
	ProjectName string
	Host        string
	Scheme      string
}

func GetVolcPortraitProjectName() string {
	return portrait_setting.GetProjectName()
}

func IsVolcPortraitConfigured() bool {
	cfg := getVolcPortraitConfig()
	return cfg.AccessKey != "" && cfg.SecretKey != ""
}

func getVolcPortraitConfig() volcPortraitConfig {
	host := strings.TrimSpace(common.GetEnvOrDefaultString("VOLC_PORTRAIT_OPENAPI_HOST", "open.volcengineapi.com"))
	scheme := strings.TrimSpace(common.GetEnvOrDefaultString("VOLC_PORTRAIT_OPENAPI_SCHEME", "https"))
	if scheme == "" {
		scheme = "https"
	}
	return volcPortraitConfig{
		AccessKey:   portrait_setting.GetAccessKey(),
		SecretKey:   portrait_setting.GetSecretKey(),
		Region:      portrait_setting.GetRegion(),
		ProjectName: GetVolcPortraitProjectName(),
		Host:        host,
		Scheme:      scheme,
	}
}

func GetVolcPortraitCallbackBaseURL() string {
	return portrait_setting.GetCallbackBaseURL()
}

func CreateVolcPortraitValidateSession(callbackURL string, projectName string) (*VolcPortraitValidateSession, error) {
	payload := map[string]any{
		"CallbackURL": strings.TrimSpace(callbackURL),
		"ProjectName": resolveVolcPortraitProjectName(projectName),
	}
	var result VolcPortraitValidateSession
	if err := callVolcPortraitOpenAPI("CreateVisualValidateSession", payload, &result); err != nil {
		return nil, err
	}
	if result.BytedToken == "" || result.H5Link == "" {
		return nil, fmt.Errorf("volc portrait validation session response is incomplete")
	}
	return &result, nil
}

func GetVolcPortraitValidateResult(bytedToken string, projectName string) (string, error) {
	payload := map[string]any{
		"BytedToken":  strings.TrimSpace(bytedToken),
		"ProjectName": resolveVolcPortraitProjectName(projectName),
	}
	var result struct {
		GroupId string `json:"GroupId"`
	}
	if err := callVolcPortraitOpenAPI("GetVisualValidateResult", payload, &result); err != nil {
		return "", err
	}
	if strings.TrimSpace(result.GroupId) == "" {
		return "", fmt.Errorf("volc portrait validation result has no group id")
	}
	return strings.TrimSpace(result.GroupId), nil
}

func CreateVolcPortraitAsset(groupID string, assetURL string, assetType string, name string, projectName string) (string, error) {
	assetType = normalizeVolcPortraitAssetType(assetType)
	payload := map[string]any{
		"GroupId":     strings.TrimSpace(groupID),
		"URL":         strings.TrimSpace(assetURL),
		"AssetType":   assetType,
		"ProjectName": resolveVolcPortraitProjectName(projectName),
	}
	if strings.TrimSpace(name) != "" {
		payload["Name"] = strings.TrimSpace(name)
	}
	var result struct {
		Id string `json:"Id"`
	}
	if err := callVolcPortraitOpenAPI("CreateAsset", payload, &result); err != nil {
		return "", err
	}
	if strings.TrimSpace(result.Id) == "" {
		return "", fmt.Errorf("volc portrait create asset response has no asset id")
	}
	return strings.TrimSpace(result.Id), nil
}

func GetVolcPortraitAsset(assetID string, projectName string) (*VolcPortraitAssetInfo, error) {
	payload := map[string]any{
		"Id":          strings.TrimSpace(assetID),
		"ProjectName": resolveVolcPortraitProjectName(projectName),
	}
	var result VolcPortraitAssetInfo
	if err := callVolcPortraitOpenAPI("GetAsset", payload, &result); err != nil {
		return nil, err
	}
	if result.Id == "" {
		result.Id = strings.TrimSpace(assetID)
	}
	return &result, nil
}

func resolveVolcPortraitProjectName(projectName string) string {
	projectName = strings.TrimSpace(projectName)
	if projectName != "" {
		return projectName
	}
	return GetVolcPortraitProjectName()
}

func normalizeVolcPortraitAssetType(assetType string) string {
	switch strings.ToLower(strings.TrimSpace(assetType)) {
	case "video":
		return "Video"
	case "audio":
		return "Audio"
	default:
		return "Image"
	}
}

func callVolcPortraitOpenAPI(action string, payload any, result any) error {
	cfg := getVolcPortraitConfig()
	if cfg.AccessKey == "" || cfg.SecretKey == "" {
		return fmt.Errorf("volc portrait AK/SK is not configured")
	}

	body, err := common.Marshal(payload)
	if err != nil {
		return err
	}

	query := url.Values{}
	query.Set("Action", action)
	query.Set("Version", volcPortraitVersion)
	endpoint := fmt.Sprintf("%s://%s/?%s", cfg.Scheme, cfg.Host, query.Encode())

	req, err := http.NewRequest(http.MethodPost, endpoint, bytes.NewReader(body))
	if err != nil {
		return err
	}
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Accept", "application/json")

	signVolcPortraitRequest(req, body, cfg)

	client := &http.Client{Timeout: 60 * time.Second}
	resp, err := client.Do(req)
	if err != nil {
		return err
	}
	defer resp.Body.Close()

	responseBody, err := io.ReadAll(resp.Body)
	if err != nil {
		return err
	}
	if resp.StatusCode < http.StatusOK || resp.StatusCode >= http.StatusMultipleChoices {
		return fmt.Errorf("volc portrait %s failed: HTTP %d: %s", action, resp.StatusCode, string(responseBody))
	}
	return parseVolcPortraitResult(action, responseBody, result)
}

func signVolcPortraitRequest(req *http.Request, body []byte, cfg volcPortraitConfig) {
	now := time.Now().UTC()
	date := now.Format("20060102")
	xDate := now.Format("20060102T150405Z")
	payloadHash := sha256Hex(body)

	req.Host = cfg.Host
	req.Header.Set("Host", cfg.Host)
	req.Header.Set("X-Date", xDate)
	req.Header.Set("X-Content-Sha256", payloadHash)

	signedHeaders := "content-type;host;x-content-sha256;x-date"
	canonicalHeaders := strings.Join([]string{
		"content-type:application/json",
		"host:" + cfg.Host,
		"x-content-sha256:" + payloadHash,
		"x-date:" + xDate,
		"",
	}, "\n")
	canonicalRequest := strings.Join([]string{
		req.Method,
		"/",
		req.URL.RawQuery,
		canonicalHeaders,
		signedHeaders,
		payloadHash,
	}, "\n")

	credentialScope := strings.Join([]string{date, cfg.Region, volcPortraitService, "request"}, "/")
	stringToSign := strings.Join([]string{
		"HMAC-SHA256",
		xDate,
		credentialScope,
		sha256Hex([]byte(canonicalRequest)),
	}, "\n")
	signingKey := volcPortraitSigningKey(cfg.SecretKey, date, cfg.Region, volcPortraitService)
	signature := hex.EncodeToString(hmacSHA256(signingKey, stringToSign))
	req.Header.Set("Authorization", fmt.Sprintf(
		"HMAC-SHA256 Credential=%s/%s, SignedHeaders=%s, Signature=%s",
		cfg.AccessKey,
		credentialScope,
		signedHeaders,
		signature,
	))
}

func parseVolcPortraitResult(action string, responseBody []byte, result any) error {
	var wrapped struct {
		Result           json.RawMessage `json:"Result"`
		ResponseMetadata struct {
			Error *struct {
				Code    string `json:"Code"`
				Message string `json:"Message"`
			} `json:"Error"`
		} `json:"ResponseMetadata"`
	}
	if err := common.Unmarshal(responseBody, &wrapped); err != nil {
		return err
	}
	if wrapped.ResponseMetadata.Error != nil {
		return fmt.Errorf("volc portrait %s failed: %s %s", action, wrapped.ResponseMetadata.Error.Code, wrapped.ResponseMetadata.Error.Message)
	}
	if len(wrapped.Result) > 0 && string(wrapped.Result) != "null" {
		return common.Unmarshal(wrapped.Result, result)
	}
	return common.Unmarshal(responseBody, result)
}

func volcPortraitSigningKey(secretKey string, date string, region string, service string) []byte {
	kDate := hmacSHA256([]byte(secretKey), date)
	kRegion := hmacSHA256(kDate, region)
	kService := hmacSHA256(kRegion, service)
	return hmacSHA256(kService, "request")
}

func hmacSHA256(key []byte, data string) []byte {
	mac := hmac.New(sha256.New, key)
	_, _ = mac.Write([]byte(data))
	return mac.Sum(nil)
}

func sha256Hex(data []byte) string {
	sum := sha256.Sum256(data)
	return hex.EncodeToString(sum[:])
}
