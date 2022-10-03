# TBD
### Features
* Created new engine's endpoint `GetUserServiceLogs` for consuming user service container logs from the logs database server
* Added `LogsDatabaseClient` interface for defining the behaviour for consuming logs from the centralized logs database
* Added `LokiLogsDatabaseClient` which implements `LogsDatabaseClient` for consuming logs from a Loki's server
* Added `KurtosisBackendLogsClient` which implements `LogsDatabaseClient` for consuming user service container logs using `KurtosisBackend
* Created the `LogsDatabase` object in `container-engine-lib`
* Created the `LogsCollector` object in `container-engine-lib`
* Added `LogsDatabase` CRUD methods in `Docker` Kurtosis backend
* Added `LogsCollector` CRUD methods in `Docker` Kurtosis backend

### Breaking Changes
* Updated `CreateEngine` method in `container-engine-lib`, removed the `logsCollectorHttpPortNumber` parameter
    * Users will need to update all the `CreateEngine` calls removing this parameter
* Updated `NewEngineServerArgs`,  `LaunchWithDefaultVersion` and `LaunchWithCustomVersion` methods in `engine_server_launcher` removed the `logsCollectorHttpPortNumber` parameter
  * Users will need to update these method calls removing this parameter
  
### Changes
* Untied the logs components containers and volumes creation and removal from the engine's crud in `container-engine-lib`

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
