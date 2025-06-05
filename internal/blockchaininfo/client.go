package blockchaininfo

import (
	"bytes"
	"cmp"
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"net/http"
	"strings"
	"time"

	"github.com/hashicorp/go-retryablehttp"
)

type Client interface {
	BlockInfo(ctx context.Context, hash string) (RawBlock, error)
}

func NewClient(httpClient *http.Client) Client {
	httpClient = cmp.Or(httpClient, &http.Client{
		Timeout: 20 * time.Second,
	})

	cc := retryablehttp.NewClient()
	cc.HTTPClient = httpClient
	cc.Logger = nil

	return &client{
		httpClient: cc,
	}
}

type client struct {
	httpClient *retryablehttp.Client
}

type RawBlock struct {
	Hash          string
	PreviousBlock string
	NextBlock     string

	MainChain bool

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

		MainChain bool `json:"main_chain"`

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
	case nil:
		// do nothing
	default:
		return fmt.Errorf("unexpected time type: %T - %v", t, t)
	}
	if err != nil {
		return fmt.Errorf("parsing time: %w", err)
	}

	return nil
}

const (
	maxRetries = 3
	baseDelay  = 500 * time.Millisecond
)

func (c *client) BlockInfo(ctx context.Context, hash string) (RawBlock, error) {
	address := fmt.Sprintf("https://blockchain.info/rawblock/%s", hash)

	for attempt := 0; attempt < maxRetries; attempt++ {
		resp, err := c.httpClient.Get(address)
		if err != nil {
			if strings.Contains(err.Error(), "context deadline exceeded") || strings.Contains(err.Error(), "Client.Timeout exceeded") {
				if attempt < maxRetries-1 {
					backoffDuration := baseDelay * time.Duration(1<<attempt)
					time.Sleep(backoffDuration)
					continue
				}
			}
			return RawBlock{}, fmt.Errorf("getting raw block: %w", err)
		}

		defer resp.Body.Close()

		var block RawBlock
		err = json.NewDecoder(resp.Body).Decode(&block)
		if err != nil {
			switch {
			case strings.Contains(err.Error(), "stream error"),
				strings.Contains(err.Error(), "invalid character"),
				strings.Contains(err.Error(), "unexpected EOF"):
				if attempt < maxRetries-1 {
					backoffDuration := baseDelay * time.Duration(1<<attempt)
					time.Sleep(backoffDuration)
					continue
				}
			}
			return block, fmt.Errorf("decoding response: %w", err)
		}

		if !block.MainChain {
			return block, errors.New("block not in main chain")
		}

		return block, nil
	}

	return RawBlock{}, errors.New("max retries exceeded for BlockInfo")
}
