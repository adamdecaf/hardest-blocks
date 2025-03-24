.PHONY: site
site:
	python3 -m http.server -d ./docs/

.PHONY: generate
generate:
	go run ./cmd/hardest-blocks
