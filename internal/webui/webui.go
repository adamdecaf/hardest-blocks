//go:build wasm

package main

import (
	"fmt"
	"strconv"
	"syscall/js"

	"github.com/adamdecaf/bitaxe-stats/pkg/blockchain"
)

func stringToDifficulty(s string) (blockchain.Difficulty, error) {
	value, err := strconv.ParseFloat(s, 64)
	if err != nil {
		return blockchain.Difficulty{}, fmt.Errorf("failed to convert string to float: %v", err)
	}
	return blockchain.Difficulty{RawValue: value}, nil
}

func formatDifficulty(this js.Value, args []js.Value) interface{} {
	if len(args) < 1 {
		return "Error: please provide a string value"
	}

	input := args[0].String()

	diff, err := stringToDifficulty(input)
	if err != nil {
		return fmt.Sprintf("Error: %v", err)
	}

	return diff.Format()
}

func main() {
	c := make(chan struct{}, 0)

	js.Global().Set("formatDifficulty", js.FuncOf(formatDifficulty))

	<-c
}
