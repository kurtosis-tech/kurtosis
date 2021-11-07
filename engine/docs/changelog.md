# TBD

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
* Upgraded to [engine API lib 0.6.0](https://github.com/kurtosis-tech/kurtosis-engine-api-lib/blob/develop/docs/changelog.md#060)

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
* Upgrade to [engine API lib 0.4.0](https://github.com/kurtosis-tech/kurtosis-engine-api-lib/blob/develop/docs/changelog.md#040)

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
