
GOLANGCI_LINT_VERSION=v1.10.2

LINTER=./bin/golangci-lint

.PHONY: build clean test lint

build:
	go get ./...
	go build ./...

clean:
	go clean

test:
	@# Note, we need to specify all these packages individually for go test in order to remain 1.8-compatible
	go test -race -v ./ldvalue ./lduser

$(LINTER):
	curl -sfL https://install.goreleaser.com/github.com/golangci/golangci-lint.sh | bash -s $(GOLANGCI_LINT_VERSION)

lint: $(LINTER)
	$(LINTER) run ./...
