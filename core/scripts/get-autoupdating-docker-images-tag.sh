#!/usr/bin/env bash
# 2021-07-08 WATERMARK, DO NOT REMOVE - This script was generated from the Kurtosis Bash script template

# This script will examine the Git history to generate a tag that will be given
#  to the Docker images produced by this repo

set -euo pipefail   # Bash "strict mode"
script_dirpath="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"


# ==================================================================================================
#                                             Constants
# ==================================================================================================
source "${script_dirpath}/_constants.sh"

# We release X.Y versions of this repo, so that patches will automatically get
#  applied for users
SHORTENED_RELEASE_VERSION_REGEX="^[0-9]+\.[0-9]+"

# Versions matching this regex will get shortened
RELEASE_VERSION_REGEX="${SHORTENED_RELEASE_VERSION_REGEX}\.[0-9]+$"


# ==================================================================================================
#                                             Main Logic
# ==================================================================================================
if ! fixed_version_tag="$(bash "${script_dirpath}/${GET_FIXED_DOCKER_IMAGES_TAG_SCRIPT_FILENAME}")"; then
    echo "Error: Couldn't get fixed version tag" >&2
    exit 1
fi

# Transform X.Y.Z => X.Y, and leave everything else untouched
#  get patch updates transparently
docker_tag="${fixed_version_tag}"
if [[ "${docker_tag}" =~ ${RELEASE_VERSION_REGEX} ]]; then
    docker_tag="$(echo "${docker_tag}" | egrep -o "${SHORTENED_RELEASE_VERSION_REGEX}")"
fi

echo "${docker_tag}"
