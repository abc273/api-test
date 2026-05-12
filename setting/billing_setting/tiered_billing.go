package billing_setting

import (
	"fmt"
	"strings"

	"github.com/QuantumNous/new-api/pkg/billingexpr"
	"github.com/QuantumNous/new-api/setting/config"
	"github.com/samber/lo"
)

const (
	BillingModeRatio           = "ratio"
	BillingModeTieredExpr      = "tiered_expr"
	BillingModeOutputTierPrice = "output_tier_price"
	BillingModeField           = "billing_mode"
	BillingExprField           = "billing_expr"
	OutputTierPricingField     = "output_tier_pricing"
)

type OutputTierPricing struct {
	Label         string `json:"label,omitempty"`
	Resolution    string `json:"resolution,omitempty"`
	HasVideoInput *bool  `json:"has_video_input,omitempty"`
	InputPrice    float64 `json:"input_price"`
}

type TaskPricingProfile struct {
	Resolution    string
	HasVideoInput bool
}

// BillingSetting is managed by config.GlobalConfig.Register.
// DB keys:
// - billing_setting.billing_mode
// - billing_setting.billing_expr
// - billing_setting.output_tier_pricing
type BillingSetting struct {
	BillingMode       map[string]string               `json:"billing_mode"`
	BillingExpr       map[string]string               `json:"billing_expr"`
	OutputTierPricing map[string][]OutputTierPricing `json:"output_tier_pricing"`
}

var billingSetting = BillingSetting{
	BillingMode:       make(map[string]string),
	BillingExpr:       make(map[string]string),
	OutputTierPricing: make(map[string][]OutputTierPricing),
}

func init() {
	config.GlobalConfig.Register("billing_setting", &billingSetting)
}

// ---------------------------------------------------------------------------
// Read accessors (hot path, must be fast)
// ---------------------------------------------------------------------------

func GetBillingMode(model string) string {
	if mode, ok := billingSetting.BillingMode[model]; ok {
		return mode
	}
	return BillingModeRatio
}

func GetBillingExpr(model string) (string, bool) {
	expr, ok := billingSetting.BillingExpr[model]
	return expr, ok
}

func GetOutputTierPricing(model string) ([]OutputTierPricing, bool) {
	tiers, ok := billingSetting.OutputTierPricing[model]
	if !ok {
		return nil, false
	}
	normalized := NormalizeOutputTierPricing(tiers)
	if len(normalized) == 0 {
		return nil, false
	}
	return normalized, true
}

func GetBillingModeCopy() map[string]string {
	return lo.Assign(billingSetting.BillingMode)
}

func GetBillingExprCopy() map[string]string {
	return lo.Assign(billingSetting.BillingExpr)
}

func GetOutputTierPricingCopy() map[string][]OutputTierPricing {
	result := make(map[string][]OutputTierPricing, len(billingSetting.OutputTierPricing))
	for model, tiers := range billingSetting.OutputTierPricing {
		normalized := NormalizeOutputTierPricing(tiers)
		if len(normalized) == 0 {
			continue
		}
		result[model] = normalized
	}
	return result
}

func GetPricingSyncData(base map[string]any) map[string]any {
	extra := make(map[string]any, 3)
	if modes := GetBillingModeCopy(); len(modes) > 0 {
		extra[BillingModeField] = modes
	}
	if exprs := GetBillingExprCopy(); len(exprs) > 0 {
		extra[BillingExprField] = exprs
	}
	if tiers := GetOutputTierPricingCopy(); len(tiers) > 0 {
		extra[OutputTierPricingField] = tiers
	}
	return lo.Assign(base, extra)
}

func NormalizeOutputTierPricing(tiers []OutputTierPricing) []OutputTierPricing {
	if len(tiers) == 0 {
		return nil
	}
	normalized := make([]OutputTierPricing, 0, len(tiers))
	for _, tier := range tiers {
		if tier.InputPrice <= 0 {
			continue
		}
		item := OutputTierPricing{
			Label:      tier.Label,
			Resolution: NormalizeResolution(tier.Resolution),
			InputPrice: tier.InputPrice,
		}
		if tier.HasVideoInput != nil {
			value := *tier.HasVideoInput
			item.HasVideoInput = &value
		}
		normalized = append(normalized, item)
	}
	return normalized
}

func NormalizeResolution(value string) string {
	resolution := strings.ToLower(strings.TrimSpace(value))
	switch resolution {
	case "720", "720p":
		return "720p"
	case "1080", "1080p":
		return "1080p"
	case "2160", "2160p", "4k":
		return "4k"
	}
	if strings.Contains(resolution, "x") {
		parts := strings.SplitN(resolution, "x", 2)
		for _, part := range parts {
			switch strings.TrimSpace(part) {
			case "720":
				return "720p"
			case "1080":
				return "1080p"
			case "2160":
				return "4k"
			}
		}
	}
	return resolution
}

func MatchOutputTierPricing(profile TaskPricingProfile, tiers []OutputTierPricing) (OutputTierPricing, bool) {
	normalized := NormalizeOutputTierPricing(tiers)
	bestIndex := -1
	bestScore := -1
	for i, tier := range normalized {
		score, ok := matchOutputTierPricing(profile, tier)
		if !ok {
			continue
		}
		if score > bestScore {
			bestIndex = i
			bestScore = score
		}
	}
	if bestIndex < 0 {
		return OutputTierPricing{}, false
	}
	return normalized[bestIndex], true
}

func matchOutputTierPricing(profile TaskPricingProfile, tier OutputTierPricing) (int, bool) {
	score := 0
	if tier.Resolution != "" {
		if NormalizeResolution(profile.Resolution) != tier.Resolution {
			return 0, false
		}
		score++
	}
	if tier.HasVideoInput != nil {
		if profile.HasVideoInput != *tier.HasVideoInput {
			return 0, false
		}
		score++
	}
	return score, true
}

// ---------------------------------------------------------------------------
// Smoke test (called externally for validation before save)
// ---------------------------------------------------------------------------

func SmokeTestExpr(exprStr string) error {
	return smokeTestExpr(exprStr)
}

func smokeTestExpr(exprStr string) error {
	vectors := []billingexpr.TokenParams{
		{P: 0, C: 0, Len: 0},
		{P: 1000, C: 1000, Len: 1000},
		{P: 100000, C: 100000, Len: 100000},
		{P: 1000000, C: 1000000, Len: 1000000},
	}
	requests := []billingexpr.RequestInput{
		{},
		{
			Headers: map[string]string{
				"anthropic-beta": "fast-mode-2026-02-01",
			},
			Body: []byte(`{"service_tier":"fast","stream_options":{"include_usage":true},"messages":[1,2,3,4,5,6,7,8,9,10,11,12,13,14,15,16,17,18,19,20,21]}`),
		},
	}

	for _, v := range vectors {
		for _, request := range requests {
			result, _, err := billingexpr.RunExprWithRequest(exprStr, v, request)
			if err != nil {
				return fmt.Errorf("vector {p=%g, c=%g}: run failed: %w", v.P, v.C, err)
			}
			if result < 0 {
				return fmt.Errorf("vector {p=%g, c=%g}: result %f < 0", v.P, v.C, result)
			}
		}
	}
	return nil
}
