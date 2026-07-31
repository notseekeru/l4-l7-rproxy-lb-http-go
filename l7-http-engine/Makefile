BINARY_NAME=http-go-engine
MAIN_PACKAGE_PATH=./main.go

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

.PHONY: get
get:
	curl -v http://localhost:8080/

.PHONY: post
post:
	curl -v -X POST -d "hello world" http://localhost:8080/

index:
	curl -v http://localhost:8080/index.html

query:
	curl -v "http://localhost:8080/api?age=18&name=mark"
