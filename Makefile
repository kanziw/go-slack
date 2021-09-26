GOPATH:=$(shell go env GOPATH)

.PHONY: init
## init: initialize the application
init:
	go mod download

.PHONY: format
## format: format files
format:
	@go install golang.org/x/tools/cmd/goimports@v0.1.6
	@go install github.com/aristanetworks/goarista/cmd/importsort@latest
	goimports -local github.com/kanziw -w .
	importsort -s github.com/kanziw -w $$(find . -name "*.go")
	gofmt -s -w .
	go mod tidy

.PHONY: test
## test: run tests
test:
	@go install github.com/rakyll/gotest@v0.0.6
	gotest -p 1 -race -cover -v ./...
	$(MAKE) format

.PHONY: coverage
## coverage: run tests with coverage
coverage:
	@go install github.com/rakyll/gotest@v0.0.6
	gotest -p 1 -race -coverprofile=coverage.txt -covermode=atomic -v ./...

.PHONY: lint
## lint: check everything's okay
lint:
	golangci-lint run ./...
	go mod verify

.PHONY: generate
## generate: generate source code for mocking
generate:
	@go install golang.org/x/tools/cmd/stringer@v0.1.6
	@go install github.com/golang/mock/mockgen@v1.6.0
	go generate ./...
	$(MAKE) format

.PHONY: help
## help: prints this help message
help:
	@echo "Usage: \n"
	@sed -n 's/^##//p' ${MAKEFILE_LIST} | column -t -s ':'
