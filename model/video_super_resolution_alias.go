package model

import (
	"strings"

	"github.com/QuantumNous/new-api/setting/ratio_setting"
)

const (
	seedance15ModelAlias       = "seedance1.5"
	seedance15SRModelAlias     = "seedance1.5-sr"
	seedance2ModelAlias        = "seedance2"
	seedance20ModelAlias       = "seedance2.0"
	seedance2SRModelAlias      = "seedance2-sr"
	seedance20SRModelAlias     = "seedance2.0-sr"
	seedance20FastModelAlias   = "seedance2.0fast"
	seedance20FastSRModelAlias = "seedance2.0fast-sr"
	sd20FastModelAlias         = "sd2.0fast"
	sd20FastSRModelAlias       = "sd2.0fast-sr"
)

func normalizeVideoSuperResolutionAbilityModel(modelName string) string {
	switch strings.ToLower(strings.TrimSpace(modelName)) {
	case seedance15SRModelAlias:
		return seedance15ModelAlias
	case seedance2SRModelAlias:
		return seedance2ModelAlias
	case seedance20SRModelAlias:
		return seedance20ModelAlias
	case seedance20FastSRModelAlias:
		return seedance20FastModelAlias
	case sd20FastSRModelAlias:
		return sd20FastModelAlias
	default:
		return strings.TrimSpace(modelName)
	}
}

func buildVideoSuperResolutionAliasCandidates(modelName string) []string {
	switch strings.ToLower(strings.TrimSpace(modelName)) {
	case seedance15ModelAlias, seedance15SRModelAlias:
		return []string{
			seedance15ModelAlias,
			seedance15SRModelAlias,
		}
	case seedance2ModelAlias, seedance20ModelAlias, seedance2SRModelAlias, seedance20SRModelAlias:
		return []string{
			seedance2ModelAlias,
			seedance20ModelAlias,
			seedance2SRModelAlias,
			seedance20SRModelAlias,
		}
	case seedance20FastModelAlias, seedance20FastSRModelAlias, sd20FastModelAlias, sd20FastSRModelAlias:
		return []string{
			seedance20FastModelAlias,
			seedance20FastSRModelAlias,
			sd20FastModelAlias,
			sd20FastSRModelAlias,
		}
	default:
		return []string{strings.TrimSpace(modelName)}
	}
}

func buildAbilityLookupCandidates(modelName string) []string {
	candidates := make([]string, 0, 4)
	addCandidate := func(candidate string) {
		candidate = strings.TrimSpace(candidate)
		if candidate == "" {
			return
		}
		for _, existing := range candidates {
			if existing == candidate {
				return
			}
		}
		candidates = append(candidates, candidate)
	}

	trimmedModelName := strings.TrimSpace(modelName)
	addCandidate(trimmedModelName)
	addCandidate(ratio_setting.FormatMatchingModelName(trimmedModelName))

	for _, candidate := range buildVideoSuperResolutionAliasCandidates(modelName) {
		addCandidate(candidate)
		addCandidate(ratio_setting.FormatMatchingModelName(candidate))
	}

	return candidates
}
