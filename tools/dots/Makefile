# chezmoi-dots: the `chezmoi dots` plugin. Tools are pinned go.mod / golangci-lint.mod
# dependencies, run with `go tool`; the host needs only go.
GOLANGCI := go tool -modfile=golangci-lint.mod golangci-lint
LINT_GOOS ?= linux darwin windows

.PHONY: help fmt lint vuln test check check-fast build install

help: ## List targets
	@grep -E '^[a-z-]+:.*##' $(MAKEFILE_LIST) | awk -F':.*## ' '{printf "%-11s %s\n", $$1, $$2}'

fmt: ## Format, tidy and apply modernizers
	go mod tidy
	go fix ./...
	$(GOLANGCI) fmt

lint: ## Tidy check, then format check and lint per target OS
	go mod tidy -diff
	@bin=$$(go tool -n -modfile=golangci-lint.mod golangci-lint) && \
	for os in $(LINT_GOOS); do echo "golangci-lint run (GOOS=$$os)"; GOOS=$$os $$bin run || exit 1; done

vuln: ## Reachable known vulnerabilities (needs the network)
	go tool govulncheck ./...

test: ## Unit tests with the race detector, in random order
	go test -race -shuffle=on -coverprofile=coverage.out ./...

check: lint test ## lint + test

check-fast: ## Tests without the race detector
	go test -shuffle=on ./...

build: ## Static binary into bin/
	CGO_ENABLED=0 go build -trimpath -o bin/ ./cmd/...

install: build ## Put the plugin on PATH until a release is pinned (~/.local/bin)
	install -d $(HOME)/.local/bin
	install bin/chezmoi-dots $(HOME)/.local/bin/chezmoi-dots
