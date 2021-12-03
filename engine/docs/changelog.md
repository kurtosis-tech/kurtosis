# TBD
### Features
* Added a unit test that will remind Kurtosis developers to add a breaking change to the changelog whenever they bump to a Kurt Core version that has an API break (since the engine server's API breaks when the Core API breaks due to the engine server's `KurtosisContext` returning a Core `EnclaveContext`)

### Breaking Changes
* Upgraded to kurtosis-core 1.36.0, which changes the way partition connection information is defined during repartitioning
  * Users should see [the Kurtosis Core changelog on the topic](https://docs.kurtosistech.com/kurtosis-core/changelog#breaking-changes) and implement the remediation there

# 1.6.0
_There aren't any changes in this release; it is being released to represent the breaking API change that should have happened in 1.5.7 due to Kurt Core's API version getting bumped_

# 1.5.7
### Features
* Upgraded to Kurt Core v1.35.0, which allows users to add user-friendly port IDs to the ports that they declare for their containers

### Changes
* Upgraded to minimal-grpc-server 0.5.0

### Removals
* Removed the `ListenProtocol` engine server arg, because it listens on gRPC and only TCP is supported

# 1.5.6
### Fixes
* Fixed an issue where getting enclaves that included an API container without the new `port-protocol` label would break

# 1.5.5
### Fixes
* All `stacktrace.Propagate` calls now panic on a `nil` error

# 1.5.4
### Changes
* Added an explanatory comment as to why all the directories that the engine server creates inside the engine data directory must be created with `0777` permissions

### Fixes
* Use Kurt Core 1.33.3, which creates its directories inside the enclave data dir with `0777` permissions rather than `0755`
* Make sure directories created inside the engine data directory are created with `0777` permissions rather than `0755`

# 1.5.3
### Fixes
* Don't break on old API containers

# 1.5.2
### Fixes
* Fixed bug where the launcher would always launch the `latest` version

# 1.5.1
### Fixes
* Take Core 1.33.1 and obj-attrs-schema-lib 0.3.1, which fixes a bug where containers wouldn't get the forever-labels applied to them

# 1.5.0
### Changes
* Got rid of the launcher's `GetDefaultVersion` method in favor of a public constant, `DefaultVersion`,  because the old method required instantiating a launcher to get the default version
* Upgraded to kurtosis-core 1.33.0
* Upgraded to obj-attr-schema-lib 0.3.0

### Breaking Changes
* Got rid of the launcher's `GetDefaultVersion` method in favor of a public constant
    * Users should use the `DefaultVersion` constant instead

# 1.4.0
### Changes
* Upgrade to Kurt Core 1.32.0
* Renamed `EngineServerLauncher.GetDefaultImageVersionTag` -> `GetDefaultVersion`

### Breaking Changes
* Renamed `EngineServerLauncher.GetDefaultImageVersionTag` -> `GetDefaultVersion`
    * Users should change their code appropriately

# 1.3.0
### Features
* The engine server launcher now exposes `DefaultImageVersionTag`, for viewing what version of the engine the launcher would start

### Removals
* Removed the own-version constants in the Golang & Typescript API submodules

### Breaking Changes
* The API no longer has own-version constants
    * If users are using this, they should evaluate why (and whether they should be using the `launcher` submodule instead, which has an own-version constant) because the API shouldn't know about it own version
* The `GetEngineInfoResponse` object no longer has an `engine_api_version` field, as there is no more distinction between engine version and API version

# 1.2.2
### Features
* The engine server launcher now exposes its `ListenProtocol`

### Changes
* Upgrade to `object-attributes-schema-lib` 0.2.0
* Upgrade to Kurt Core 1.31.2

# 1.2.1
### Features
* Added `EngineVersion` field that represents the engine server version in the `GetEngineInfo` response

# 1.2.0
### Features
* The launcher makes a best-effort attempt to pull the latest version of the image being started

### Changes
* Upgrade to Kurt Core 1.31.1
* `CreateEnclave` now takes in the version tag of the API container to use, and an emptystring indicates that the engine server should use its own default version
* The launcher's `Launch` function has been replaced with `LaunchWithCustomVersion` and `LaunchWithDefaultVersion`

### Breaking Changes
* `CreateEnclaveArgs`'s `api_container_image` property has been replaced with `api_container_image_version_tag`
    * Users should leave this blank if they want the default API container, and set it to use a custom version of the API container
* `EngineServerLauncher.Launch` no longer exists
    * Users should use either `LaunchWithCustomVersion` or `LaunchWithDefaultVersion` depending on their needs

# 1.1.0
### Changes
* Use Node 16.13.0

### Breaking Changes
* The Typescript library now depends on Node 16.13.0

# 1.0.0
### Changes
* Refactored the entire structure of this repo to now contain the API

### Breaking Changes
* The API now lives inside this repo in the `github.com/kurtosis-tech/kurtosis-engine-server/api/golang` package

# 0.6.0
### Changes
* Upgraded to Engine API Lib 0.11.0, which uses Kurt Core API Lib rather than Kurt Client
* Stop depending on Kurt Core server directly, and switch to depending on the launcher submodule
* Switch to using `object-attributes-schema-lib`, rather than the custom-defined attributes schema from Kurt Core
* Switch to using `free-ip-addr-tracker-lib` rather than a version from Kurt Core

### Breaking Changes
* Upgraded to Engine API Lib 0.11.0

# 0.5.3
### Fixes
* Use `container-engine-lib` 0.8.3, which returns host port bindings with 127.0.0.1 IP address rather than `0.0.0.0`

# 0.5.2
### Fixes
* Use Kurt Core 1.27.3, which returns host machine port bindings in `127.0.0.1` (rather than `0.0.0.0`) for Windows users

### Changes
* Go back down to Engine API Lib 0.9.0, because going up to 0.10.0 should have been an API break

# 0.5.1
### Features
* The `EnclaveInfo` object has a new enclave-data-dirpath-on-host-machine property

### Changes
* Implement the Engine API Lib 0.10.0

# 0.5.0
### Features
* The loglevel of the engine server can now be controlled via a `logLevelStr` JSON args property

### Changes
* Changed the way enclave data is stored, which were done in preparation of merging the APIC and engine container:
    * An "engine data directory" is created on the Docker host machine
    * That directory is bind-mounted into the engine container
    * The engine creates enclave directories inside that engine data dir
    * The enclave directories are bind-mounted into the modules/services of the enclave

### Breaking Changes
* The engine container now requires a `SERIALIZED_ARGS` environment variable containing JSON-serialized args to run the engine server with

# 0.4.7
### Changes
* When `CreateEnclave` fails halfway through, the created API container will be killed rather than stopped (as there's no reason to wait)

# 0.4.6
### Changes
* Pull in Kurt Core 1.26.4, which no longer stops containers when the API container is exiting

# 0.4.5
### Features
* Upgrade to Kurt Core 1.26.3, which tags testsuite containers with their type

# 0.4.4
### Fixes
* Fixed bug where files artifact expansion volume wouldn't get deleted

# 0.4.3
### Fixes
* Use the `EnclaveObjectLabelsProvider` for finding Kurtosis networks & enclave data volumes

# 0.4.2
### Fixes
* Fixed issue where `DestroyEnclave` would expect exactly one enclave volume (which isn't true for enclaves where files artifact expansion is done)

# 0.4.1
### Features
* Upgraded to Kurt Core 1.26.1, which adds a framework for labelling testsuite containers

# 0.4.0
### Fixes
* Fixed bug where the nonexistent enclave check wasn't working
* Upgraded to engine-api-lib 0.7.2, which allows for the case where the API container isn't running (which means it won't have host machine info)
* Fixed bug where `DestroyEnclave` would hang due to reentrant mutex issues

### Breaking Changes
* The `EnclaveAPIContainerInfo` object that gets returned inside `EnclaveInfo` has had its host machine info split off into `EnclaveAPIContainerHostMachineInfo`, which will only be populated if the API container is running

# 0.3.0
### Fixes
* Upgraded to engine-api-lib 0.7.1, which contains various bugfixes

### Breaking Changes
* All Go `KurtosisContext` methods now take a context

# 0.2.0
### Features
* Upgraded to engine API lib 0.6.0

### Breaking Changes
* `GetEnclave` endpoint is replaced with `GetEnclaves`
    * Users should switch to `GetEnclaves` instead
* `CreateEnclave`'s return value has been replaced with an `EnclaveInfo` object
    * Users should consume the `EnclaveInfo` object instead

# 0.1.8
### Features
* Upgrade to `engine-api-lib` 0.4.2, which adds the `GetEngineInfo` endpoint

# 0.1.7
### Fixes
* Try again to fix the binary working inside Alpine Linux

# 0.1.6
### Fixes
* Fix issue with binary not being able to run on Alpine Linux

# 0.1.5
### Changes
* Upgrade to engine API lib 0.4.0

# 0.1.4
### Features
* Added a `KurtosisEngineServerVersion` constant that gets set to this repo's version

### Removals
* Removed non-functioning, unused `update-package-version.sh` script

# 0.1.3
### Features
* Publish server as a Docker image to `kurtosistech/kurtosis-engine-server`

# 0.1.2
### Changes
* Upgraded to `kurtosis-core` 1.25.2, which brings in container-engine-lib 0.7.0 (and with it, the ability to specify host machine port bindings for containers)

# 0.1.1
### Features
* Try to pull `api-container` latest image before running the API container Docker container
* Upgraded to Kurt Core 1.25.1, which add `com.kurtosistech.app-id` container label to all enclave containers
* Added a `StopEnclave` endpoint

### Changes
* `DestroyEnclave` endpoint actually destroys the objects associated with the enclave (e.g. network, containers, volume, etc.)

### Fixes
* Added a mutex to `EnclaveManager` to fix race conditions when modifying enclaves

# 0.1.0
### Features
* Added `EngineServerService` which will be in charge of receive requests for creating and destroying Kurtosis Enclaves
* Ported the `EnclaveManager` from `Kurtosis CLI` to here and it also includes `DockerNetworkAllocator`
* Added `MinimalRPCServer` used to implement the Kurtosis Engine RPC Server
* Added `build` script to automatize building Engine Server process
* Added `Docker` file for Kurtosis Engine Server Docker image
* Added `get-docker-image-tag` script to automatize Docker image tag creation process
* Added `.dockerignore` to enable Docker caching
* Added `Enclave` struct in `enclave_manager` packaged

### Changes
* Removed check Typescript job and publish Typescript artifact job in Circle CI configuration

### Removals
* Removed the `log` inside `EnclaveManager`, as it's no longer needed
