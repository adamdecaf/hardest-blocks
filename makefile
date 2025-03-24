.PHONY: site
site:
	python3 -m http.server -d ./docs/

.PHONY: webui
webui:
	cp "$(shell go env GOROOT)/lib/wasm/wasm_exec.js" ./docs/wasm_exec.js
	GOOS=js GOARCH=wasm go build -o docs/hardest-blocks.wasm ./internal/webui/

.PHONY: generate
generate:
	go run ./cmd/hardest-blocks
