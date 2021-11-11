#!/usr/bin/env bash
# 2021-07-08 WATERMARK, DO NOT REMOVE - This script was generated from the Kurtosis Bash script template

set -euo pipefail   # Bash "strict mode"
script_dirpath="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"
root_dirpath="$(dirname "${script_dirpath}")"



# ==================================================================================================
#                                             Constants
# ==================================================================================================
source "${script_dirpath}/_constants.sh"

INTERNAL_TESTSUITE_DIRNAMES=(
    "golang_internal_testsuite"
)

# ==================================================================================================
#                                             Main Logic
# ==================================================================================================
goarch="$(go env GOARCH)"
goos="$(go env GOOS)"
cli_binary_filepath="${root_dirpath}/${CLI_MODULE_DIRNAME}/${GORELEASER_OUTPUT_DIRNAME}/${GORELEASER_CLI_BUILD_ID}_${goos}_${goarch}/${CLI_BINARY_FILENAME}"

for internal_testsuite_dirname in "${INTERNAL_TESTSUITE_DIRNAMES[@]}"; do
    internal_testsuite_buildscript_filepath="${root_dirpath}/${internal_testsuite_dirname}/scripts/build.sh"
    if ! [ -f "${internal_testsuite_dirpath}" ]; then
        echo "Error: Expected a build script for internal testsuite '${internal_testsuite_dirname}' at '${internal_testsuite_buildscript_filepath}' but none was found" >&2
        exit 1
    fi
    if ! "${internal_testsuite_buildscript_filepath}"; then
        echo "Error: Internal testsuite '${internal_testsuite_dirname}' failed" >&2
        exit 1
    fi
done
echo "All internal testsuites completed successfully"
