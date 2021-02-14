_For details about Kurtosis' versioning scheme, as well as how to upgrade, see [the versioning & upgrading page](./versioning-and-upgrading.md)_

_This changelog is in [KeepAChangelog format](https://keepachangelog.com/en/1.0.0/)_

# TBD
### Changed

# 1.8.1
### Fixed
* Issue where unsigned int was being validated as an int in timeout API

# 1.8.0
### Changed
* Modified API container API to control test setup and execution timeouts in Kurtosis Core instead of kurtosis libs

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
