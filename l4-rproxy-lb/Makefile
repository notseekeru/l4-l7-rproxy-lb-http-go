BINARY_NAME=l4-rproxy-lb-go
MAIN_PACKAGE_PATH=.

.PHONY: tidy
tidy:
	go fmt ./...
	go vmt ./...
	go mod tidy

.PHONY: audit
audit:
	go mod verify
	go vet ./...
	@if command -v golangci-lint > /dev/null; then \
		golangci-lint run ./...; \
	else \
		echo "golangci-lint not installed. Skipping..."; \
	fi

.PHONY: run
run:
	go run $(MAIN_PACKAGE_PATH)

.PHONY: test
test:
	go test -v -race -buildvcs ./...

.PHONY: build
build:
	go build -ldflags="-s -w" -o bin/$(BINARY_NAME) $(MAIN_PACKAGE_PATH)

.PHONY: clean
clean:
	go clean
	rm -rf bin/ coverage.out
