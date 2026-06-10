package controller

import (
	"testing"
	"time"

	"github.com/stretchr/testify/require"
)

func TestNormalizePortraitPreviewCandidateFallsBackWhenSignedURLExpired(t *testing.T) {
	expired := time.Now().UTC().Add(-2 * time.Minute).Format("20060102T150405Z")
	fallback := "https://example.com/uploads/source.jpg"
	candidate := "https://tos.example.com/object?X-Tos-Date=" + expired + "&X-Tos-Expires=60"

	result := normalizePortraitPreviewCandidate(candidate, fallback)

	require.Equal(t, fallback, result)
}

func TestNormalizePortraitPreviewCandidateKeepsValidSignedURL(t *testing.T) {
	future := time.Now().UTC().Add(-30 * time.Second).Format("20060102T150405Z")
	fallback := "https://example.com/uploads/source.jpg"
	candidate := "https://tos.example.com/object?X-Tos-Date=" + future + "&X-Tos-Expires=300"

	result := normalizePortraitPreviewCandidate(candidate, fallback)

	require.Equal(t, candidate, result)
}
