# TBD

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
