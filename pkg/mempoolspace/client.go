package mempoolspace

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"os"
	"time"
)

var (
	httpClient = &http.Client{
		Timeout: 10 * time.Second,
	}
)

type DifficultyAdjustment struct {
	BlockTimestamp   time.Time
	Height           int32
	Difficulty       int64
	DifficultyChange float64
}

func GetDifficultyAdjustments(ctx context.Context) ([]DifficultyAdjustment, error) {
	req, err := http.NewRequest("GET", "https://mempool.space/api/v1/mining/difficulty-adjustments", nil)
	if err != nil {
		return nil, fmt.Errorf("building difficulty adjustments request: %w", err)
	}

	resp, err := httpClient.Do(req)
	if err != nil {
		return nil, fmt.Errorf("problem getting difficulty adjustments: %w", err)
	}
	if resp != nil && resp.Body != nil {
		defer resp.Body.Close()
	}

	var wrapper [][]float64
	err = json.NewDecoder(resp.Body).Decode(&wrapper)
	if err != nil {
		return nil, fmt.Errorf("reading difficulty adjustments json: %w", err)
	}

	var out []DifficultyAdjustment
	for _, adj := range wrapper {
		out = append(out, DifficultyAdjustment{
			BlockTimestamp:   time.Unix(int64(adj[0]), 0),
			Height:           int32(adj[1]),
			Difficulty:       int64(adj[2]),
			DifficultyChange: adj[3],
		})
	}
	return out, nil
}

const difficultyAdjustmentsPath = "docs/difficulty-adjustments.json"

func SaveDifficultyAdjustments(ctx context.Context, adjustments []DifficultyAdjustment) error {
	// Convert to the raw format for saving
	var raw [][]float64
	for _, adj := range adjustments {
		raw = append(raw, []float64{
			float64(adj.BlockTimestamp.Unix()),
			float64(adj.Height),
			float64(adj.Difficulty),
			adj.DifficultyChange,
		})
	}

	data, err := json.Marshal(raw)
	if err != nil {
		return fmt.Errorf("marshaling difficulty adjustments: %w", err)
	}

	err = os.WriteFile(difficultyAdjustmentsPath, data, 0644)
	if err != nil {
		return fmt.Errorf("writing difficulty adjustments file: %w", err)
	}

	return nil
}

func LoadDifficultyAdjustments() ([]DifficultyAdjustment, error) {
	data, err := os.ReadFile(difficultyAdjustmentsPath)
	if err != nil {
		return nil, fmt.Errorf("reading difficulty adjustments file: %w", err)
	}

	var raw [][]float64
	err = json.Unmarshal(data, &raw)
	if err != nil {
		return nil, fmt.Errorf("unmarshaling difficulty adjustments: %w", err)
	}

	var out []DifficultyAdjustment
	for _, adj := range raw {
		out = append(out, DifficultyAdjustment{
			BlockTimestamp:   time.Unix(int64(adj[0]), 0),
			Height:           int32(adj[1]),
			Difficulty:       int64(adj[2]),
			DifficultyChange: adj[3],
		})
	}
	return out, nil
}

func EnsureDifficultyAdjustmentsCover(ctx context.Context, latestHeight int64) ([]DifficultyAdjustment, error) {
	adjustments, err := LoadDifficultyAdjustments()
	if err != nil {
		// File doesn't exist or invalid, fetch fresh
		adjustments, err = GetDifficultyAdjustments(ctx)
		if err != nil {
			return nil, err
		}
		err = SaveDifficultyAdjustments(ctx, adjustments)
		if err != nil {
			return nil, err
		}
		return adjustments, nil
	}

	// Check if the latest adjustment covers the latestHeight
	if len(adjustments) == 0 {
		// Refetch
		adjustments, err = GetDifficultyAdjustments(ctx)
		if err != nil {
			return nil, err
		}
		err = SaveDifficultyAdjustments(ctx, adjustments)
		if err != nil {
			return nil, err
		}
		return adjustments, nil
	}

	// Sort by height descending to find the latest
	maxHeight := int64(0)
	for _, adj := range adjustments {
		if int64(adj.Height) > maxHeight {
			maxHeight = int64(adj.Height)
		}
	}

	if maxHeight < latestHeight {
		// Need to refetch
		adjustments, err = GetDifficultyAdjustments(ctx)
		if err != nil {
			return nil, err
		}
		err = SaveDifficultyAdjustments(ctx, adjustments)
		if err != nil {
			return nil, err
		}
	}

	return adjustments, nil
}
