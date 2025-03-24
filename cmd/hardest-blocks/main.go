package main

import (
	"bytes"
	"cmp"
	"context"
	"encoding/json"
	"fmt"
	"log"
	"os"
	"path/filepath"
	"slices"
	"time"

	"github.com/adamdecaf/bitaxe-stats/pkg/blockchain"
	"github.com/adamdecaf/hardest-blocks/internal/blockchaininfo"
)

const (
	blockHistory = 1000 // Keep the largest 1000 blocks

	iterationDelay = 500 * time.Millisecond
)

func main() {
	// Read the latest block hash and the current data
	// Look for blocks on the chain until we find one without NextBlock
	latestHashFilepath := filepath.Join("docs", "LATEST")
	latestHash, err := readFile(latestHashFilepath)
	if err != nil {
		log.Fatalf("ERROR reading last block processed: %v", err)
	}

	dataFilepath := filepath.Join("docs", "data.json")
	data, err := readLatestData(dataFilepath)
	if err != nil {
		log.Fatalf("ERROR reading hardest blocks data: %v", err)
	}

	ctx := context.Background()
	chainclient := blockchaininfo.NewClient(nil)

	// Grab blocks newer than latestHash and record them if
	for {
		block, err := chainclient.BlockInfo(ctx, latestHash)
		if err != nil {
			log.Fatalf("ERROR getting block: %v", err)
		}

		diff, err := blockchain.HashDifficulty(block.Hash)
		if err != nil {
			log.Fatalf("ERROR calculating block difficulty: %v", err)
		}

		log.Printf("INFO: found block %v (%v)", block.Hash, diff.Format())

		// Add the block and sort, trim to our length
		data.LargestDifficulties = append(data.LargestDifficulties, Block{
			Hash:       block.Hash,
			Difficulty: diff,
			Raw:        block,
		})
		slices.SortFunc(data.LargestDifficulties, func(b1, b2 Block) int {
			return -1 * cmp.Compare(b1.Difficulty.RawValue, b2.Difficulty.RawValue)
		})

		// Remove smaller difficulties
		if len(data.LargestDifficulties) > blockHistory {
			data.LargestDifficulties = data.LargestDifficulties[:blockHistory]
		}

		// Write both files
		err = os.WriteFile(latestHashFilepath, []byte(block.Hash), 0644)
		if err != nil {
			log.Fatalf("ERROR writing latest hash: %v", err)
		}

		bs, _ := json.Marshal(data)
		err = os.WriteFile(dataFilepath, bs, 0644)
		if err != nil {
			log.Fatalf("ERROR writing block data: %v", err)
		}

		// Update the cursor for the next block
		latestHash = block.NextBlock

		// Sleep so we don't exceed rate limits
		time.Sleep(iterationDelay)
	}
}

type Data struct {
	LargestDifficulties []Block
}

type Block struct {
	Hash       string
	Difficulty blockchain.Difficulty
	Raw        blockchaininfo.RawBlock
}

func readFile(path string) (string, error) {
	bs, err := os.ReadFile(path)
	if err != nil {
		return "", fmt.Errorf("reading %s failed: %w", path, err)
	}
	return string(bytes.TrimSpace(bs)), nil
}

func readLatestData(path string) (Data, error) {
	bs, err := os.ReadFile(path)
	if err != nil {
		return Data{}, fmt.Errorf("reading %s failed: %w", path, err)
	}

	var data Data
	err = json.Unmarshal(bs, &data)
	if err != nil {
		return Data{}, fmt.Errorf("json unmarshal of data: %w", err)
	}
	return data, nil
}
