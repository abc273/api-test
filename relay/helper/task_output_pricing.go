package helper

import (
	"fmt"
	"strings"

	"github.com/QuantumNous/new-api/common"
	relaycommon "github.com/QuantumNous/new-api/relay/common"
	"github.com/QuantumNous/new-api/service"
	"github.com/QuantumNous/new-api/setting/billing_setting"
	"github.com/gin-gonic/gin"
)

func buildTaskPricingProfile(c *gin.Context) (billing_setting.TaskPricingProfile, error) {
	req, err := relaycommon.GetTaskRequest(c)
	if err != nil {
		return billing_setting.TaskPricingProfile{}, err
	}
	return billing_setting.TaskPricingProfile{
		Resolution:    resolveTaskPricingResolution(req),
		HasVideoInput: hasVideoInput(req.Metadata),
	}, nil
}

func resolveTaskPricingResolution(req relaycommon.TaskSubmitReq) string {
	if req.Metadata != nil {
		if raw, ok := req.Metadata["resolution"]; ok {
			if resolution := billing_setting.NormalizeResolution(fmt.Sprint(raw)); resolution != "" {
				return resolution
			}
		}
	}
	if resolution := billing_setting.NormalizeResolution(req.Resolution); resolution != "" {
		return resolution
	}
	if resolution := billing_setting.NormalizeResolution(req.Size); resolution != "" {
		return resolution
	}

	switch strings.TrimSpace(req.Model) {
	case service.Seedance2ModelAlias, service.Seedance2SRModelAlias, service.SD20FastModelAlias, service.SD20FastSRModelAlias:
		// These aliases are billed by output tier. Defaulting to 720p keeps the
		// request on a valid pricing tier when the client omits resolution.
		return "720p"
	default:
		return ""
	}
}

func hasVideoInput(metadata map[string]interface{}) bool {
	if metadata == nil {
		return false
	}
	contentRaw, ok := metadata["content"]
	if !ok {
		return false
	}
	contentItems, ok := contentRaw.([]interface{})
	if !ok {
		return false
	}
	for _, item := range contentItems {
		itemMap, ok := item.(map[string]interface{})
		if !ok {
			continue
		}
		if strings.EqualFold(fmt.Sprint(itemMap["type"]), "video_url") {
			return true
		}
		if _, has := itemMap["video_url"]; has {
			return true
		}
	}
	return false
}

func outputTierPriceToModelRatio(inputPrice float64) float64 {
	if inputPrice <= 0 {
		return 0
	}
	return inputPrice / 2
}

func buildOutputTierPriceData(inputPrice float64, groupRatio float64) (modelRatio float64, quota int) {
	modelRatio = outputTierPriceToModelRatio(inputPrice)
	quota = int(modelRatio / 2 * common.QuotaPerUnit * groupRatio)
	return modelRatio, quota
}
