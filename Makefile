GOPATH:=$(shell go env GOPATH)

.PHONY: init
## init: initialize the application
init:
	go mod download

.PHONY: format
## format: format files
format:
	@go get -d github.com/aristanetworks/goarista/cmd/importsort
	importsort -s github.com/kanziw -w $$(find . -name "*.go")
	gofmt -s -w .
	go mod tidy

.PHONY: test
## test: run tests
test:
	@go get github.com/rakyll/gotest
	gotest -p 1 -race -cover -v ./...
	$(MAKE) format

.PHONY: coverage
## coverage: run tests with coverage
coverage:
	@go get github.com/rakyll/gotest
	gotest -p 1 -race -coverprofile=coverage.txt -covermode=atomic -v ./...

.PHONY: lint
## lint: check everything's okay
lint:
	golangci-lint run ./...
	go mod verify

.PHONY: generate
## generate: generate source code for mocking
generate:
	@go get -d golang.org/x/tools/cmd/stringer
	@go get -d github.com/golang/mock/gomock
	@go install github.com/golang/mock/mockgen
	go generate ./...
	$(MAKE) format

.PHONY: help
## help: prints this help message
help:
	@echo "Usage: \n"
	@sed -n 's/^##//p' ${MAKEFILE_LIST} | column -t -s ':'
