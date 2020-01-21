GOLANGCILINT := $(shell command -v golangci-lint 2> /dev/null)

.PHONY: vendor
vendor:
	@GO111MODULE=on go mod tidy
	@GO111MODULE=on go mod vendor

build: build_omctl build_swagger

build_omctl:
	@GO111MODULE=on go build -mod=vendor -o $(PWD)/omctl github.com/checkr/openmock/cmd/omctl

build_swagger:
	@echo "Building OpenMock Server to $(PWD)/om-swagger ..."
	GO111MODULE=on go build -mod=vendor -o $(PWD)/om github.com/checkr/openmock/swagger_gen/cmd/open-mock-server

test: lint
	@GO111MODULE=on go test -mod=vendor -race -covermode=atomic  . ./pkg/admin 

run: build
	OPENMOCK_TEMPLATES_DIR=./demo_templates ./om

lint:
ifndef GOLANGCILINT
	@GO111MODULE=off go get -u github.com/myitcv/gobin
	@gobin github.com/golangci/golangci-lint/cmd/golangci-lint@v1.17.1
endif
	@golangci-lint run

#################
# Swagger stuff #
#################

PWD := $(shell pwd)
GOPATH := $(shell go env GOPATH)

gen: api_docs swagger

api_docs:
	@echo "Installing swagger-merger" && npm install swagger-merger -g
	@swagger-merger -i $(PWD)/swagger/index.yaml -o $(PWD)/docs/api_docs/bundle.yaml

verify_swagger:
	@echo "Running $@"
	@swagger validate $(PWD)/docs/api_docs/bundle.yaml

# list of files that contain custom edits and shouldn't be overwritten by generation
PROTECTED_FILES := restapi/configure_open_mock.go restapi/server.go cmd/open-mock-server/main.go

swagger: verify_swagger
	@echo "Regenerate swagger files"
	@for file in $(PROTECTED_FILES); do \
		echo $$file ; \
		rm -f /tmp/`basename $$file`; \
		cp $(PWD)/swagger_gen/$$file /tmp/`basename $$file` 2>/dev/null ; \
	done
	@rm -f /tmp/configure_open_mock.go
	@cp $(PWD)/swagger_gen/restapi/configure_open_mock.go /tmp/configure_open_mock.go 2>/dev/null || :
	@rm -rf $(PWD)/swagger_gen
	@mkdir $(PWD)/swagger_gen
	@swagger generate server -t ./swagger_gen -f $(PWD)/docs/api_docs/bundle.yaml
	@for file in $(PROTECTED_FILES); do \
		cp /tmp/`basename $$file` $(PWD)/swagger_gen/$$file 2>/dev/null ; \
	done
