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

db-test:
	@echo "Rebuilding test DB..."
	@dropdb --if-exists newsportal
	@createdb -E UTF-8 -O postgres -T template0 --lc-collate C --lc-ctype=ru_RU.UTF-8 newsportal
	@psql newsportal < schema.sql

test:
	@go test -v ./...

generate:
	@go generate ./api

mod:
	@go mod tidy
