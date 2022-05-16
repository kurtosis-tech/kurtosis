# TBD
### Breaking Changes
* Engine server now requires a "KurtosisBackendType" to be defined as an input in order to launch.
* Engine server now requires Kubernetes enclave sizes to be specified in megabytes, not gigabytes
  * Remediation: change all engine input parameters specifying enclave sizes in gigabytes to megabytes

### Fixes
* Bumped to container-engine-lib 0.23.0 which has:
    * A bugfix for instantiating the DockerKurtosisBackend inside the API container
    * Actually working register-service stuff

### Changes
* Engine server can choose the correct KurtosisBackend depending on input arguments (docker, kubernetes cluster)

### Breaking Changes
* Engine server now requires a "KurtosisBackendType" to be defined as an input in order to launch.
* Upgraded to kurtosis-core 1.49.0
    * Users will need to run `kurtosis engine restart`

# 1.20.0
### Fixes
* Fixed a null pointer exception bug in launcher when backend returns an engine with no public GRPC port

### Breaking Changes
* The engine server launchers now return a public `PortSpec` rather than just a port number
    * Users should consume the port spec instead

# 1.19.0
### Changes
* Upgraded to container-engine-lib 0.21.0 and core 1.47.0

### Breaking Changes
* Remove all the following fields from the `EnclaveInfo` object returned by the API:
    * `network_id`
    * `network_cidr`
    * Users should remove these fields

# 1.18.2
### Fixes
* Added conditions to handle not having api container's public IP and ports when getting enclave response from the backend

### Changes
* Upgraded to container-engine-lib 0.20.2, which add some backend methods

# 1.18.1
### Features
* Upgrade to Core 1.46.1 and container-engine-lib 0.20.0, which allow for more Kubernetes functionality

# 1.18.0
### Changes
* The engine server no longer requires a host machine directory, as it now creates enclave data volumes for API containers

### Breaking Changes
* Upgraded to Kurtosis Core 1.46.0, which uses enclave data volumes exclusively
    * Users will need to restart their engine to start using the new Core
* Removed the `EnclaveDataDirpathOnHostMachine` property from the `EnclaveInfo` object returned by the API
    * Users interested in investigating the enclave data will now need to mount the enclave data volume on a container
* Removed the `engineDataDirpathOnHostMachine` arg from `EngineServerLauncher`'s `LaunchWithDefaultVersion` and `LaunchWithCustomVersion`
    * Users no longer need to pass in this parameter

# 1.17.5
### Changes
* Upgrade to Core 1.45.5, which uses the enclave data volume (rather than the enclave data dirpath) for storing data

# 1.17.4
### Fixes
* Fixed a bug where clean wouldn't remove empty enclaves

### Changes
* Upgraded to Core 1.45.4 to pass through service pause/unpause

# 1.17.3
### Fixes
* Upgraded to Core 1.45.3 to finally fix the bug with the Typescript Node TGZ archiver

# 1.17.2
### Fixes
* Upgraded Core to 1.45.2 to fix more bugs with node archiver

# 1.17.1
### Changes
* Upgraded to Core 1.45.1 to fix some bugs with the node archiver

# 1.17.0
### Breaking Changes
* Upgraded to Kurtosis Core 1.45.0
    * Users should follow the changes [described in the changelog](https://docs.kurtosistech.com/kurtosis-core/changelog)

# 1.16.0
### Changes
* Switched to using Kurtosis Core 1.44.0, which forces use of the files API

### Breaking Changes
* Upgraded to Kurtosis Core 1.44.0
    * Users should upgrade their engine-api-lib to version 1.44.0
* The `EnclaveContext` has been modified fairly significantly; for a list of breaking changes and remediations see [the 1.44.0 docs](https://docs.kurtosistech.com/kurtosis-core/changelog)

# 1.15.6
### Features
* Upgraded to Core 1.43.6, which has `EnclaveContext.StoreFilesFromService`

# 1.15.5
### Changes
* Upgrade to container-engine-lib 0.16.0 and core 1.43.5, which contains `EnclaveContext.StoreWebFiles`

# 1.15.4
### Fixes
* Upgraded to core 1.43.4, to fix a bug for UUID-keyed file artifacts

# 1.15.3
### Features
* Upgraded to Core 1.43.3, which has `ContainerConfigBuilder.WithFiles`

# 1.15.2
### Fixes
* Upgraded to container-engine-lib 0.15.3 to fix a bug where module enclave data volume & dirpath were getting mounted to the same place

# 1.15.1
### Features
* Sped up many things through parallelization, most notably `clean`

# 1.15.0
### Breaking Changes
* Bumped Dependencies for Kurtosis Core which is now version 1.43.0

# 1.14.5
### Changes
* Added clearer remediation steps to the error message thrown when the engine API version that `KurtosisContext` expects doesn't match the running engine version

# 1.14.4
### Fixes
* Removed the files-artifact-expansion destroy flow when destroying enclaves because it was throwing and error if volumes were still in use and now this flow was moved to `KurtosisBackend`

### Changes
* Upgrade to container-engine-lib 0.15.0
* Upgrade to kurtosis-core-api-lib 1.42.4

# 1.14.3
### Fixes
* Fixed a bug where Docker exec commands to user services were getting erroneously wrapped in `sh -c`

# 1.14.2
### Fixes
* Use container-engine-lib 0.14.3 and core 1.42.2, which supports the old port specs temporarily

# 1.14.1
### Fixes
* Upgraded to container-engine-lib 0.14.2 and core 1.42.1, which fixes a bug where stopped user services cause an error because they don't have public host port bindings

# 1.14.0
### Breaking Changes
* Updated enclave_manager to use kurtosis_backend for enclave operations
* Updated enclave_manager to use kurtosis_backend for api container operations in an enclave
* Removed cleanMetadataAcquisitionTestsuites from enclave_manager
* Removed docker specific code from enclave_manager
* Removed docker_network_allocator code

### Fixes
* Updated enclave_manager to handle reporting on stopped containers where the GRPC port_specs are nil

# 1.13.3
### Features
* Updated the destroy-enclave process adding `destroy-files-artifact-expansion-volumes` flow on it

### Changes
* Added new `KurtosisBackend` object in `EnclaveManager.NewEnclaveManager`
* Upgraded to `container-engine-lib` 0.14.0, which implement `files artifact expansion volume` and `files artifact expander` objects
* Upgraded to Kurt Core 1.41.2 which uses `files artifact expansion volumes` and `files artifact expander` objects

# 1.13.2

### Fixes
* Bumped dependencies for Core 1.41.1 and Container Engine Lib 0.13.0

# 1.13.1
### Changes
* Upgrade to container-engine-lib 0.12.0

# 1.13.0
### Breaking Changes
* Bumped Dependencies for Kurtosis Core which is now version 1.41.0
  * Users using the ExecuteBulkCommands API should remove code referencing it.
  * Additionally Enclaves should be restarted.

### Fixes
* Added CGO_ENABLED for go tests.

# 1.12.0
### Breaking Changes
* Upgraded to Kurt Core 1.40.1, which removes external-container-registration features

# 1.11.5
### Features
* Use container-engine-lib 0.10.1 and core 1.39.9 to provide API container & enclave CRUD functions

# 1.11.4
### Features
* Changed `launcher` to use generic `KurtosisBackend` instead of `DockerManager`

### Fixes
* Upgraded core & container-engine-lib dependencies to fix bug where Docker containers in the `removing` state were counted as running
* Don't treat containers in the `removing` state as running

# 1.11.3
### Changes
* Upgrade to container-engine-lib 0.9.0

# 1.11.2
### Features
* Added GenericEngineClient interface to abstract grpc-web vs grpc-js Typescript backends

# 1.11.1
### Changes
* Kurtosis enclaves created via `KurtosisContext.CreateEnclave` default to the debug loglevel
* Upgrade to container-engine-lib 0.8.7 which contains dormant Kubernetes code

# 1.11.0
### Breaking Changes
* Upgraded to Kurt Core 1.39.0, which will no properly report its own version

### Changes
* Product analytic events are sent when the user attempts to make the action, rather than after the action succeeded, so that we're not dropping actions on error

### Fixes
* The metrics produced by the engine server now properly report their own version

# 1.10.5
### Changes
* Upgraded to object-attributes-schema-lib 0.7.2 and core 1.38.2 which allows for 256-char label values

# 1.10.4
### Fixes
* Actually handle old port specs

# 1.10.3
### Fixes
* Don't break on old port spec (`portId:1234/TCP`)

# 1.10.2
### Fixes
* Fixes a panic due to passing `nil` into `stacktrace.Propagate`
* Fix some variable name shadowing in `EnclaveManager`

# 1.10.1
### Changes
* Upgrade to `object-attributes-schema-lib` v0.7.1
* Upgrade to `kurtosis-core` v1.38.1

### Fixes
* Don't choke when trying to destroy an enclave whose enclave ID is a subset of another

# 1.10.0
### Breaking Changes
* `kurtosis-core` API changed : `ApiContainerLauncher.LaunchWithDefaultVersion()` and `ApiContainerLauncher.LaunchWithCustomVersion()` now adds two new arguments `grpcListenPort` and `grpcProxyListenPort` and deleted the one named `listenPort`
  * Users should update, `object-attributes-schema-lib` to v0.7.0 and `kurtosis-engine` to v.1.10.0

### Changes
* Upgrade to `object-attributes-schema-lib` v0.7.0
* Upgrade to `kurtosis-core` v1.38.0

# 1.9.2
### Fixes
* Fix `kurtosis-core` dependency version in Typescript library, upgraded to v1.37.1

# 1.9.1
### Changes
* Add metrics client close call to flush the queue
* Upgrade to `metrics-client-library` v0.1.2
* Upgrade to `kurtosis-core` v1.37.1

# 1.9.0
### Features
* Added metrics client to track enclave events (e.g.: when users create an enclave)

### Fixes
* Refactored engine/library semVer checking to be more lax, and let continue the code run if semVer couldn't be parsed

### Changes
* Upgraded to `Kurt Core` 1.37.0 which implements module's metrics

### Breaking Changes
* Change the `EngineServerLauncher.LaunchWithDefaultVersion()` and `EngineServerLauncher.LaunchWithCustomVersion()` methods API, adding two new arguments `metricsUserID` and `didUserAcceptSendingMetrics`
  * Users should add these two new arguments in every call
* Change `NewEngineServerService` constructor to now receive three new arguments `metricsUserID`, `didUserAcceptSendingMetrics` and `metricsClient`
  * Users should add there three new arguments in every call
* Change `EngineServerArgs` constructor, adding two new arguments `metricsUserID` and `didUserAcceptSendingMetrics`
  * Users should add these two new arguments in every call

# 1.8.3
### Fixes
* Upgraded to Kurtosis Core v1.36.12 which fixes a bug when creating soft network partitions in Tyepscript

# 1.8.2
### Features
* Added deletion of dangling folders in clean endpoint

# 1.8.1
### Fixes
* Fixed `KurtosisContext.newKurtosisContextFromLocalEngine` method converting it to static method again

# 1.8.0
### Features
* Added a public constant `KurtosisEngineVersion` in golang and typescript libraries
* Added a validation to check if the running engine version is the expected in `KurtosisContext` creation
* Added a user-friendly error text when a client try to create a `KurtosisContext` but the engine is unavailable

### Breaking Changes
* Renamed the public constant, `DefaultVersion` to `KurtosisEngineVersion` which is more descriptive
  * Users should replace the constant name in their implementations

# 1.7.7
### Fixes
* Bump to Core 1.36.11, which attempts to pull module & user service images
* Added small fix in clean endpoint

# 1.7.6
### Fixes 
* Added mutex lock in GetEnclaves

# 1.7.5
### Features
* Added clean endpoint which gets rid of the enclaves and the dependant containers

# 1.7.4
### Changes
* Upgraded to `Kurt Core` 1.36.10 which adds the latest version of `object-attributes-schema-lib`
* Upgrade to `object-attributes-schema-lib` v0.6.0 to support `id` label

# 1.7.3
### Fixes
* Upgraded to Kurt Core 1.36.9, which fixes a bug with an error value not getting checked

# 1.7.2
### Fixes
* Upgrade to Core 1.36.8, in an attempt to fix a weird error around protobuf empty object

# 1.7.1
### Changes
* Switch to using `@grpc/grpc-js` in the Typescript library for communicating with the engine, as the `grpc` package is deprecated now
* Upgrade to Kurt Core 1.36.7 to support the `@grpc/grpc-js` upgrade

# 1.7.0
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
