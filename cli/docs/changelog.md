# TBD
### Fixes
* Fixed error where `service logs` command is executed with a nonexistent enclave ID or nonexistent GUID just silently exits

### Changes
* Replaced `kurtosistech/example-microservices_datastore` Docker image with `kurtosistech/example-datastore-server` in `golang_internal_testsuite`
* Replaced `kurtosistech/example-microservices_api` Docker image with `kurtosistech/example-api-server` in `golang_internal_testsuite`
* Replaced `kurtosistech/example-microservices_datastore` Docker image with `docker/getting-started` in `bulk_command_execution_test` and `wait_for_endpoint_availability_test`
* Upgraded `datastore army module` Docker image to the latest version `kurtosistech/datastore-army-module:0.1.5` in `module_test` 

# 0.6.1
### Fixes
* Fixed a bug where a testsuite could be reported as passing, even when the tests were failing

# 0.6.0
### Features
* The engine loglevel can be configured with the new `--log-level` flag to `engine start`

### Changes
* Renamed the `repl new` image flag to make more sense
* Update to using the engine server that stores engine/enclave data on the user's local machine

### Breaking Changes
* The `--js-repl-image` flag of `repl new` has been renamed to `--image`, with shorthand `-i`

# 0.5.5
### Fixes
* A failed `module exec` or `sandbox` stops, rather than destroys, the enclave it created

# 0.5.4
### Features
* `enclave new` prints the new enclave's ID
* Information about how to stop or remove the enclave created by `sandbox` is printed after the REPL exits
* Added a `clean` command, to clean up accumulated Kurtosis artifacts
* Added a `repl inspect` command to list installed packages on the REPL

### Fixes
* Use `--image` flag in `kurtosis engine start` command, it was not being used when the engine is being executed
* Fix the returning values order when `DestroyEnclave` method is called in `kurtosis sandbox` command
* Fixed a bug where `engine status` wouldn't check the error value from getting the engine status object
* The Javascript REPL's module installation paths in the Dockerfile are now filled from Go code constants (rather than being hardcoded)

### Changes
* The `sandbox` command no longer destroys the enclave after the REPL exits
* Upgrade to engine server 0.4.7, where API container doesn't shut down the containers after it exits (instead relying on the engine server to do that)

# 0.5.3
### Fixes
* Upgrade to the `goreleaser-ci-image` 0.1.1 to publish a new Homebrew formula with a fix for the `bottle :unneeded` deprecation warning

# 0.5.2
### Features
* Added `enclave stop` and `enclave rm` commands

# 0.5.1
### Features
* Add instructions for users on what to do if no Kurtosis engine is running
* If an engine isn't running, the CLI will try to start one automatically

# 0.5.0
### Changes
* Replaced `EnclaveManager` with `Kurtosis Engine API Libs` which handle all the interactions with the `Kurtosis Engine Server`

### Features
* Add a `version` command to print the CLI's version, with a test
* Added a global `--cli-log-level` flag that controls what level the CLI will log at

### Fixes
* The Kurtosis Client version used by the JS REPL image will now use the `KurtosisApiVersion` constant published by Kurt Client
* Fixed bug where testsuite containers weren't getting any labels

### Breaking Changes
* Interactions with the CLI now require a Kurtosis engine to be running
    * Users should run `kurtosis engine start` to start an engine

# 0.4.3
### Features
* Added documentation in README about how to develop on this repo
* Upgraded to `kurtosis-core` 1.25.2, which contains fixes for `container-engine-lib` 0.7.0 feature that allows binding container ports to specific host machine ports
* Added `engine start` command to the CLI
* Added `engine stop` command to the CLI
* `engine start` waits until the engine is responding to gRPC requests before it reports the engine as up
* Added `engine status` command to the CLI
* Start a Kurtosis engine server in the CI environment

### Fixes
* Clarified the difference between the two types of params in `module exec`
* `engine start` won't start another container if one is already running
* `engine start` waits for gRPC availability before it reports the engine up

# 0.4.2
### Features
* `enclave` commands also show enclave state
* Standardized table-printing logic into a `TablePrinter` object 
* Added a `KeyValuePrinter` for pretty-printing key-value pairs
* `enclave inspect` also prints the enclave ID & state

### Fixes
* `module exec` will attempt to update the module & API container images before running
* Fixed a bug where having a `node_modules` directory in your current directory when starting a REPL will cause the REPL to fail

### Changes
* Upgrade to testsuite-api-lib 0.11.0, which uses Kurt Client 0.19.0 (already handled in v0.4.0 of this repo)
* When running a REPL, your current directory is now mounted at `/local` rather than the present directory

# 0.4.1
### Fixes
* Update the Javascript CLI's `core-api-lib` version to 0.19.0, to match the rest of the code

# 0.4.0
### Changes
* Switched all references to "Lambda" to "module"

### Fixes
* `ModuleTest` (renamed from `LambdaTest`) now uses the ports returned by the Datastore Army module
* Fixed bug in CI where `pipefail` wasn't set which would result in the testsuite-running step passing when it shouldn't

### Breaking Changes
* Renamed the `lambda` command to `module`

# 0.3.4
### Fixes
* Stop attempting to upload APK packages to Gemfury (which can't accept APK packages and throws an error)

# 0.3.3
### Features
* Added a `repl install` command for installing NPM packages to a running REPL container
* `ParsePositionalArgs` (renamed to `ParsePositionalArgsAndRejectEmptyStrings`) now also errors on empty strings

# 0.3.2
### Features
* Added `enclave new` command to create a new enclave
* Added `repl new` command to start a new Javascript REPL
* Added `REPL runner` to reuse the creation and execution of the REPL container
* Print interactive REPLs in `enclave inspect`
* Added `GetEnclave` method in `EnclaveManager` in order to get information of a running enclave
* Upgrade Kurtosis Core Engine Libs to v0.6.0 which adds `Network` type
* Upgrade Kurtosis Core to v1.24.0 which splits `api-container-url` into `api-container-ip` and `api-container-port`

# 0.3.1
### Fixes
* Pinned the default API container version to the same version as in the `go.mod`, so that its version can't silently upgrade under users and break everything

# 0.3.0
### Changes
* Changed the Homebrew/deb/rpm package name to `kurtosis-cli` (was `kurtosis`)

### Breaking Changes
* The CLI is now installed via the `kurtosis-cli` package (for Homebrew, APT, and Yum) rather than just `kurtosis`

# 0.2.1
### Fixes
* Fixed missing `FURY_TOKEN` when publishing

# 0.2.0
### Features
* Ported over the CLI & internal testsuite from `kurtosis-core`

### Breaking Changes
* Changed pretty much everything to add the CLI

# 0.1.0
* Initial commit
