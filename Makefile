.PHONY: vendor
vendor:
	@GO111MODULE=on go mod tidy
	@GO111MODULE=on go mod vendor

build:
	@GO111MODULE=on go build -mod=vendor -o $(PWD)/om github.com/checkr/openmock/cmd/om

test:
	@GO111MODULE=on go test -mod=vendor -race -covermode=atomic .

run: build
	OPENMOCK_TEMPLATES_DIR=./demo_templates ./om
