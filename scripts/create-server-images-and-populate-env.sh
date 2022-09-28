#!/usr/bin/env bash
# 2021-07-08 WATERMARK, DO NOT REMOVE - This script was generated from the Kurtosis Bash script template

set -euo pipefail   # Bash "strict mode"
script_dirpath="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"
git_repo_dirpath="$(dirname "${script_dirpath}")"
root_dirpath="$(dirname "${script_dirpath}")"
core_server_constants="${git_repo_dirpath}/core/server/scripts/_constants.env"
engine_server_constants="${git_repo_dirpath}/engine/server/scripts/_constants.env"

# ==================================================================================================
#                                             Constants
# ==================================================================================================

BUILD_SCRIPT_RELATIVE_FILEPATHS=(
  "core/server/scripts/build.sh"
  "engine/server/scripts/build.sh"
)


# ==================================================================================================
#                                             Main Logic
# ==================================================================================================

# Generate Docker image tag
if ! cd "${git_repo_dirpath}"; then
  echo "Error: Couldn't cd to the git root dirpath '${server_root_dirpath}'" >&2
  exit 1
fi
if ! docker_tag="$(kudet get-docker-tag)"; then
    echo "Error: Couldn't get the Docker image tag" >&2
    exit 1
fi

source "${core_server_constants}"
core_image_name="${IMAGE_ORG_AND_REPO}:${docker_tag}"

source "${engine_server_constants}"
engine_image_name="${IMAGE_ORG_AND_REPO}:${docker_tag}"

for build_script_rel_filepath in "${BUILD_SCRIPT_RELATIVE_FILEPATHS[@]}"; do
    build_script_abs_filepath="${root_dirpath}/${build_script_rel_filepath}"
    if ! bash "${build_script_abs_filepath}"; then
        echo "Error: Build script '${build_script_abs_filepath}' failed" >&2
        exit 1
    fi
done

echo "Writing versions to a temporary file so that it can be exported"
echo "export KURTOSIS_CORE_SERVER_IMAGE_AND_TAG=${core_image_name}" > /tmp/_kurtosis_servers.env
echo "export KURTOSIS_ENGINE_SERVER_IMAGE_AND_TAG=${engine_image_name}" >> /tmp/_kurtosis_servers.env