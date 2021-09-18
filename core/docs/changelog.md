# TBD

# 1.19.14
### Changes
* Changed the name of the CLI's published Homebrew formula to `kurtosis` (was `cli`) so that users can do `brew upgrade kurtosis`
* Required the subcommand `sandbox` to be passed in to start a sandbox enclave, to make room for extra commands

### Fixes
* Fix `launch-interactive.sh` for Cobra arg-parsing

# 1.19.13
### Features
* The CLI makes a best-effort attempt to pull the latest version of the API container & Javascript REPL images on each run

### Fixes
* Upgrade to `container-engine-lib` 0.2.9, which fixes the issue with host ports not getting bound

# 1.19.12
### Features
* Add a global unique identifier for services `ServiceGUID` to avoid docker containers collisions and to match docker container name with services folders in enclave data volume
* Add a global unique identifier for lambdas `LambdaGUID` to avoid docker containers collisions and to match docker container name with lambdas folders in enclave data volume
* Disconnect service container from the network when a service is removed with `ServiceNetworkImpl.RemoveService()` method

# 1.19.11
### Changes
* Upgrade to `container-engine-lib` 0.2.7, which has even more logging to track down the empty container ID issue

# 1.19.10
### Fixes
* Fixed LambdaStore not getting passed a `DockerManager`, which led to it segfaulting when it would go to tear down Lambdas upon `LambdaStore.Destroy`

### Changes
* Changed the grace time that an API container has to kill all the services it's managing from 30 seconds to 3 minutes
* When destroying a `ServiceNetworkImpl`, only give the containers 1ms to stop (because we're destroying the network - no need to do so gracefully)
* Upgrade to `container-engine-lib` 0.2.6, which has extra debugging to track down an issue with container ID getting set to emptystring

# 1.19.9
### Fixes
* Upgraded to container-engine-lib 0.2.5, which fixes a bug where not specifying an image tag (which should default to `latest`) wouldn't actually pull the image if it didn't exist locally

# 1.19.8
### Features
* Also push `latest` tag versions of the API container, initializer container, and Javascript REPL image so that the CLI can consume them

### Changes
* Don't push the Go internal testsuite image to Dockerhub

# 1.19.7
### Fixes
* Correct the Homebrew tap's repo name
* Fixed the `kurtosis-core` README getting published with the binaries
* Fix issue with Homebrew formula not having the right binary install instruction

# 1.19.6
### Features
* First attempt at adding Kurtosis CLI installation via Homebrew tap

# 1.19.5
### Fixes
* Don't publish source code to Github as a release
* Publish APKs, DEBs, etc. to the Github release

# 1.19.4
### Fixes
* Upgraded to latest `container-engine-lib`, which fixes an error that would be thrown when an `EXPOSE` directive was declared in the Dockerfile
* Throw an error if the `DockerManager` returns a host port binding map with nil objects (which should never happen)

### Features
* Push the CLI binary up to Github as a Github release

# 1.19.3
### Changes
* Iterate testsuite's tests in alphabetical order

# 1.19.2
### Features
* Add `GetServices` endpoint to get a set of running service IDs
* Add `GetLambdas` endpoint to get a set of running Kurtosis Lambda IDs

# 1.19.1
### Fixes
* Fix broken artifacts-publishing job in CI

# 1.19.0
### Changes
* Prep internal testsuites for having multiple internal testsuites, one per language
* Switch to using `container-engine-lib` for `DockerManager`
* Switch to using `kurtosis-core-launcher-lib`
* Split the `WaitForEndpointAvailability` api container function to `WaitForEndpointAvailabilityHttpGet` and `WaitForEndpointAvailabilityHttpPost`

### Breaking Changes
* Split the `WaitForEndpointAvailability` api container function to `WaitForHttpGetEndpointAvailability` and `WaitForHttpPostEndpointAvailability`
  * Users should replace their `WaitForEndpointAvailability` calls with `WaitForHttpGetEndpointAvailability` or `WaitForHttpPostEndpointAvailability` depending on the endpoint used to check availability

### Removals
* Removed the API container `docker_api` package
* Removed the `ApiContainerLauncher` class here, in favor of the one from `kurtosis-core-launcher-lib`

# 1.18.8
### Changes
* Switched to using `goreleaser` for building our binaries & Docker images

# 1.18.7
### Features
* Add alias to user services' docker container

# 1.18.6
### Fixes
* Fixed bug in image-publishing

# 1.18.5
### Changes
* The interactive CLI now requires an API container image version

### Features
* Build the interactive CLI & Javascript REPL image with `build-and-run.sh`
* Split `build_and_run.sh` into two scripts: `build.sh` and `run-internal-testsuite.sh`
* Added a `launch-interactive.sh` script for running Kurtosis Interactive
* Publish the Javascript REPL image to `kurtosistech/javascript-interactive-repl`

### Fixes
* Fix bug with not checking enclave creation error in the interactive CLI
* The CircleCI artifact-publishing now uses the same constants as all the other scripts in the `scripts` directory
* The `launch-interactive` script will now appropriately use whichever version of the Javascript repl that you're working on

# 1.18.4
### Changes
* Update `Kurtosis Client` to version 0.13.0 which adds a new argument in `kurtosis_core_rpc_api_bindings.WaitForEndpointAvailabilityArgs` to specify the http request body used in the http call to service's availability endpoint.
* Switch to using check-docs orb
* Use the updated `minimal-grpc-server` Golang module, which is in a subdirectory

### Removals
* Removed docs that have been ported to the main docs repo

# 1.18.3
### Changes
* Update `Kurtosis Client` to version 0.12.0 which adds a new argument in `kurtosis_core_rpc_api_bindings.WaitForEndpointAvailabilityArgs` to specify the http method used in the http call to service's availability endpoint. The allowed values are GET or POST
* Update internal testsuite tests adding the new argument `httpMethod` in every `WaitForEndpointAvailability` call

# 1.18.2
### Changes
* Switched some less-important log levels (e.g. "Starting API container..." from INFO -> DEBUG)

### Features
* Added a watermark with support information that displays on every run
* When we need to get a token for the user, also give them the signup link in case they don't have an account
* Add a test for our user support URLs to verify they're all valid URLs

### Fixes
* Wait for the API container to start up before we return the enclave to the user so there's no risk of dialling an API container and getting a connection refused
* Check error when creating the request object that will get sent to Auth0 to get the device authorization

# 1.18.1
### Fixes
* Correct links now that `kurtosis-libs` is renamed to `kurtosis-testsuite-starter-pack`
* There are more disallowed IP ranges than just the multicast addresses (see [this Wikipedia article](https://en.wikipedia.org/wiki/IPv4#Special-use_addresses)), so prevent the Docker network allocator from choosing those

### Features
* The API container will now shut down all the containers in its network as it shuts down, which is a step towards enclaves being independent of the testing framework
* Created `EnclaveManager`, to start & stop enclaves independent of the testing framework
* Added a CLI for starting a Kurtosis enclave with an attached Javascript REPL

# 1.18.0
### Fixes
* Updated copyright notice to 2021, with entity as Kurtosis Technologies Inc.
* Standardized enclave naming convention for the testing framework to `KTTYYYY-MM-DDTHH.MM.SS-RANDOMNUM_TESTNAME`, where:
    * `KTT` is a prefix indicating "Kurtosis testing"
    * `YYYY-MM-DDTHH.MM.SS` is the timestamp of when the testsuite execution was launched
    * `RANDOMNUM` is a random salt to ensure that two testsuites run at exactly the same second don't collide
    * `TESTNAME` is the name of the test running inside the enclave

### Features
* When running in debug mode, let Docker handle host-port binding
    * This allows multiple versions of Kurtosis to be running in debug mode at the same time

### Removals
* Removed OptionalHostPortBindingSupplier, which is no longer needed

### Breaking Changes
* Add explicit copyright notice to all files (including `kurtosis.sh`)

# 1.17.1
### Features
* Allow multiple instances of Kurtosis to run at the same time!
* Upgraded to [testsuite API lib 0.4.0](https://github.com/kurtosis-tech/kurtosis-testsuite-api-lib/blob/develop/docs/changelog.md#040)

### Fixes
* Add an extra guard to make sure that DockerNetworkAllocator can't be instantiated without `rand.Seed` being called
* Skip [multicast addresses](https://en.wikipedia.org/wiki/Multicast_address) when choosing network IPs

# 1.17.0
### Features
* Added a test for the static file cache
* There is no longer a single "suite execution volume" across multiple tests; instead, each test gets its own "enclave data volume"
    * This is one of the necessary steps to get to Kurtosis Interactive
* Use the testsuite API that reads environment variables directly (so that users don't need to ever touch their Dockerfile)

### Changes
* Upgraded to testsuite API lib 0.3.0
* Backed the `FilesArtifactCache` and `StaticFilesCache` by the same object, for better code quality

### Fixes
* Stopped the scary `use of closed network connection` error from appearing with the log streamer, as it's expected

### Breaking Changes
* `kurtosis.sh` no longer creates a suite execution volume, and the initializer container no longer accepts a param for it
    * Users should remove the `SUITE_EXECUTION_VOLUME` flag/parameter from their Dockerfile

# 1.16.6
### Features
* Moved the Kurtosis-internal testsuite from Kurtosis Libs into here, to break the circular dependency that used to exist between the two repos

# 1.16.5
### Changes
* Depend on `kurtosis-testsuite-api-lib`, rather than `kurtosis-libs`, to get testsuite API bindings

# 1.16.4
### Fixes
* Fixes a very occasional failure with exec commands due to a race condition in the Docker engine

# 1.16.3
### Features
* Added Kurtosis Lambda support!

### Changes
* Upgraded to Kurtosis Client 0.9.0
* Inserted an extra `user-service` element to user service container names, for easier identification
* Switched the API container's `main.go` to read environment variables directly, rather than taking in Dockerfile flags
    * This means that we won't need to change the Dockerfile if we add new parameters!
* Added CircleCI step in check_code to check for any running docker containers after Kurtosis testsuite builds and runs 

### Fixes
* Added check to account for error when calling Destroy method inside api_container/main.go

# 1.16.2
### Changes
* Added Destroy method that would tear down the docker side containers

### Fixes
* Fix the bug where the side containers weren't torn down properly

# 1.16.1
### Changes
* Upgraded to Kurt Client 0.5.0
* Upgraded example Go testsuite being used in `build-and-run.sh` to v1.28.0 (was v1.24.2)
* `release.sh` is now a simple wrapper around the devtools `release-repo.sh` script

### Fixes
* Fixed bug where a testsuite with no static files would trip an overly-aggressive validation check

# 1.16.0
### Changes
* Upgraded to Kurtosis Client v0.4.0
* Implemented the `LoadStaticFiles` endpoint

### Breaking Changes
* Testsuites must now provide used static files in testsuite metadata

# 1.15.7
### Changes
* Upgraded to Kurtosis Client v0.3.0

### Features
* Add a new method `GetServiceInfo` in API container which can be used to get relevant information about a service running in the network

# 1.15.6
### Changes
* Upgraded to using Kurtosis Client v0.2.2, with the `ExecuteBulkCommands` endpoint

### Features
* Implemented the `ExecuteBulkCommands` endpoint in the API container's API for running multiple API container endpoint commands at once

### Fixes
* Fixed bug where `ServiceNetwork.GetServiceIP` didn't use the mutex

# 1.15.5
### Features
* Added a new method `WaitForEndpointAvailability` to the API container which can be used to wait for a service's HTTP endpoint to come up
* Added a new method `GetServiceIP` in service network which returns an IP Address by Service ID.

# 1.15.4
### Fixes
* Fixed a bug where Docker's `StdCopy` was being used to copy logs from testsuite tempfile to STDOUT

# 1.15.3
### Features
* Add custom params log to show users that Kurtosis has loaded this configuration

### Changes
* Change the API container to depend on `kurtosis-client`, rather than duplicating the `.proto` file in here
* Change the API container to depend on `kurtosis-libs`, rather than duplicating the testsuite `.proto` file in here
* Removed the `regenerate-protobuf-bindings.sh`, as it's no longer necessary
* Upgraded Kurtosis Client version to 0.2.0 (was 0.1.1)

# 1.15.2
### Fixes
* Fixed an issue with the LogStreamer where StdCopy is a blocking method which could halt all of Kurtosis if not dealt with appropriately

# 1.15.1
### Changes
* Add a new user-friendly log message when setup or run timeout is exceeded during a test starting process

### Fixes
* Added `#!/usr/bin/env bash` shebang to the start of all shell scripts, to solve the shell incompatibility issues we've been seeing

# 1.15.0
### Changes
* Renamed --test-suite-log-level flag to kurtosis.sh to be --suite-log-level

### Breaking Changes
* The flag to set the testsuite's log level has been renamed from `--test-suite-log-level` -> `--suite-log-level`

# 1.14.4
* Add a release script to automate the release process

# 1.14.3
### Changes
* Made the docs for customizing a testsuite more explicit
* Make `build-and-run-core.sh` more explicit about what it's building

### Fixes
* Switch back to productized version of Kurtosis Libs (v1.24.2) for the `build-and-run.sh` script
* Updated `DockerContainerInitializer` -> `ContainerConfigFactory` in diagram in testsuite customization docs

# 1.14.2
### Fixes
* Fixed an issue where `kurtosis.sh` would break on some versions of Zsh when trying to uppercase the execution instance UUID
* Fixed an occasional issue where the initializer would try to connect to the testsuite container before it was up (resulting in a "connection refused" and a failed test) by adding an `IsAvailable` endpoint to the testsuite that the initializer will poll before running test setup

# 1.14.1
### Fixes
* Fixed an occasional failure by Docker to retrieve the initializer container's ID
* Fixed an issue where `user_service_launcher` wasn't setting container used ports correctly

# 1.14.0
### Changes
* Significantly refactored the project to invert the relationship between Core & a testsuite container: rather than the testsuite container being linear code that registers itself with the API container, the testsuite container now runs a server and the initializer container calls the testsuite container directly
    * This sets the stage for Kurtosis Modules, where modules run servers that receive calls just like a library
    * This necessitated the initializer container now being mounted inside the test subnetwork
* Significantly simplified the API container by removing all notion of test-tracking, e.g.:
    * Removed the `suite_registration` API entirely
    * Removed the concept of "modes" from the API container
    * Pushed the burden for timeout-tracking to the initializer itself, so the API container is a fairly simple proxy for Docker itself and doesn't do any test lifecycle tracking

### Breaking Changes
* Completely reworked the API container Protobuf API
* Added a new testsuite container API

# 1.13.1
NOTE: Empty release to get `master` back on track after the reverting

# 1.13.0
NOTE: Contains the changes in 1.12.1, which were incorrectly released as a patch version bump

# 1.12.2
NOTE: Undoes 1.12.1, which should have been released as a minor version bump

# 1.12.1
### Features
* Containers are given descriptive names, rather than using Docker autogenerated ones

### Changes
* Added a big comment to the top of `build-and-run-core.sh` warning users not to modify it, as it'll be overridden by Kurtosis upgrades
* Added copyright notices to the top of `kurtosis.sh` and `build-and-run-core.sh`

# 1.12.0
### Features
* Added a `--debug` flag to `kurtosis.sh` that will:
    1. Set parallelism to 1, so test logs are streamed in realtime
    1. Bind a port on the user's local machine to a port on the testsuite container, so a debugger can be run
    1. Bind every port that a service uses to a port on the user's local machine, so they can make requests to the services for debugging

### Fixes
* Fixed `FreeHostPortBindingSupplier` using the `tcp` protocol to check ports, regardless of what protocol it was configured with
* Fixed `docker build` on initializer & API images not writing output to file, since Docker Buildkit (which is now enabled by default) writes everything to STDERR

### Breaking Changes
* The API container's `StartService` call now returns a `StartServiceResponse` object containing ports that were bound on the local machine (if any) rather than an empty object

# 1.11.1
### Fixes
* Fixed issue with `kurtosis.sh` in `zsh` throwing `POSITIONAL[@] unbound`
* Upgraded example Go testsuite being used for CI checks to `1.20.0`
* Fixed broken links in `testsuite-customization` documentation
* Fixed issue with free host port checker not accurately detecting free host ports

### Changes
* Removed the `--debugger-port` arg, which was exposed in `kurtosis.sh` but didn't actually do anything
* Upped the time for testsuites to register themselves from 10s to 20s, to give users more time to connect to the testsuite when running inside a debugger

# 1.11.0
### Features
* Allowed files to be generated for a service at any point during the service's lifecycle

### Changes
* Renamed `test_execution_timeout_in_seconds` to `test_run_timeout_in_seconds` on the testsuite metadata serialization call, to better reflect the actual value purpose

### Breaking Changes
* The `RegisterService` API container function no longer takes in a set of files to generate, nor does it return the relative filepaths of generated files
* Added a new API container function, `GenerateFiles`, that generates files inside the suite execution volume of a specified service
* Renamed the `test_execution_timeout_in_seconds` arg to `SerializeSuiteMetadata` API container function to `test_run_timeout_in_seconds`

# 1.10.6
### Changes
* Refactored `PrintContainerLogs` function out of `banner_printer` module (where it didn't belong)

### Fixes
* Fixed an issue where a testsuite that hangs forever could hang the initializer because the log streamer's `io.Copy` operation is blocking

# 1.10.5
### Features
* Whenenever a single test is running (either because one test is specified or parallelism == 1), test logs will stream in realtime

### Changes
* Removed now-unused `index.md` and `images/horizontal-logo.png` from the `docs` folder (has been superseded by https://github.com/kurtosis-tech/kurtosis-tech.github.io )
* Pushed the logline "Attempting to remove Docker network with ID ...." down to "debug" level (was "info")

### Fixes
* Actually abort tests & shut everything down when the user presses Ctrl-C
* Fixed issue where hung calls to the API container (e.g. a long-running Docker exec command) could prevent the API container from shutting down
* Normalized test name and testsuite log banner widths

# 1.10.4
### Fixes
* Broken links since we combined this repo's docs with Kurtosis Libs

# 1.10.3
### Changes
* Switched the log messages pertaining to test params and work queues to debug, as they provide no useful information for the lay user
* Removed `docs.kurtosistech.com` CNAME record, in preparation for an org-wide Github Pages docs account

# 1.10.2
### Features
* `build-and-run-core.sh` will now set a Docker build arg called `BUILD_TIMESTAMP` that can be used to intentionally bust Docker's cache in cases where it's incorrectly caching steps (see also: https://stackoverflow.com/questions/31782220/how-can-i-prevent-a-dockerfile-instruction-from-being-cached )

# 1.10.1
### Fixes
* Don't throw an error when adding the same artifact to the artifact cache multiple times

# 1.10.0
### Features
* Added Docker exec log output to protobuf response for ExecCommand, with a limit of 10MB for size of logs returned

# 1.9.1
### Fixes
* If the API container's gRPC server doesn't gracefully stop after 10s, hard-stop it to prevent hung calls to the server from hanging the API container exit (e.g. AddService with a super-huge Docker image)

# 1.9.0
### Features
* Added the ability to override a Docker image's `ENTRYPOINT` directive via the new `entrypoint_args` field to the API container's `StartServiceArgs` object
 
### Breaking Changes
* Renamed the `start_cmd_args` field on the API container's `StartServiceArgs` Protobuf object to `cmd_args`

# 1.8.2
_NOTE: Changelog entries from this point on will abandon the KeepAChangelog format, as it has done a poor job of highlighting the truly important things - features, fixes, and breaking changes_

### Features
* Add a new endpoint to the Kurtosis API container, `ExecCommand`, to provide the ability for testsuite authors to run commands against running containers via `docker exec`
    * NOTE: As currently written, this is a synchronous operation - no other changes to the network will be possible while an `ExecCommand` is running!

### Fixes
* Don't give any grace time for containers to stop when tearing down a test network because we know we're not going to use those services again (since we're tearing down the entire test network)


# 1.8.1
### Changed
* Update the name of the Kurtosis Go example testsuite image (now `kurtosis-golang-example` rather than `kurtosis-go-example`)
* Use a pinned version of `kurtosis-go-example` when doing the "make sure testsuites still work" sanity check, so that we don't have to build a `develop` version of the Kurt Libs testsuites
* Added extra monitoring inside the API container such that if a testsuite exits during the test setup phase (which should never happen), the API container will exit with an error immediately (rather than the user needing to wait for the test setup timeout)

### Fixed
* Error with `TestsuiteLauncher` printing log messages to the standard logger when it should be printing them to the test-specific logger

# 1.8.0
### Changed
* * Modified API container API to control test setup and execution timeouts in Kurtosis Core instead of kurtosis libs

# 1.7.4
### Changed
* Swapped all `kurtosis-go` link references to point to `kurtosis-libs`

# 1.7.3
### Fixed
* Issue where `kurtosis.sh` errors with unset variable when `--help` flag is passed in

# 1.7.2
### Added
* New `testsuite-customization.md` and corresponding docs page, to contain explicit instructions on customizing testsuites
* A link `testsuite-customization.md` to the bottom of every other docs page
* `build-and-run-core.sh` under the `testsuite_scripts` directory
* Publishing of `build-and-run-core.sh` to the public-access S3 bucket via the CircleCI config

### Changed
* All "quickstart" links to `https://github.com/kurtosis-tech/kurtosis-libs/tree/master#testsuite-quickstart`
* All docs to reflect that the script is now called `build-and-run.sh` (hyphens), rather than `build_and_run.sh` (underscores)
* "Versioning & Upgrading" docs to reflect the new world with `kurtosis.sh` and `build-and-run-core.sh`

### Removed
* `quickstart.md` docs page in favor of pointing to [the Kurtosis libs quickstart instructions](https://github.com/kurtosis-tech/kurtosis-libs/tree/master#testsuite-quickstart)

# 1.7.1
* Update docs to reflect the changes that came with v1.7.0
* Remove "Testsuite Details" doc (which contained a bunch of redundant information) in favor of "Building & Running" (which now distills the unique information that "Testsuite Details" used to contain)
* Pull down latest version of Go suite, so we're not using stale versions when running
* Remove the `isPortFree` check in `FreeHostPortProvider` because it doesn't actually do what we thought - it runs on the initializer, so `isPortFree` had actually been checking if a port was free on the _initializer_ rather than the host
* Color `ERRORED`/`PASSED`/`FAILED` with green and red colors
* Added a "Further Reading" section at the bottom of eaach doc page

# 1.7.0
* Refactor API container's API to be defined via Protobuf
* Split the adding of services into two steps, which removes the need for an "IP placeholder" in the start command:
    1. Register the service and get back the IP and filepaths of generated files
    2. Start the service container
* Modified API container to do both test execution AND suite metadata-printing, so that the API container handles as much logic as possible (and the Kurtosis libraries, written in various languages, handle as little as possible)
* Modified the contract between Kurtosis Core and the testsuite, such that the testsuite only takes in four Docker environment variables now (notably, all the user-custom params are now passed in via `CUSTOM_PARAMS_JSON` so that they don't need to modify their Dockerfile to pass in more params)
    * `DEBUGGER_PORT`
    * `KURTOSIS_API_SOCKET`
    * `LOG_LEVEL`
    * `CUSTOM_PARAMS_JSON`
* To match the new `CUSTOM_PARAMS_JSON`, the `--custom-env-vars` flag to `kurtosis.sh`/`build_and_run` has been replaced with `--custom-params`

# 1.6.5
* Refactor ServiceNetwork into several smaller components, and add tests for them
* Switch API container to new mode-based operation, in preparation for multiple language clients
* Make the "Supported Languages" docs page send users to the master branch of language client repos
* Fix `build_and_run` breaking on empty `"${@}"` variable for Zsh/old Bash users
* Added explicit quickstart instruction to check out `master` on the client language repo

# 1.6.4
* Modify CI to fail the build when `ERRO` shows up, to catch bugs that may not present in the exit code
* When a container using an IP is destroyed, release it back into the free IP address tracker's pool
* When network partitioning is enabled, double the allocated test network width to make room for the sidecar containers

# 1.6.3
* Generate kurtosis.sh to always try and pull the latest version of the API & initializer containers

# 1.6.2
* Prevent Kurtosis from running when the user is restricted to the free trial and has too many tests in their testsuite

# 1.6.1
* Switch to using the background context for pulling test suite container logs, so that we get the logs regardless of context cancellation
* Use `KillContainer`, rather than `StopContainer`, on network remove (no point waiting for graceful shutdown)
* Fix TODO in "Advanced Usage" docs

# 1.6.0
* Allow users to mount external files-containing artifacts into Kurtosis services

# 1.5.1
* Clarify network partitioning docs

# 1.5.0
* Add .dockerignore file, and a check in `build.sh` to ensure it exists
* Give users the ability to partition their testnets
* Fixed bug where the timeout being passed in to the `RemoveService` call wasn't being used
* Added a suite of tests for `PartitionTopology`
* Add a `ReleaseIpAddr` method to `FreeIpAddrTracker`
* Resolve the race condition that was occurring when a node was started in a partition, where it wouldn't be sectioned off from other nodes until AFTER its start
* Add docs on how to turn on network partitioning for a test
* Resolve the brief race condition that could happen when updating iptables, in between flushing the iptables contents and adding the new rules
* Tighten up error-checking when deserializing test suite metadata
* Implement network partitioning PR fixes

# 1.4.5
* Use `alpine` base for the API image & initializer image rather than the `docker` Docker-in-Docker image (which we thought we needed at the start, but don't actually); this saves downloading 530 MB _per CI build_, and so should significantly speed up CI times

# 1.4.4
* Correcting bug with `build_and_run` Docker tags

# 1.4.3
* Trying to fix the CirlceCI config to publish the wrapper script & Docker images

# 1.4.2
* Debugging CircleCI config

# 1.4.1
* Empty commit to debug why Circle suddenly isn't building tags

# 1.4.0
* Add Go code to generate a `kurtosis.sh` wrapper script to call Kurtosis, which:
    * Has proper flag arguments (which means proper argument-checking, and no more `--env ENVVAR="some-env-value"`!)
    * Contains the Kurtosis version embedded inside, so upgrading Kurtosis is now as simple as upgrading the wrapper script
* Fixed the bug where whitespace couldn't be used in the `CUSTOM_ENV_VARS_JSON` variable
    * Whitespaces and newlines can be happily passed in to the wrapper script's `--custom-env-vars` flag now!
* Add CircleCI logic to upload `kurtosis.sh` versions to the `wrapper-script` folder in our public-access S3 bucket
* Updated docs to reflect the use of `kurtosis.sh`

# 1.3.0
* Default testsuite loglevel to `info` (was `debug`)
* Running testsuites can now be remote-debugged by updating the `Dockerfile` to run a debugger that listens on the `DEBUGGER_PORT` Docker environment variable; this port will then get exposed as an IP:port binding on the user's local machine for debugger attachment

# 1.2.4
* Print the names of the tests that will be run before running any tests
* Fix bug with test suite results not ordered by test name alphabetically
* Add more explanation to hard test timeout error, that this is often caused by testnet setup taking too long
* Switch Docker volume format from `SUITEIMAGE_TAG_UNIXTIME` to `YYYY-MM-DDTHH.MM.SS_SUITEIMAGE_TAG` so it's better sorted in `docker volume ls` output
* Prefix Docker networks with `YYYY-MM-DDTHH.MM.SS` so it sorts nicely on `docker network ls` output

# 1.2.3
* Only run the `docker_publish_images` CI job on `X.Y.Z` tags (used to be `master` and `X.Y.Z` tags, with the `master` one failing)

# 1.2.2
* Switch to Midnight theme for docs instead of Hacker
* Migrate CI check from kurtosis-docs for verifying all links work
* Move this changelog file from `CHANGELOG.md` to `docs/changelog.md` for easier client consumption
* Don't run Go code CI job when only docs have changed
* Switch to `X.Y` tagging scheme, from `X.Y.Z`
* Only build Docker images for release `X.Y` tags (no need to build `develop` any time a PR merges)
* Remove `PARALLELISM=2` flag from CI build, since we now have 3 tests and there isn't a clear reason for gating it given we're spinning up many Docker containers

# 1.2.1
* Add a more explanatory help message to `build_and_run`
* Correct `build_and_run.sh` to use the example microservices for kurtosis-go 1.3.0 compatibility
* Add `docs` directory for publishing user-friendly docs (including CHANGELOG!)

# 1.2.0
* Changed Kurtosis core to attempt to print the test suite log in all cases (not just success and `NoTestSuiteRegisteredExitCode`)
* Add multiple tests to the `AccessController` so that any future changes don't require manual testing to ensure correctness
* Removed TODOs related to IPv6 and non-TCP ports
* Support UDP ports (**NOTE:** this is an API break for the Kurtosis API container, as ports are now specified with `string` rather than `int`!)

# 1.1.0
* Small comment/doc changes from weekly review
* Easy fix to sort `api_container_env_vars` alphabetically
* Remove some now-unneeded TODOs
* Fix the `build_and_run.sh` script to use the proper way to pass in Docker args
* Disallow test names with our test name delimiter: `,`
* Create an access controller with basic license auth
* Connect access controller to auth0 device authorization flow
* Implement machine-to-machine authorization flow for CI jobs
* Bind-mount Kurtosis home directory into the initializer image
* Drop default parallelism to `2` so we don't overwhelm slow machines (and users with fast machines can always bump it up)
* Don't run `validate` workflow on `develop` and `master` branches (because it should already be done before merging any PRs in)
* Exit with error code of 1 when `build_and_run.sh` receives no args
* Make `build_and_run.sh` also print the logfiles of the build threads it launches in parallel, so the user can follow along
* Check token validity and expiration
* Renamed all command-line flags to the initializer's `main.go` to be `UPPER_SNAKE_CASE` to be the same name as the corresponding environment variable passed in by Docker, which allows for a helptext that makes sense
* Added `SHOW_HELP` flag to Kurtosis initializer
* Switched default Kurtosis loglevel to `info`
* Pull Docker logs directly from the container, removing the need for the `LOG_FILEPATH` variable for testsuites
* Fixed bug where the initializer wouldn't attempt to pull a new token if the token were beyond the grace period
* Switch to using `permissions` claim rather than `scope` now that RBAC is enabled

# 1.0.3
* Fix bug within CircleCI config file

# 1.0.2
* Fix bug with tagging `X.Y.Z` Docker images

# 1.0.1
* Modified CircleCI config to tag Docker images with tag names `X.Y.Z` as well as `develop` and `master`

# 1.0.0
* Add a tutorial explaining what Kurtosis does at the Docker level
* Kill TODOs in "Debugging Failed Tests" tutorial
* Build a v0 of Docker container containing the Kurtosis API 
* Add registration endpoint to the API container
* Fix bugs with registration endpoint in API container
* Upgrade new initializer to actually run a test suite!
* Print rudimentary version of testsuite container logs
* Refactor the new intializer's `main` method, which had become 550 lines long, into separate classes
* Run tests in parallel
* Add copyright headers
* Clean up some bugs in DockerManager where `context.Background` was getting used where it shouldn't
* Added test to make sure the IP placeholder string replacement happens as expected
* Actually mount the test volume at the location the user requests in the `AddService` Kurtosis API endpoint
* Pass extra information back from the testsuite container to the initializer (e.g. where to mount the test volume on the test suite container)
* Remove some unnecessary `context.Context` pointer-passing
* Made log levels of Kurtosis & test suite independently configurable
* Switch to using CircleCI for builds
* Made the API image & parallelism configurable
* Remove TODO in run.sh about parameterizing binary name
* Allow configurable, custom Docker environment variables that will be passed as-is to the test suite
* Added `--list` arg to print test names in test suite
* Kill unnecessary `TestSuiteRunner`, `TestExecutorParallelizer`, and `TestExecutor` structs
* Change Circle config file to:
    1. Build images on pushes to `develop` or `master`
    2. Run a build on PR commits
* Modify the machinery to only use a single Docker volume for an entire test suite execution
* Containerize the Docker initializer
* Refactored all the stuff in `scripts` into a single script

# 0.9.0
* Change ConfigurationID to be a string
* Print test output as the tests finish, rather than waiting for all tests to finish to do so
* Gracefully clean up tests when SIGINT, SIGQUIT, or SIGTERM are received
* Tiny bugfix in printing test output as tests finish

# 0.8.0
* Simplify service config definition to a single method
* Add a CI check to make sure changelog is updated each commit
* Use custom types for service and configuration IDs, so that the user doesn't have a ton of `int`s flying around
* Made TestExecutor take in the long list of test params as constructor arguments, rather than in the runTest() method, to simplify the code
* Make setup/teardown buffer configurable on a per-test basis with `GetSetupBuffer` method
* Passing networks by id instead of name inside docker manager
* Added a "Debugging failed tests" tutorial
* Bugfix for broken CI checks that don't verify CHANGELOG is actually modified
* Pass network ID instead of network name to the controller
* Switching FreeIpAddrTracker to pass net.IP objects instead of strings
* Renaming many parameters and variables to represent network IDs instead of names
* Change networks.ServiceID to strings instead of int
* Documenting every single public function & struct for future developers

# 0.7.0
* Allow developers to configure how wide their test networks will be
* Make `TestSuiteRunner.RunTests` take in a set of tests (rather than a list) to more accurately reflect what's happening
* Remove `ServiceSocket`, which is an Ava-specific notion
* Add a step-by-step tutorial for writing a Kurtosis implementation!

# 0.6.0
* Clarified the README with additional information about what happens during Kurtosis exit, normal and abnormal, and how to clean up leftover resources
* Add a test-execution-global timeout, so that a hang during setup won't block Kurtosis indefinitely
* Switch the `panickingLogWriter` for a log writer that merely captures system-level log events during parallel test execution, because it turns out the Docker client uses logrus and will call system-level logging events too
* `DockerManager` no longer stores a Context, and instead takes it in for each of its functions (per Go's recommendation)
* To enable the test timeout use case, try to stop all containers attached to a network before removing it (otherwise removing the network will guaranteed fail)
* Normalize banners in output and make them bigger

# 0.5.0
* Remove return value of `DockerManager.CreateVolume`, which was utterly useless
* Create & tear down a new Docker network per test, to pave the way for parallel tests
* Move FreeIpAddrTracker a little closer to handling IPv6
* Run tests in parallel!
* Print errors directly, rather than rendering them through logrus, to preserve newlines
* Fixed bug where the `TEST RESULTS` section was displaying in nondeterministic order
* Switch to using `nat.Port` object to represent ports to allow for non-TCP ports

# 0.4.0
* remove freeHostPortTracker and all host-container port mappings
* Make tests declare a timeout and mark them as failed if they don't complete in that time
* Explicitly declare which IP will be the gateway IP in managed subnets
* Refactored the big `for` loop inside `TestSuiteRunner.RunTests` into a separate helper function
* Use `defer` to stop the testnet after it's created, so we stop it even in the event of unanticipated panics
* Allow tests to stop network nodes
* Force the user to provide a static service configuration ID when running `ConfigureNetwork` (rather than leaving it autogenerated), so they can reference it later when running their test if they wanted to add a service during the test
* Fix very nasty bug with tests passing when they shouldn't
* Added stacktraces for `TestContext.AssertTrue` by letting the user pass in an `error` that's thrown when the assertion fails

# 0.3.1
* explicitly specify service IDs in network configurations

# 0.3.0
* Stop the node network after the test controller runs
* Rename ServiceFactory(,Config) -> ServiceInitializer(,Core)
* Fix bug with not actually catching panics when running a test
* Fix a bug with the TestController not appropriately catching panics
* Log test result as soon as the test is finished
* Add some extra unit tests
* Implement controller log propagation
* Allow services to declare arbitrary file-based dependencies (necessary for staking)
