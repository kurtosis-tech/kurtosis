# Kurtosis

This is the Kurtosis Mono Repo. Individual repos have their own readmes.
This repo currently contains 
- `container-engine-lib`
- `core`

## Build Instructions

1. To build the entire project run `./scripts/build.sh`
2. To build just container-engine-lib run `./container-engine-lib/build.sh`
3. To build just the core run `./core/scripts/build.sh`

## Test Instructions

1. To run all `container-engine-lib` unit tests run `go test ./...` from the `kurtosis/container-engine-lib` subdirectory.
2. To run the unit tests for the core server run `go test ./...` in `core/server`
3. To run the unit tests for the core launcher run `go test ./...` in `core/launcher`

## Run instructions