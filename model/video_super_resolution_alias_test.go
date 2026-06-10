package model

import (
	"testing"

	"github.com/QuantumNous/new-api/common"
)

func TestBuildAbilityLookupCandidatesForDedicatedSRModel(t *testing.T) {
	candidates := buildAbilityLookupCandidates("seedance2-sr")
	expected := []string{"seedance2-sr", "seedance2", "seedance2.0", "seedance2.0-sr"}

	if len(candidates) != len(expected) {
		t.Fatalf("expected %d candidates, got %d: %#v", len(expected), len(candidates), candidates)
	}
	for i := range expected {
		if candidates[i] != expected[i] {
			t.Fatalf("expected candidate %d to be %q, got %q", i, expected[i], candidates[i])
		}
	}
}

func TestBuildAbilityLookupCandidatesForSeedance15SRModel(t *testing.T) {
	candidates := buildAbilityLookupCandidates("seedance1.5-sr")
	expected := []string{"seedance1.5-sr", "seedance1.5"}

	if len(candidates) != len(expected) {
		t.Fatalf("expected %d candidates, got %d: %#v", len(expected), len(candidates), candidates)
	}
	for i := range expected {
		if candidates[i] != expected[i] {
			t.Fatalf("expected candidate %d to be %q, got %q", i, expected[i], candidates[i])
		}
	}
}

func TestBuildAbilityLookupCandidatesForBaseModel(t *testing.T) {
	candidates := buildAbilityLookupCandidates("sd2.0fast")
	expected := []string{"sd2.0fast", "seedance2.0fast", "seedance2.0fast-sr", "sd2.0fast-sr"}

	if len(candidates) != len(expected) {
		t.Fatalf("expected %d candidates, got %d: %#v", len(expected), len(candidates), candidates)
	}
	for i := range expected {
		if candidates[i] != expected[i] {
			t.Fatalf("expected candidate %d to be %q, got %q", i, expected[i], candidates[i])
		}
	}
}

func TestGetRandomSatisfiedChannelFallsBackToBaseSRAbility(t *testing.T) {
	originalMemoryCacheEnabled := common.MemoryCacheEnabled
	originalGroup2Model2Channels := group2model2channels
	originalChannelsIDM := channelsIDM
	defer func() {
		common.MemoryCacheEnabled = originalMemoryCacheEnabled
		group2model2channels = originalGroup2Model2Channels
		channelsIDM = originalChannelsIDM
	}()

	common.MemoryCacheEnabled = true
	group2model2channels = map[string]map[string][]int{
		"default": {
			"seedance2": {101},
		},
	}
	priority := int64(1)
	weight := uint(100)
	channelsIDM = map[int]*Channel{
		101: {
			Id:       101,
			Priority: &priority,
			Weight:   &weight,
		},
	}

	channel, err := GetRandomSatisfiedChannel("default", "seedance2-sr", 0)
	if err != nil {
		t.Fatalf("expected no error, got %v", err)
	}
	if channel == nil {
		t.Fatal("expected fallback channel, got nil")
	}
	if channel.Id != 101 {
		t.Fatalf("expected channel id 101, got %d", channel.Id)
	}
}

func TestGetRandomSatisfiedChannelPrefersExactBaseModelOverSRSibling(t *testing.T) {
	originalMemoryCacheEnabled := common.MemoryCacheEnabled
	originalGroup2Model2Channels := group2model2channels
	originalChannelsIDM := channelsIDM
	defer func() {
		common.MemoryCacheEnabled = originalMemoryCacheEnabled
		group2model2channels = originalGroup2Model2Channels
		channelsIDM = originalChannelsIDM
	}()

	common.MemoryCacheEnabled = true
	group2model2channels = map[string]map[string][]int{
		"default": {
			"sd2.0fast":          {7},
			"seedance2.0fast-sr": {8},
		},
	}
	priority := int64(1)
	weight := uint(100)
	channelsIDM = map[int]*Channel{
		7: {Id: 7, Priority: &priority, Weight: &weight},
		8: {Id: 8, Priority: &priority, Weight: &weight},
	}

	channel, err := GetRandomSatisfiedChannel("default", "sd2.0fast", 0)
	if err != nil {
		t.Fatalf("expected no error, got %v", err)
	}
	if channel == nil {
		t.Fatal("expected channel, got nil")
	}
	if channel.Id != 7 {
		t.Fatalf("expected base channel id 7, got %d", channel.Id)
	}
}

func TestGetRandomSatisfiedChannelPrefersExactSRModelOverBaseSibling(t *testing.T) {
	originalMemoryCacheEnabled := common.MemoryCacheEnabled
	originalGroup2Model2Channels := group2model2channels
	originalChannelsIDM := channelsIDM
	defer func() {
		common.MemoryCacheEnabled = originalMemoryCacheEnabled
		group2model2channels = originalGroup2Model2Channels
		channelsIDM = originalChannelsIDM
	}()

	common.MemoryCacheEnabled = true
	group2model2channels = map[string]map[string][]int{
		"default": {
			"seedance2.0fast":    {7},
			"seedance2.0fast-sr": {8},
		},
	}
	priority := int64(1)
	weight := uint(100)
	channelsIDM = map[int]*Channel{
		7: {Id: 7, Priority: &priority, Weight: &weight},
		8: {Id: 8, Priority: &priority, Weight: &weight},
	}

	channel, err := GetRandomSatisfiedChannel("default", "seedance2.0fast-sr", 0)
	if err != nil {
		t.Fatalf("expected no error, got %v", err)
	}
	if channel == nil {
		t.Fatal("expected channel, got nil")
	}
	if channel.Id != 8 {
		t.Fatalf("expected sr channel id 8, got %d", channel.Id)
	}
}

func TestBuildAbilityLookupCandidatesForProductionFastSRModel(t *testing.T) {
	candidates := buildAbilityLookupCandidates("seedance2.0fast-sr")
	expected := []string{"seedance2.0fast-sr", "seedance2.0fast", "sd2.0fast", "sd2.0fast-sr"}

	if len(candidates) != len(expected) {
		t.Fatalf("expected %d candidates, got %d: %#v", len(expected), len(candidates), candidates)
	}
	for i := range expected {
		if candidates[i] != expected[i] {
			t.Fatalf("expected candidate %d to be %q, got %q", i, expected[i], candidates[i])
		}
	}
}
