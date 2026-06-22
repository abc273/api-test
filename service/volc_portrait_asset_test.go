package service

import (
	"errors"
	"testing"

	"github.com/stretchr/testify/require"
)

func TestIsVolcPortraitGroupNotFoundError(t *testing.T) {
	require.True(t, IsVolcPortraitGroupNotFoundError(errors.New(`volc portrait CreateAsset failed: HTTP 404: {"ResponseMetadata":{"Error":{"Code":"NotFound.group_id"}}}`)))
	require.False(t, IsVolcPortraitGroupNotFoundError(errors.New("other error")))
	require.False(t, IsVolcPortraitGroupNotFoundError(nil))
}
