package blockchaininfo

import (
	"bytes"
	"cmp"
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"time"
)

type Client interface {
	BlockInfo(ctx context.Context, hash string) (RawBlock, error)
}

func NewClient(httpClient *http.Client) Client {
	httpClient = cmp.Or(httpClient, &http.Client{
		Timeout: 20 * time.Second,
	})

	return &client{
		httpClient: httpClient,
	}
}

type client struct {
	httpClient *http.Client
}

type RawBlock struct {
	Hash          string
	PreviousBlock string
	NextBlock     string
}

func (r *RawBlock) UnmarshalJSON(data []byte) error {
	var aux struct {
		Hash      string   `json:"hash"`
		PrevBlock string   `json:"prev_block"`
		NextBlock []string `json:"next_block"`
	}

	err := json.NewDecoder(bytes.NewReader(data)).Decode(&aux)
	if err != nil {
		return fmt.Errorf("reading block: %w", err)
	}

	r.Hash = aux.Hash
	r.PreviousBlock = aux.PrevBlock

	if len(aux.NextBlock) > 0 {
		r.NextBlock = aux.NextBlock[0]
	}

	return nil
}

func (c *client) BlockInfo(ctx context.Context, hash string) (RawBlock, error) {
	address := fmt.Sprintf("https://blockchain.info/rawblock/%s", hash)

	resp, err := c.httpClient.Get(address)
	if err != nil {
		return RawBlock{}, fmt.Errorf("getting raw block: %w", err)
	}
	defer resp.Body.Close()

	var block RawBlock
	err = json.NewDecoder(resp.Body).Decode(&block)
	if err != nil {
		return block, err
	}
	return block, nil
}
