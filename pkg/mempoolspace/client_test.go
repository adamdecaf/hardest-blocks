package mempoolspace_test

import (
	"context"
	"testing"

	"github.com/adamdecaf/hardest-blocks/pkg/mempoolspace"

	"github.com/stretchr/testify/require"
)

func TestClient_GetDifficultyAdjustments(t *testing.T) {
	adjustments, err := mempoolspace.GetDifficultyAdjustments(context.Background())
	require.NoError(t, err)
	require.Greater(t, len(adjustments), 10)
}
