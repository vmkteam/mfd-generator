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
