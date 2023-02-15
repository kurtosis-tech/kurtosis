#!/usr/bin/env bash
# 2021-07-08 WATERMARK, DO NOT REMOVE - This script was generated from the Kurtosis Bash script template

set -euo pipefail   # Bash "strict mode"
script_dirpath="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"
internal_testsuites_root_dirpath="$(dirname "${script_dirpath}")"

# ==================================================================================================
#                                             Constants
# ==================================================================================================
CHILD_TEST_SCRIPT_FILENAME="test.sh"
STARLARK_DIR_PATH="starlark"

# ==================================================================================================
#                                       Arg Parsing & Validation
# ==================================================================================================
# ==================================================================================================
#                                             Main Logic
# ==================================================================================================
script_dirname="$(basename "${script_dirpath}")"

for maybe_testsuite_rel_dirpath in $(find "${internal_testsuites_root_dirpath}" -type d -mindepth 1 -maxdepth 1 ); do
    maybe_testsuite_dirname="$(basename "${maybe_testsuite_rel_dirpath}")"
    if [ "${maybe_testsuite_dirname}" == "${script_dirname}" ] || [ "${maybe_testsuite_dirname}" == "${STARLARK_DIR_PATH}" ]; then
        continue
    fi

    echo "Running buildscript for testsuite '${maybe_testsuite_dirname}'..."
    testsuite_buildscript_filepath="${maybe_testsuite_rel_dirpath}/scripts/${CHILD_TEST_SCRIPT_FILENAME}"
    if ! [ -f "${testsuite_buildscript_filepath}" ]; then
        echo "Error: Expected a buildscript for testsuite directory '${maybe_testsuite_dirname}' at '${testsuite_buildscript_filepath}', but none was found" >&2
        exit 1
    fi
    if ! "${testsuite_buildscript_filepath}"; then
        echo "Error: Testsuite buildscript at '${testsuite_buildscript_filepath}' failed" >&2
        exit 1
    fi
    echo "Buildscript for testsuite '${maybe_testsuite_dirname}' succeeded"
done
echo "All testsuites succeeded!"
