.PHONY: update-go build test fmt clean

BINARY_NAME := eastloader

update-go:
	go get -u .
	go mod tidy

build: ## Build the binary
	go build -o $(BINARY_NAME) .

test: ## Run tests with coverage and race detector
	go test -race -coverprofile=coverage.out -v $$(go list ./... | grep -v "/cmd$$")
	go tool cover -func=coverage.out
	go tool cover -html=coverage.out -o coverage.html

fmt: ## Format all go files
	go fmt ./...

clean: ## Remove built files
	rm -f $(BINARY_NAME) coverage.out coverage.html
