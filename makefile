.PHONY: site
site:
	python3 -m http.server -d ./site/

.PHONY: generate
generate:
	go run ./cmd/hardest-blocks
