GOPATH := $(shell go env GOPATH)
MUTEST := $(GOPATH)/bin/go-mutesting
GOFUMPT := $(GOPATH)/bin/gofumpt

# ── Formatting ───────────────────────────────────────────────────────────────

.PHONY: fmt
fmt: $(GOFUMPT)  ## Format code using gofumpt
	$(GOFUMPT) -w .

$(GOFUMPT):
	go install mvdan.cc/gofumpt@latest

# ── Docs / ADRs ───────────────────────────────────────────────────────────────

.PHONY: adr-index docs-lint
adr-index: ## Regenerate docs/adr/README.md from ADR files
	go run ./cmd/adr-lint -write-index -check-index=false

docs-lint: ## Validate ADR template, unique numbers, and index freshness
	go run ./cmd/adr-lint -check-index
	go test ./internal/adr/... -count=1 -cover

# ── Deploy assets (image signing / Kyverno policy) ────────────────────────────

.PHONY: validate-deploy
validate-deploy: ## Static invariants check for release workflow + Kyverno policy
	go test ./internal/deploylint/... -count=1 -v

# ── Mutation testing ──────────────────────────────────────────────────────────

.PHONY: mutation-state-machine
mutation-state-machine: $(MUTEST)  ## Run mutation tests on the subscription state machine
	$(MUTEST) ./internal/subscriptions/...

$(MUTEST):
	go install github.com/avito-tech/go-mutesting/cmd/go-mutesting@latest
