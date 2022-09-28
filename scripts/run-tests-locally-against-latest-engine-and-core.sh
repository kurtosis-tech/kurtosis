#!/usr/bin/env bash
# 2021-07-08 WATERMARK, DO NOT REMOVE - This script was generated from the Kurtosis Bash script template

set -euo pipefail   # Bash "strict mode"
script_dirpath="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"
root_dirpath="$(dirname "${script_dirpath}")"
cli_launch_path="${root_dirpath}/cli/cli/scripts/launch-cli.sh"
internal_test_suite_build_script_path="${root_dirpath}/internal_testsuites/scripts/build.sh"



# ==================================================================================================
#                                             Constants
# ==================================================================================================
BUILD_SCRIPT_RELATIVE_FILEPATHS=(
    "scripts/create-server-images-and-populate-env.sh"
    "cli/scripts/build.sh"
)

# ==================================================================================================
#                                             Main Logic
# ==================================================================================================
for build_script_rel_filepath in "${BUILD_SCRIPT_RELATIVE_FILEPATHS[@]}"; do
    build_script_abs_filepath="${root_dirpath}/${build_script_rel_filepath}"
    if ! bash "${build_script_abs_filepath}"; then
        echo "Error: Build script '${build_script_abs_filepath}' failed" >&2
        exit 1
    fi
done


source ${root_dirpath}/_kurtosis_servers.env

if ! bash "${cli_launch_path}" engine restart --version ${CORE_ENGINE_VERSION_TAG}; then
    echo "Error: Build script '${cli_launch_path}' failed" >&2
    exit 1
fi

if ! bash "${internal_test_suite_build_script_path}"; then
    echo "Error: Build script '${internal_test_suite_build_script_path}' failed" >&2
    exit 1
fi

