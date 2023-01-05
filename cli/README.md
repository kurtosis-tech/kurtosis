Kurtosis CLI
============
This repo contains:
* The `kurtosis` CLI
* An internal testsuite to verify that the CLI (and Kurtosis) works

Developing
----------
* Run `cli/scripts/build.sh` to build the CLI into a binary
* Run `cli/scripts/launch-cli.sh` to run arbitrary CLI commands with the locally-built binary
* Run `internal_testsuites/scripts/build.sh` to build and run all the tests in all the supported languages.
* Run `internal_testsuites/golang/scripts/build.sh` to run only `golang` tests. Replace `golang` with `typescript` to run typescript tests.
* Run `internal_testsuites/golang/scripts/build.sh minikube` to build golang test suites against Kubernetes. Replace `golang` with `typescript` to run typescript tests against kubernetes. Note use `./scripts/run-all-tests-against-latest-code.sh minikube` if you want to run all tests and let the script handle Minikube setup for you.

Launching the built `cli` before running the tests is recommended as it pulls the latest `kurtosis-engine` if you need one.

Developers should be able to run and debug tests within Goland in both `typescript` and `golang`. Just click the play button on the single test that needs to be run.

Debugging User Issues
---------------------
### The CLI's not working and there's not enough info to figure out why
The CLI has its own log level (separate from the engine & core). Set the `--cli-log-level` flag to `debug` or `trace` to see more info about what the CLI is doing (can be set on any command).

### Tab completion isn't working
Have the user run the following command, so that all the logs during completion get logged:

```
export BASH_COMP_DEBUG_FILE="/tmp/completion-debugging.log"
```

Cobra also ships with an invisible `__complete` command that will allow you to test various different scenarios like so (note that there needs to be an extra `""` at the point where the user is hitting tab!):

```
kurtosis __complete enclave inspect ""
```
