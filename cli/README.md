Kurtosis CLI
============
This repo contains:
* The `kurtosis` CLI
* An internal testsuite to verify that the CLI (and Kurtosis) works

### Developing
* Run `scripts/build.sh` to build the CLI into a binary & testsuite into a Docker image
* Run `scripts/launch-cli.sh` to run arbitrary CLI commands with the locally-built binary
* Run `scripts/run-one-internal-testsuite.sh LANG` (replacing `LANG` with a language from the `supported-languages.txt` file) to run `kurtosis test` using the locally-built binary to run the internal testuite
* Run `scripts/run-all-internal-testsuites.sh` to run the internal testsuites in all languages
