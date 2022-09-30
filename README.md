# Kurtosis

This is the Kurtosis Mono Repo. Individual repos have their own readmes.
This repo currently contains 
- `container-engine-lib`
- `core`
- `engine`
- `api`
- `cli`

## Build Instructions

1. To build the entire project run `./scripts/build.sh`
2. To build just container-engine-lib run `./container-engine-lib/build.sh`
3. To build just the core run `./core/scripts/build.sh`
4. To build just the api run `./api/scripts/build.sh`
5. You can choose to build APIs in just one language `./api/<typescript|golang>/build.sh`
6. To build just the engine run `./engine/scripts/build.sh`
7. To regenerate the `engine` protobuf bindings do `./api/scripts/regenerate-engine-api-protobuf-bindings.sh`
8. To regenerate the `core` protobuf bindings do `./api/scripts/regenerate-core-api-protobuf-bindings.sh`
9. To build just the `cli` run `./cli/scripts/build.sh`

## Test Instructions

1. To run all `container-engine-lib` unit tests run `go test ./...` from the `kurtosis/container-engine-lib` subdirectory.
2. To run the unit tests for the core server run `go test ./...` in `core/server`
3. To run the unit tests for the core launcher run `go test ./...` in `core/launcher`
4. To run the unit tests for the engine server run `go test ./...` in `engine/server`
5. To run the unit tests for the engine launcher run `go test ./...` in `engine/launcher`
6. To run all the integration tests against Docker run `./scripts/run-all-tests-against-latest-code.sh docker`
7. To run all the integration tests against Minikube run `./scripts/run-all-tests-against-latest-code.sh minikube`

## Run instructions
1. To use the built cli run `./cli/cli/scripts/launch_cli.sh`

### Developer Notes

If you are developing the Typescript test, make sure that you have first built `api/typescript`. Any
changes made to the Typescript package within `api/typescript` aren't hot loaded as of 2022-09-29.

Running tests from the testsuite would build the `api/typescript` for you automatically so you don't have to
worry about it.