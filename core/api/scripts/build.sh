#!/usr/bin/env bash
# 2021-07-08 WATERMARK, DO NOT REMOVE - This script was generated from the Kurtosis Bash script template

set -euo pipefail   # Bash "strict mode"
script_dirpath="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"
api_root_dirpath="$(dirname "${script_dirpath}")"



# ==================================================================================================
#                                             Constants
# ==================================================================================================
PROTOBUF_DIRNAME="protobuf"  # Should be skipped; doesn't have a build.sh
BUILDSCRIPT_REL_DIRPATH="scripts/build.sh"

# ==================================================================================================
#                                             Main Logic
# ==================================================================================================
api_subproject_dirpaths="$(find "${api_root_dirpath}" -type d -maxdepth 1 -mindepth 1)"
for api_subproject_dirpath in ${api_subproject_dirpaths}; do
    if [ "$(basename "${api_subproject_dirpath}")" == "${PROTOBUF_DIRNAME}" ]; then
        continue
    fi
    if [ "${api_subproject_dirpath}" == "${script_dirpath}" ]; then
        continue
    fi
    build_script_dirpath="${api_subproject_dirpath}/${BUILDSCRIPT_REL_DIRPATH}"
    if ! bash "${build_script_dirpath}"; then
        echo "Error: Build script '${build_script_dirpath}' failed" >&2
        exit 1
    fi
done
