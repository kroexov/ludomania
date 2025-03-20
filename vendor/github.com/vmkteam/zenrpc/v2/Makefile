PKG := `go list -f {{.Dir}} ./...`

fmt:
	@goimports -local "github.com/vmkteam/zenrpc" -l -w $(PKG)

lint:
	@golangci-lint run -c .golangci.yml

test:
	@go test -v ./...

mod:
	@go mod tidy
