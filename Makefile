BINARY  := binconv
MODULE  := github.com/f0xtek/binconv
VERSION ?= $(shell git describe --tags --always --dirty 2>/dev/null || echo dev)
LDFLAGS := -X $(MODULE)/internal/cli.version=$(VERSION)

.DEFAULT_GOAL := build

.PHONY: build
build: ## Build the binconv binary into ./bin
	go build -ldflags "$(LDFLAGS)" -o bin/$(BINARY) .

.PHONY: test
test: ## Run the test suite
	go test ./...

.PHONY: vet
vet: ## Run go vet
	go vet ./...

.PHONY: fmt
fmt: ## Check gofmt formatting (fails if any file is unformatted)
	@unformatted="$$(gofmt -l .)"; \
	if [ -n "$$unformatted" ]; then \
		echo "gofmt needed on:"; echo "$$unformatted"; exit 1; \
	fi

.PHONY: check
check: vet fmt test ## Run vet, fmt and test together

.PHONY: install
install: ## Install binconv to $GOBIN (or $GOPATH/bin)
	go install -ldflags "$(LDFLAGS)" .

.PHONY: run
run: build ## Build and run, e.g. make run ARGS="dec hex 1337"
	./bin/$(BINARY) $(ARGS)

.PHONY: clean
clean: ## Remove build artifacts
	rm -rf bin dist

.PHONY: help
help: ## Show this help
	@grep -E '^[a-zA-Z_-]+:.*?## .*$$' $(MAKEFILE_LIST) | sort | \
		awk 'BEGIN {FS = ":.*?## "}; {printf "%-10s %s\n", $$1, $$2}'
