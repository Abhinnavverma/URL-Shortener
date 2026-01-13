.PHONY: run-dev test build prettier

run-dev:
	go run cmd/api/main.go

test:
	go test ./...

build:
	go build -o auth-go cmd/main.go

prettier:
	gofmt -w .