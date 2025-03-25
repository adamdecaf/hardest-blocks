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

	Time   time.Time
	Height int64
}

var (
	possibleTimeFormats = []string{
		time.RFC3339,
		"2006-01-02T15:04:05-07:00",
	}
)

func (r *RawBlock) UnmarshalJSON(data []byte) error {
	var aux struct {
		Hash      string   `json:"hash"`
		PrevBlock string   `json:"prev_block"`
		NextBlock []string `json:"next_block"`

		Time   any   `json:"time"`
		Height int64 `json:"height"`
	}

	err := json.NewDecoder(bytes.NewReader(data)).Decode(&aux)
	if err != nil {
		return fmt.Errorf("reading block: %w", err)
	}

	r.Hash = aux.Hash
	r.PreviousBlock = aux.PrevBlock
	r.Height = aux.Height

	if len(aux.NextBlock) > 0 {
		r.NextBlock = aux.NextBlock[0]
	}

	switch t := aux.Time.(type) {
	case string:
		for _, format := range possibleTimeFormats {
			r.Time, err = time.Parse(format, t)
			if err == nil {
				break
			}
		}

	case float64:
		r.Time = time.Unix(int64(t), 0)
	default:
		return fmt.Errorf("unexpected time type: %T - %v", t, t)
	}
	if err != nil {
		return fmt.Errorf("parsing time: %w", err)
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
