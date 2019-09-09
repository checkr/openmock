GOLANGCILINT := $(shell command -v golangci-lint 2> /dev/null)

.PHONY: vendor
vendor:
	@GO111MODULE=on go mod tidy
	@GO111MODULE=on go mod vendor

build:
	@GO111MODULE=on go build -mod=vendor -o $(PWD)/om github.com/checkr/openmock/cmd/om

test: lint
	@GO111MODULE=on go test -mod=vendor -race -covermode=atomic .

run: build
	OPENMOCK_TEMPLATES_DIR=./demo_templates ./om

lint:
ifndef GOLANGCILINT
	@GO111MODULE=off go get -u github.com/myitcv/gobin
	@gobin github.com/golangci/golangci-lint/cmd/golangci-lint@v1.17.1
endif
	@golangci-lint run
