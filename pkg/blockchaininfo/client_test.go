package blockchaininfo

import (
	"context"
	"testing"

	"github.com/stretchr/testify/require"
)

func TestClient(t *testing.T) {
	cc := NewClient(nil)

	block, err := cc.BlockInfo(context.Background(), "000000000019d6689c085ae165831e934ff763ae46a2a6c172b3f1b60a8ce26f")
	require.NoError(t, err)

	require.Equal(t, "000000000019d6689c085ae165831e934ff763ae46a2a6c172b3f1b60a8ce26f", block.Hash)
	require.Equal(t, "0000000000000000000000000000000000000000000000000000000000000000", block.PreviousBlock)
	require.Equal(t, "00000000839a8e6886ab5951d76f411475428afc90947ee320161bbf18eb6048", block.NextBlock)
}
