package relay

import (
	"testing"

	"github.com/QuantumNous/new-api/model"
	"github.com/stretchr/testify/require"
)

func TestShouldCancelTask(t *testing.T) {
	require.True(t, shouldCancelTask(model.TaskStatusNotStart))
	require.True(t, shouldCancelTask(model.TaskStatusSubmitted))
	require.True(t, shouldCancelTask(model.TaskStatusQueued))
	require.True(t, shouldCancelTask(model.TaskStatusUnknown))

	require.False(t, shouldCancelTask(model.TaskStatusInProgress))
	require.False(t, shouldCancelTask(model.TaskStatusSuccess))
	require.False(t, shouldCancelTask(model.TaskStatusFailure))
	require.False(t, shouldCancelTask(model.TaskStatusCancelled))
}
