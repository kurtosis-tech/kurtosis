# TBD
### Changes
- Added `startosis_add_service_with_empty_ports` Golang and Typescript internal tests
- Added exponential back-off and retries to `get_value`

### Fixes
- Make validation more human-readable for missing docker images and instructions that depend on invalid service ids

### Changes
- Make arg parsing errors more explicit on structs
- Updated Starlark section of core-lib-documentation.md to match the new streaming endpoints
- Updated `datastore-army-module` -> `datastore-army-package`

### Features
- Log file name and function like [filename.go:FunctionName()] while logging in `core` & `engine`
- All Kurtosis instructions now returns a simple but explicit output
- The object returned by Starlark's `run()` function is serialized as JSON and returned to the CLI output.
- Enforce `run(args)` for individual scripts 

# 0.57.1

### Changes
- Added tab-completion (suggestions) to commands that require Service GUIDs, i.e.  `service shell` and `service logs` paths

# 0.57.0
### Breaking Changes
- Renamed `src_path` parameter in `read_file` to `src`
  - Users will have to upgrade their `read_file` calls to reflect this change

### Features
- Progress information (spinner, progress bar and quick progress message) is now printed by the CLI
- Instruction are now printed before the execution, and the associated result is printed once the execution is finished. This allows failed instruction to be printed before the error message is returned.

### Breaking changes
- Endpoints `ExecuteStartosisScript` and `ExecuteStartosisModule` were removed
- Endpoints `ExecuteKurtosisScript` was renamed `RunStarlarkScript` and `ExecuteKurtosisModule` was renamed `RunStarlarkPackage`

### Changes
- Starlark execution progress is now returned to the CLI via the KurtosisExecutionResponseLine stream
- Renamed `module` to `package` in the context of the Startosis engine

### Fixes
- Fixed the error message when the relative filename was incorrect in a Starlark import
- Fixed the error message when package name was incorrect
- Don't proceed with execution if there are validation errors in Starlark
- Made missing `run` method interpretation error more user friendly

# 0.56.0
### Breaking Changes
- Removed `module` key in the `kurtosis.yml` (formerly called `kurtosis.mod`) file to don't have nested keys
  - Users will have to update their `kurtosis.yml` to remove the key and move the `name` key in the root

# 0.55.1

### Changes
- Re-activate tests that had to be skipped because of the "Remove support for protobuf in Startosis" breaking change
- Renamed `input_args` to `args`. All Starlark packages should update `run(input_args)` to `run(args)`

# 0.55.0
### Fixes
- Fix failing documentation tests by linking to new domain in `cli`
- Fix failing `docs-checker` checks by pointing to `https://kurtosis-tech.github.io/kurtosis/` instead of `docs.kurtosistech.com`

### Breaking Changes
- Renamed `kurtosis.mod` file to `kurtosis.yml` this file extension enable syntax highlighting
  - Users will have to rename all theirs `kurtosis.mod` files

### Changes
- Made `run` an EngineConsumingKurtosisCommand, i.e. it automatically creates an engine if it doesn't exist
- Added serialized arguments to KurtosisInstruction API type such that the CLI can display executed instructions in a nicer way.

### Features
- Added one-off HTTP requests, `extract` and `assert`

# 0.54.1
### Fixes
- Fixes a bug where the CLI was returning 0 even when an error happened running a Kurtosis script

### Changes
- Small cleanup in `grpc_web_api_container_client` and `grpc_node_api_container_client`. They were implementing executeRemoteKurtosisModule unnecessarily

# 0.54.0
### Breaking Changes
- Renamed `kurtosis exec` to `kurtosis run` and `main in main.star` to `run in main.star`
  - Upgrade to the latest CLI, and use the `run` function instead
  - Upgrade existing modules to have `run` and not `main` in `main.star`

### Features
- Updated the CLI to consume the streaming endpoints to execute Startosis. Kurtosis Instructions are now returned live, but the script output is still printed at the end (until we have better formatting).
- Update integration tests to consume Startosis streaming endpoints

# 0.53.12
### Changes
- Changed occurrences of `[sS]tartosis` to `Starlark` in errors sent by the CLI and its long and short description
- Changed some logs and error messages inside core that which had references to Startosis to Starlark
- Allow `dicts` & `structs` to be passed to `render_templates.config.data`

# 0.53.11
### Changes
- Published the log-database HTTP port to the host machine

# 0.53.10
### Changes
- Add 2 endpoints to the APIC that streams the output of a Startosis script execution
- Changed the syntax of render_templates in Starlark

### Fixes
- Fixed the error that would happen if there was a missing `kurtosis.mod` file at the root of the module

# 0.53.9
### Fixes
- Renamed `artifact_uuid` to `artifact_id` and `src` to `src_path` in `upload_files` in Starlark

# 0.53.8

# 0.53.7
### Changes
- Modified the `EnclaveIdGenerator` now is a user defined type and can be initialized once because it contains a time-seed inside
- Simplify how the kurtosis instruction canonicalizer works. It now generates a single line canonicalized instruction, and indentation is performed at the CLI level using Bazel buildtools library.

### Fixes
- Fixed the `isEnclaveIdInUse` for the enclave validator, now uses on runtime for `is-key-in-map`

### Features
- Add the ability to execute remote modules using `EnclaveContext.ExecuteStartoisRemoteModule`
- Add the ability to execute remote module using cli `kurtosis exec github.com/author/module`

# 0.53.6

# 0.53.5
### Changes
- Error types in ExecuteStartosisResponse type is now a union type, to better represent they are exclusive and prepare for transition to streaming
- Update the KurtosisInstruction API type returned to the CLI. It now contains a combination of instruction position, the canonicalized instruction, and an optional instruction result 
- Renamed `store_files_from_service` to `store_service_files`
- Slightly update the way script output information are passed from the Startosis engine back the API container main class. This is a step to prepare for streaming this output all the way back the CLI.
- Removed `load` statement in favour of `import_module`. Calling load will now throw an InterpretationError
- Refactored startosis tests to enable parallel execution of tests

# 0.53.4

# 0.53.3
### Fixes
- Fixed a bug with dumping enclave logs during the CI run

### Features
- Log that the module is being compressed & uploaded during `kurtosis exec`
- Added `file_system_path_arg` in the CLI which provides validation and tab auto-completion for filepath, dirpath, or both kind of arguments
- Added tab-auto-complete for the `script-or-module-path` argument in `kurtosis exec` CLI command

### Changes
- `print()` is now a regular instructions like others, and it takes effect at execution time (used to be during interpretation)
- Added `import_module` startosis builtin to replace `load`. Load is now deprecated. It can still be used but it will log a warning. It will be entirely removed in a future PR
- Added exhaustive struct linting and brought code base into exhaustive struct compliance
- Temporarily disable enclave dump for k8s in CircleCI until we fix issue #407
- Small cleanup to kurtosis instruction classes. It now uses a pointer to the position object.

### Fixes
- Renamed `cmd_args` and `entrypoint_args` inside `config` inside `add_service` to `cmd` and `entrypoint`

### Breaking Changes
- Renamed `cmd_args` and `entrypoint_args` inside `config` inside `add_service` to `cmd` and `entrypoint`
  - Users will have to replace their use of `cmd_args` and `entry_point_args` to the above inside their Starlark modules 

# 0.53.2
### Features
- Make facts referencable on `add_service`
- Added a new log line for printing the `created enclave ID` just when this is created in `kurtosis exec` and `kurtosis module exec` commands

# 0.53.1
### Features
- Added random enclave ID generation in `EnclaveManager.CreateEnclave()` when an empty enclave ID is provided
- Added the `created enclave` spotlight message when a new enclave is created from the CLI (currently with the `enclave add`, `module exec` and `exec` commands)

### Changes
- Moved the enclave ID auto generation and validation from the CLI to the engine's server which will catch all the presents and future use cases

### Fixes
- Fixed a bug where we had renamed `container_image_name` inside the proto definition to `image`
- Fix a test that dependent on an old on existent Starlark module

# 0.53.0
### Features
- Made `render_templates`, `upload_files`, `store_Files_from_service` accept `artifact_uuid` and
return `artifact_uuid` during interpretation time
- Moved `kurtosis startosis exec` to `kurtosis exec`

### Breaking Features
- Moved `kurtosis startosis exec` to `kurtosis exec`
  - Users now need to use the new command to launch Starlark programs

### Fixes
- Fixed building kurtosis by adding a conditional to build.sh to ignore startosis folder under internal_testsuites

# 0.52.5
### Fixes
- Renamed `files_artifact_mount_dirpaths` to just `files`

# 0.52.4
### Features
- Added the enclave's creation time info which can be obtained through the `enclave ls` and the `enclave inspect` commands

### Fixes
- Smoothened the experience `used_ports` -> `ports`, `container_image_name` -> `name`, `service_config` -> `config`

# 0.52.3
### Changes
- Cleanup Startosis interpreter predeclared

# 0.52.2

# 0.52.1
### Features
- Add `wait` and `define` command in Startosis
- Added `not found service GUIDs information` in `KurtosisContext.GetServiceLogs` method
- Added a warning message in `service logs` CLI command when the request service GUID is not found in the logs database
- Added ip address replacement in the JSON for `render_template` instruction

### Changes
- `kurtosis_instruction.String()` now returns a single line version of the instruction for more concise logging

### Fixes
- Fixes a bug where we'd propagate a nil error
- Adds validation for `service_id` in `store_files_from_service`
- Fixes a bug where typescript (jest) unit tests do not correctly wait for grpc services to become available
- Fixed a panic that would happen cause of a `nil` error being returned
- Fixed TestValidUrls so that it checks for the correct http return code

# 0.52.0
### Breaking Changes
- Unified `GetUserServiceLogs` and `StreamUserServiceLogs` engine's endpoints, now `GetUserServiceLogs` will handle both use cases
  - Users will have to re-adapt `GetUserServiceLogs` calls and replace the `StreamUserServiceLogs` call with this
- Added the `follow_logs` parameter in `GetUserServiceLogsArgs` engine's proto file
  - Users should have to add this param in all the `GetUserServiceLogs` calls
- Unified `GetUserServiceLogs` and `StreamUserServiceLogs` methods in `KurtosisContext`, now `GetUserServiceLogs` will handle both use cases
  - Users will have to re-adapt `GetUserServiceLogs` calls and replace the `StreamUserServiceLogs` call with this
- Added the `follow_logs` parameter in `KurtosisContext.GetUserServiceLogs`
  - Users will have to addition this new parameter on every call

### Changes
- InterpretationError is now able to store a `cause`. It simplifies being more explicit on want the root issue was
- Added `upload_service` to Startosis
- Add `--args` to `kurtosis startosis exec` CLI command to pass in a serialized JSON
- Moved `read_file` to be a simple Startosis builtin in place of a Kurtosis instruction

# 0.51.13
### Fixes
- Set `entrypoint` and `cmd_args` to `nil` if not specified instead of empty array 

# 0.51.12
### Features
- Added an optional `--dry-run` flag to the `startosis exec` (defaulting to false) command which prints the list of Kurtosis instruction without executing any. When `--dry-run` is set to false, the list of Kurtosis instructions is printed to the output of CLI after being executed.

# 0.51.11
### Features
- Improve how kurtosis instructions are canonicalized with a universal canonicalizer. Each instruction is now printed on multiple lines with a comment pointing the to position in the source code.
- Support `private_ip_address_placeholder` to be passed in `config` for `add_service` in Starlark

### Changes
- Updated how we generate the canonical string for Kurtosis `upload_files` instruction

# 0.51.10
### Changes
- Added Starlark `proto` module, such that you can now do `proto.has(msg, "field_name")` in Startosis to differentiate between when a field is set to its default value and when it is unset (the field has to be marked as optional) in the proto file though.

# 0.51.9
### Features
- Implemented the new `StreamUserServiceLogs` endpoint in the Kurtosis engine server
- Added the new `StreamUserServiceLogs` in the Kurtosis engine Golang library
- Added the new `StreamUserServiceLogs` in the Kurtosis engine Typescript library
- Added the `StreamUserServiceLogs` method in Loki logs database client
- Added `stream-logs` test in Golang and Typescript `internal-testsuites`
- Added `service.GUID` field in `Service.Ctx` in the Kurtosis SDK

### Changes
- Updated the CLI `service logs` command in order to use the new `KurtosisContext.StreamUserServiceLogs` when user requested to follow logs
- InterpretationError is now able to store a `cause`. It simplifies being more explicit on want the root issue was
- Added `upload_service` to Startosis
- Add `--args` to `kurtosis startosis exec` CLI command to pass in a serialized JSON

# 0.51.8
### Features
- Added exec and HTTP request facts
- Prints out the instruction line, col & filename in the execution error
- Prints out the instruction line, col & filename in the validation error
- Added `remove_service` to Startosis

### Fixes
- Fixed nil accesses on Fact Engine

### Changes
- Add more integration tests for Kurtosis modules with input and output types

# 0.51.7
### Fixes
- Fixed instruction position to work with nested functions

### Features
- Instruction position now contains the filename too

# 0.51.6
### Features
- Added an `import_types` Starlark instruction to read types from a .proto file inside a module
- Added the `time` module for Starlark to the interpreter
- Added the ability for a Starlark module to take input args when a `ModuleInput` in the module `types.proto` file

# 0.51.5
### Fixes
- Testsuite CircleCI jobs also short-circuit if the only changes are to docs, to prevent them failing due to no CLI artifact

# 0.51.4
### Fixes
- Fixed a bug in `GetLogsCollector` that was failing when there is an old logs collector container running that doesn't publish the TCP port
- Add missing bindings to Kubernetes gateway

### Changes
- Adding/removing methods from `.proto` files will now be compile errors in Go code, rather than failing at runtime
- Consolidated the core & engine Protobuf regeneration scripts into a single one

### Features
- Validate service IDs on Startosis commands

# 0.51.3
### Fixes
- Added `protoc` install step to the `publish_api_container_server_image` CircleCI task

# 0.51.2
### Features
- Added a `render_templates` command to Startosis
- Implemented backend for facts engine
- Added a `proto_file_store` in charge of compiling Startosis module's .proto file on the fly and storing their FileDescriptorSet in memory

### Changes
- Simplified own-version constant generation by checking in `kurtosis_version` directory

# 0.51.1
- Added an `exec` command to Startosis
- Added a `store_files_from_service` command to Startosis
- Added the ability to pass `files` to the service config
- Added a `read_file` command to Startosis
- Added the ability to execute local modules in Startosis

### Changes
- Fixed a typo in a filename

### Fixes
- Fixed a bug in exec where we'd propagate a `nil` error
- Made the `startosis_module_test` in js & golang deterministic and avoid race conditions during parallel runs

### Removals
- Removed  stale `scripts/run-pre-release-scripts` which isn't used anywhere and is invalid.

# 0.51.0
### Breaking Changes
- Updated `kurtosisBackend.CreateLogsCollector` method in `container-engine-lib`, added the `logsCollectorTcpPortNumber` parameter
  - Users will need to update all the `kurtosisBackend.CreateLogsCollector` setting the logs collector `TCP` port number 

### Features
- Added `KurtosisContext.GetUserServiceLogs` method in `golang` and `typescript` api libraries
- Added the public documentation for the new `KurtosisContext.GetUserServiceLogs` method
- Added `GetUserServiceLogs` in Kurtosis engine gateway
- Implemented IP address references for services
- Added the `defaultTcpLogsCollectorPortNum` with `9713` value in `EngineManager`
- Added the `LogsCollectorAvailabilityChecker` interface

### Changes
- Add back old enclave continuity test
- Updated the `FluentbitAvailabilityChecker` constructor now it also receives the IP address as a parameter instead of using `localhost`
- Published the `FluentbitAvailabilityChecker` constructor for using it during starting modules and user services
- Refactored `service logs` Kurtosis CLI command in order to get the user service logs from the `logs database` (implemented in Docker cluster so far)

# 0.50.2
### Fixes
- Fixes how the push cli artifacts & publish engine runs by generating kurtosis_version before hand

# 0.50.1
### Fixes
- Fix generate scripts to take passed version on release

# 0.50.0
### Features
- Created new engine's endpoint `GetUserServiceLogs` for consuming user service container logs from the logs database server
- Added `LogsDatabaseClient` interface for defining the behaviour for consuming logs from the centralized logs database
- Added `LokiLogsDatabaseClient` which implements `LogsDatabaseClient` for consuming logs from a Loki's server
- Added `KurtosisBackendLogsClient` which implements `LogsDatabaseClient` for consuming user service container logs using `KurtosisBackend`
- Created the `LogsDatabase` object in `container-engine-lib`
- Created the `LogsCollector` object in `container-engine-lib`
- Added `LogsDatabase` CRUD methods in `Docker` Kurtosis backend
- Added `LogsCollector` CRUD methods in `Docker` Kurtosis backend
- Added `ServiceNetwork` (interface), `DefaultServiceNetwork` and `MockServiceNetwork`

### Breaking Changes
- Updated `CreateEngine` method in `container-engine-lib`, removed the `logsCollectorHttpPortNumber` parameter
    - Users will need to update all the `CreateEngine` calls removing this parameter
- Updated `NewEngineServerArgs`,  `LaunchWithDefaultVersion` and `LaunchWithCustomVersion` methods in `engine_server_launcher` removed the `logsCollectorHttpPortNumber` parameter
    - Users will need to update these method calls removing this parameter
  
### Changes
- Untied the logs components containers and volumes creation and removal from the engine's crud in `container-engine-lib`
- Made some changes to the implementation of the module manager based on some PR comments by Kevin

### Features
- Implement Startosis add_service image pull validation
- Startosis scripts can now be run from the CLI: `kurtosis startosis exec path/to/script/file --enclave-id <ENCLAVE_ID>`
- Implemented Startosis load method to load from Github repositories

### Fixes
- Fix IP address placeholder injected by default in Startosis instructions. It used to be empty, which is invalid now
it is set to `KURTOSIS_IP_ADDR_PLACEHOLDER`
- Fix enclave inspect CLI command error when there are additional port bindings
- Fix a stale message the run-all-test-against-latest-code script
- Fix bug that creates database while running local unit tests
- Manually truncate string instead of using `k8s.io/utils/strings`

### Removals
- Removes version constants within launchers and cli in favor of centralized generated version constant
- Removes remote-docker-setup from the `build_cli` job in Circle

# 0.49.9
### Features
- Implement Startosis add_service method
- Enable linter on Startosis codebase

# 0.49.8
### Changes
- Added a linter
- Made changes based on the linters output
- Made the `discord` command a LowLevelKurtosisCommand instead of an EngineConsumingKurtosisCommand

### Features
- API container now saves free IPs on a local database

### Fixes
- Fix go.mod for commons & cli to reflect monorepo and replaced imports with write package name
- Move linter core/server linter config to within core/server

# 0.49.7
### Features
- Implement skeleton for the Startosis engine

### Fixes
- Fixed a message that referred to an old repo.

### Changes
- Added `cli` to the monorepo

# 0.49.6
### Fixes
- Fixed a bug where engine launcher would try to launch older docker image `kurtosistech/kurtosis-engine-server`.

# 0.49.5
### Changes
- Added `kurtosis-engine-server` to the monorepo
- Merged the `kurtosis-engine-sdk` & `kurtosis-core-sdk`

### Removals
- Remove unused variables from Docker Kurtosis backend

# 0.49.4
### Fixes
- Fix historical changelog for `kurtosis-core`
- Don't check for grpc proxy to be available

# 0.49.3
### Fixes
- Fix typescript package releases

# 0.49.2
### Removals
- Remove envoy proxy from docker image. No envoy proxy is being run anymore, effectively removing HTTP1.

### Changes
- Added `kurtosis-core` to the monorepo

### Fixes
- Fixed circle to not docs check on merge

# 0.49.1
### Fixes
- Attempting to fix the release version
### Changes
- Added container-engine-lib

# 0.49.0
### Changes
- This version is a dummy version to set the minimum. We pick a version greater than the current version of the CLI (0.29.1). 
