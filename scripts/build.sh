#!/usr/bin/env bash
# 2021-07-08 WATERMARK, DO NOT REMOVE - This script was generated from the Kurtosis Bash script template

set -euo pipefail   # Bash "strict mode"
script_dirpath="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"
root_dirpath="$(dirname "${script_dirpath}")"

if ! bash "${script_dirpath}/versions_check.sh"; then
  exit 1
fi


# ==================================================================================================
#                                             Constants
# ==================================================================================================
BUILD_SCRIPT_RELATIVE_FILEPATHS=(
    "scripts/generate-kurtosis-version.sh"
    "container-engine-lib/scripts/build.sh"
    "contexts-config-store/scripts/build.sh"
    "grpc-file-transfer/scripts/build.sh"
    "name_generator/scripts/build.sh"
    "api/scripts/build.sh"
    "core/scripts/build.sh"
    "engine/scripts/build.sh"
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
