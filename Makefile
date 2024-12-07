#GOLANGCI_LINT_VERSION := "v1.43.0" # Optional configuration to pinpoint golangci-lint version.
#MOCKERY_VERSION="2.36.0" # Optional configuration to pinpoint mockery version.
#PROTOBUF_VERSION="28.3" # Optional configuration to pinpoint protobuf version.
#PROTOC_GEN_GO_VERSION="v1.35.2" # Optional configuration to pinpoint protoc-gen-go version.
#PROTOC_GEN_GO_GRPC_VERSION="1.5.1" # Optional configuration to pinpoint protoc-gen-go-grpc version.
#PROTOC_GEN_GRPC_GATEWAY_VERSION="v2.24.0" # Optional configuration to pinpoint protoc-gen-grpc-gateway version.

PWD := $(shell pwd)

MODULES := \
    DEVGO_PATH=github.com/bool64/dev \
    DEVSERVICEGO_PATH=github.com/dohernandez/go-grpc-service \
    DEVGRPCGO_PATH=github.com/dohernandez/dev-grpc \

-include $(PWD)/vendor/github.com/dohernandez/go-grpc-service/dev/makefiles/main.mk

# Add your include here with based path to the module.

-include $(DEVGO_PATH)/makefiles/lint.mk
-include $(DEVGO_PATH)/makefiles/test-unit.mk
-include $(DEVGO_PATH)/makefiles/bench.mk

BUILD_LDFLAGS="-s -w"
BUILD_PKG = ./cmd/...
BINARY_NAME = kit-template

DOCKER_COMPOSE_PROFILE = "all"

-include $(DEVSERVICEGO_PATH)/makefiles/dep.mk
-include $(DEVSERVICEGO_PATH)/makefiles/docker.mk
-include $(DEVSERVICEGO_PATH)makefiles/test-integration.mk
-include $(DEVSERVICEGO_PATH)/makefiles/database.mk
-include $(DEVSERVICEGO_PATH)/makefiles/mockery.mk

SRC_PROTO_PATH = $(PWD)/resources/proto
GO_PROTO_PATH = $(PWD)/internal/platform/service/pb
SWAGGER_PATH = $(PWD)/resources/swagger

-include $(DEVGRPCGO_PATH)/makefiles/protoc.mk

# Add your custom targets here.


## Run tests
test: test-unit test-integration

## Check the commit compile and test the change
check: lint test

## Install all require tools to work with the project
tools: protoc-cli mockery-cli

## Runs commands described by directives within existing files, usage: "make generate SOURCE=<file.go... | packages>"
generate: protoc-cli mockery-cli
	@echo "Running go generate $(or $(SOURCE),./...)"
	@go generate $(or $(SOURCE),./...)

## Generate code from proto file(s) and swagger file
proto-gen: proto-gen-code-swagger
	@cat $(SWAGGER_PATH)/service.swagger.json | jq del\(.paths[][].responses.'"default"'\) > $(SWAGGER_PATH)/service.swagger.json.tmp
	@mv $(SWAGGER_PATH)/service.swagger.json.tmp $(SWAGGER_PATH)/service.swagger.json

## Run docker-compose down from file DOCKER_COMPOSE_PATH with project name DOCKER_COMPOSE_PROJECT_NAME
dc-up-dev:
	@echo "Starting docker compose for development."
	@DOCKER_COMPOSE_PROFILE=dev docker-compose -f docker-compose.yml -f docker-compose.development.yml up -d

## Run docker-compose down with project name DOCKER_COMPOSE_PROJECT_NAME
dc-down-dev:
	@echo "Stopping docker compose for development."
	@DOCKER_COMPOSE_PROFILE=dev docker-compose -f docker-compose.yml -f docker-compose.development.yml down -v