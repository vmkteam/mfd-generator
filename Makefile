PKG := `go list -f {{.Dir}} ./...`

fmt:
	@goimports -local "github.com/vmkteam/mfd-generator" -l -w $(PKG)

lint:
	@golangci-lint run -c .golangci.yml

test:
	@go test -v ./...

generate:
	@go generate ./api

mod:
	@go mod tidy

db-test:
	@echo "Rebuilding test DB..."
	@dropdb --if-exists newsportal
	@createdb -E UTF-8 -O postgres -T template0 --lc-collate C --lc-ctype=ru_RU.UTF-8 newsportal
	@psql newsportal < docs/schema.sql
