# TBD

# 0.50.1

### Fixes
* Fix generate scripts to take passed version on release

# 0.50.0
### Features
* Created the `LogsDatabase` object in `container-engine-lib`
* Created the `LogsCollector` object in `container-engine-lib`
* Added `LogsDatabase` CRUD methods in `Docker` Kurtosis backend
* Added `LogsCollector` CRUD methods in `Docker` Kurtosis backend
* Added `ServiceNetwork` (interface), `DefaultServiceNetwork` and `MockServiceNetwork` 

### Breaking Changes
* Updated `CreateEngine` method in `container-engine-lib`, removed the `logsCollectorHttpPortNumber` parameter
    * Users will need to update all the `CreateEngine` calls removing this parameter
* Updated `NewEngineServerArgs`,  `LaunchWithDefaultVersion` and `LaunchWithCustomVersion` methods in `engine_server_launcher` removed the `logsCollectorHttpPortNumber` parameter
  * Users will need to update these method calls removing this parameter
  
### Changes
* Untied the logs components containers and volumes creation and removal from the engine's crud in `container-engine-lib`
* Made some changes to the implementation of the module manager based on some PR comments by Kevin

### Features
* Implement Startosis add_service image pull validation
* Startosis scripts can now be run from the CLI: `kurtosis startosis exec path/to/script/file --enclave-id <ENCLAVE_ID>`
* Implemented Startosis load method to load from Github repositories

### Fixes
* Fix IP address placeholder injected by default in Startosis instructions. It used to be empty, which is invalid now
it is set to `KURTOSIS_IP_ADDR_PLACEHOLDER`
* Fix enclave inspect CLI command error when there are additional port bindings
* Fix a stale message the run-all-test-against-latest-code script
* Fix bug that creates database while running local unit tests
* Manually truncate string instead of using `k8s.io/utils/strings`

### Removals
* Removes version constants within launchers and cli in favor of centralized generated version constant
* Removes remote-docker-setup from the `build_cli` job in Circle

# 0.49.9

### Features
* Implement Startosis add_service method
* Enable linter on Startosis codebase

# 0.49.8
### Changes
* Added a linter
* Made changes based on the linters output
* Made the `discord` command a LowLevelKurtosisCommand instead of an EngineConsumingKurtosisCommand

### Features
* API container now saves free IPs on a local database

### Fixes
* Fix go.mod for commons & cli to reflect monorepo and replaced imports with write package name
* Move linter core/server linter config to within core/server

# 0.49.7
### Features
* Implement skeleton for the Startosis engine

### Fixes
* Fixed a message that referred to an old repo.

### Changes
* Added `cli` to the monorepo

# 0.49.6
### Fixes
* Fixed a bug where engine launcher would try to launch older docker image `kurtosistech/kurtosis-engine-server`.

# 0.49.5
### Changes
* Added `kurtosis-engine-server` to the monorepo
* Merged the `kurtosis-engine-sdk` & `kurtosis-core-sdk`

### Removals
* Remove unused variables from Docker Kurtosis backend

# 0.49.4
### Fixes
* Fix historical changelog for `kurtosis-core`
* Don't check for grpc proxy to be available

# 0.49.3
### Fixes
* Fix typescript package releases

# 0.49.2
### Removals
* Remove envoy proxy from docker image. No envoy proxy is being run anymore, effectively removing HTTP1.

### Changes
* Added `kurtosis-core` to the monorepo

### Fixes
* Fixed circle to not docs check on merge

# 0.49.1
### Fixes
* Attempting to fix the release version
### Changes
* Added container-engine-lib

# 0.49.0
### Changes
* This version is a dummy version to set the minimum. We pick a version greater than the current version of the CLI (0.29.1). 
