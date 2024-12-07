#GOLANGCI_LINT_VERSION := "v1.43.0" # Optional configuration to pinpoint golangci-lint version.
#MOCKERY_VERSION="2.36.0" # Optional configuration to pinpoint mockery version.
#PROTOBUF_VERSION="28.3" # Optional configuration to pinpoint protobuf version.
#PROTOC_GEN_GO_VERSION="v1.35.2" # Optional configuration to pinpoint protoc-gen-go version.
#PROTOC_GEN_GO_GRPC_VERSION="1.5.1" # Optional configuration to pinpoint protoc-gen-go-grpc version.
#PROTOC_GEN_GRPC_GATEWAY_VERSION="v2.24.0" # Optional configuration to pinpoint protoc-gen-grpc-gateway version.

PWD := $(shell pwd)

MODULES := \
    DEVGO_PATH=github.com/bool64/dev \
    DEVGRPCGO_PATH=github.com/dohernandez/dev-grpc \

GO ?= go
export GO111MODULE = on

ifneq "$(wildcard ./vendor )" ""
  modVendor =  -mod=vendor
  ifeq (,$(findstring -mod,$(GOFLAGS)))
      export GOFLAGS := ${GOFLAGS} ${modVendor}
  endif
  ifneq "$(wildcard ./vendor/github.com/dohernandez/go-grpc-service)" ""
  	DEVSERVICEGO_PATH := ./vendor/github.com/dohernandez/go-grpc-service
  endif
endif

ifeq ($(DEVSERVICEGO_PATH),)
	DEVSERVICEGO_PATH := $(shell GO111MODULE=on $(GO) list ${modVendor} -f '{{.Dir}}' -m github.com/dohernandez/go-grpc-service)
	ifeq ($(DEVSERVICEGO_PATH),)
    	$(info Module github.com/dohernandez/go-grpc-service not found, downloading.)
    	DEVSERVICEGO_PATH := $(shell export GO111MODULE=on && $(GO) get github.com/dohernandez/go-grpc-service && $(GO) list -f '{{.Dir}}' -m github.com/dohernandez/go-grpc-service)
	endif
endif

-include $(DEVSERVICEGO_PATH)/dev/makefiles/main.mk

# Add your include here with based path to the module.

-include $(DEVGO_PATH)/makefiles/lint.mk
-include $(DEVGO_PATH)/makefiles/test-unit.mk
-include $(DEVGO_PATH)/makefiles/build.mk
-include $(DEVGO_PATH)/makefiles/bench.mk

BUILD_LDFLAGS="-s -w"
BUILD_PKG = ./cmd/...
BINARY_NAME = kit-template


-include $(DEVSERVICEGO_PATH)/makefiles/dep.mk
-include $(DEVSERVICEGO_PATH)/makefiles/docker.mk
-include $(DEVSERVICEGO_PATH)/makefiles/test-integration.mk
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

check-envfile:
	@test ! -f .env && echo "Please create .env file before. Run \`make envfile\`." && exit 1 || true

## Run docker-compose up for app profile
dc-up-app: check-envfile
	@echo "Starting docker compose for application."
	@DOCKER_COMPOSE_PROFILE=app  \
	DOCKER_COMPOSE_PATH="docker-compose.yml docker-compose.app.yml" \
	make dc-up

## Run docker-compose down for app profile
dc-down-app: check-envfile
	@echo "Stopping docker compose for application."
	@DOCKER_COMPOSE_PROFILE=app \
    DOCKER_COMPOSE_PATH="docker-compose.yml docker-compose.app.yml" \
    make dc-down

## Run docker-compose up for dev profile
dc-up-dev: check-envfile
	@echo "Starting docker compose for development."
	@DOCKER_COMPOSE_PROFILE=app  \
	DOCKER_COMPOSE_PATH="docker-compose.yml docker-compose.app.yml docker-compose.dev.yml" \
	make dc-up

## Run docker-compose down for dev profile
dc-down-dev: check-envfile
	@echo "Stopping docker compose for development."
	@DOCKER_COMPOSE_PROFILE=app \
	DOCKER_COMPOSE_PATH="docker-compose.yml docker-compose.app.yml docker-compose.dev.yml" \
	make dc-down