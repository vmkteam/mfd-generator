PKG := `go list -f {{.Dir}} ./...`

LINT_VERSION := v2.1.6


tools:
	@curl -sfL https://raw.githubusercontent.com/golangci/golangci-lint/master/install.sh | sh -s -- -b $(go env GOPATH)/bin ${LINT_VERSION}

fmt:
	@golangci-lint fmt

lint:
	@golangci-lint version
	@golangci-lint config verify
	@golangci-lint run

test:
	@go test -v ./...

generate:
	@go generate ./api

mod:
	@go mod tidy
