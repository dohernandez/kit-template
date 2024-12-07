# Contributing

* If you are a new contributor see: [Steps to Contribute](#steps-to-contribute).

* Relevant coding style guidelines are the [Go Code Review
  Comments](https://code.google.com/p/go-wiki/wiki/CodeReviewComments)
  and the _Formatting and style_ section of Peter Bourgon's [Go: Best
  Practices for Production
  Environments](https://peter.bourgon.org/go-in-production/#formatting-and-style).


## Steps to Contribute

Before you start contributing, make sure you have the following tools installed:

* [Go](https://golang.org/dl/) version 1.23.x or greater.
* [Docker](https://docs.docker.com/get-docker/) version 20.10.x or greater.
* [Docker Compose](https://docs.docker.com/compose/install/) version 1.29.x or greater.
* [golangci-lint](https://github.com/golangci/golangci-lint/releases) version v1.61.x or greater.
* [gofumpt](https://github.com/mvdan/gofumpt/releases) version v0.7.x or greater.
* [protoc](https://github.com/protocolbuffers/protobuf/releases) version 28.3.x or greater.
* [protoc-gen-go](https://github.com/protocolbuffers/protobuf-go/releases) version v1.35.x or greater.
* [protoc-gen-go-grpc](https://github.com/grpc/grpc-go/releases) version 1.5.x or greater.
* [protoc-gen-grpc-gateway](https://github.com/grpc-ecosystem/grpc-gateway/releases) version v2.24.x or greater.
* [protoc-gen-openapiv2](https://github.com/grpc-ecosystem/grpc-gateway/releases) version v2.24.x or greater.
* [mockery](https://github.com/vektra/mockery/releases) version v2.46.x or greater.

For quickly installing the tools (`protoc`, `protoc-gen-go`, `protoc-gen-go-grpc`, `protoc-gen-grpc-gateway`, `protoc-gen-openapiv2`, `mockery`) do:
    
```bash
make tools        # Install all require tools to work with the project
```

For quickly compiling and testing your change(s) do:

```bash
make test         # Make sure all the tests pass before you commit and push :)
```

For linting the code do:

```bash
make lint        # Make sure your change(s) follow our coding standards.
```

For checking the commit do:

```bash
make check        # Make sure test and lint pass before you commit and push :)
```

For generating proto code and api documentation do:

```bash
make proto-gen    # Generate code from proto file(s)
```

For more tools and options see the [Makefile](Makefile).

```bash
Usage
  test:                 Run tests
  check:                Check the commit compile and test the change
  tools:                Install all require tools to work with the project
  generate:             Runs commands described by directives within existing files, usage: "make generate SOURCE=<file.go... | packages>"
  proto-gen:            Generate code from proto file(s) and swagger file
  dc-up-app:            Run docker-compose up for app profile
  dc-down-app:          Run docker-compose down for app profile
  dc-up-dev:            Run docker-compose up for dev profile
  dc-down-dev:          Run docker-compose down for dev profile
  lint:                 Check with golangci-lint
  fix-lint:             Apply goimports and gofmt
  test-unit:            Run unit tests
  test-unit-multi:      Run unit tests multiple times, use `UNIT_TEST_COUNT=10 make test-unit-multi` to control count
  build-linux:          Build Linux binary
  build:                Build binary
  run:                  Build and run binary
  bench:                Run benchmark and show result stats, iterations count controlled by BENCH_COUNT, default 5.
  bench-run:            Run benchmark, iterations count controlled by BENCH_COUNT, default 5.
  bench-stat-diff:      Show benchmark comparison with base branch.
  bench-stat:           Show result of benchmark.
  deps:                 Ensure dependencies according to lock file
  env:                  Run with .env vars
  run-compile-daemon:   Run application with CompileDaemon (automatic rebuild on code change)
  build-image:          Build docker image
  dc-up:                Run docker-compose up from file DOCKER_COMPOSE_PATH with project name DOCKER_COMPOSE_PROJECT_NAME and profile DOCKER_COMPOSE_PROFILE.
                        Usage: "make dc-up PROFILE=<profile>, if PROFILE is not provide, start only default services"
  dc-down:              Run docker-compose down from file DOCKER_COMPOSE_PATH with project name DOCKER_COMPOSE_PROJECT_NAME
  dc-logs:              Run docker-compose logs from file DOCKER_COMPOSE_PATH with project name DOCKER_COMPOSE_PROJECT_NAME. Usage: "make generate APP=<docker-composer-service-name>"
  test-integration:     Run integration tests
  create-migration:     Create database migration, usage: "make create-migration NAME=<migration-name>"
  migrate:              Apply migrations
  migrate-down:         Rollback migrations
  migrate-cli:          Check/install migrations tool
  mockery-cli:          Check/install mockery tool
  protoc-cli:           Check/install protoc tool
  proto-gen-code:       Generate code from proto file(s)
  proto-gen-code-swagger:  Generate code from proto file(s) and swagger doc
```


## Pull Request

* Branch from the main branch and, if needed, rebase to the current main branch before submitting your pull request. If it doesn't merge cleanly with main you may be asked to rebase your changes.

* Commits should be as small as possible, while ensuring that each commit is correct independently (i.e., each commit should compile and pass tests).

* Add tests relevant to the fixed bug or new feature.

## Dependency management

Avoid introducing external dependencies without a good reason, but if so the project uses [Go modules](https://golang.org/cmd/go/#hdr-Modules__module_versions__and_more) to manage dependencies on external packages. This requires a working Go environment with version 1.12 or greater installed (version of the project 1.23.3).

All dependencies are vendored in the `vendor/` directory.

To add or update a new dependency, use the `go get` command:

```bash
# Pick the latest tagged release.
go get example.com/some/module/pkg

# Pick a specific version.
go get example.com/some/module/pkg@vX.Y.Z
```

Tidy up the `go.mod` and `go.sum` files and copy the new/updated dependency to the `vendor/` directory:


```bash
# The GO111MODULE variable can be omitted when the code isn't located in GOPATH.
GO111MODULE=on go mod tidy

GO111MODULE=on go mod vendor
```

You have to commit the changes to `go.mod` and `go.sum` before submitting the pull request.


Happy coding!!!