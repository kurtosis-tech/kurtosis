#!/usr/bin/env bash
# 2021-07-08 WATERMARK, DO NOT REMOVE - This script was generated from the Kurtosis Bash script template

# This script will examine the Git history to generate a tag that will be given
#  to the Docker images produced by this repo

set -euo pipefail   # Bash "strict mode"
script_dirpath="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"

# ==================================================================================================
#                                             Constants
# ==================================================================================================
# Regex that the Git ref will be passed through to sanitize it for becoming
#  a Docker image tag
GIT_REF_SANITIZING_SED_REGEX="s,[/:],_,g"

# Versions matching this regex will get shortened
RELEASE_VERSION_REGEX="^[0-9]+\.[0-9]+\.[0-9]+$"

# ==================================================================================================
#                                             Main Logic
# ==================================================================================================
# Captures the first of tag > branch > commit
if ! git_ref="$(git describe --tags --exact-match 2> /dev/null || git symbolic-ref -q --short HEAD || git rev-parse --short HEAD)"; then
    echo "Error: Couldn't get a Git ref to use for a Docker tag" >&2
    exit 1
fi
# Sanitize git ref to be acceptable Docker tag format
if ! docker_tag="$(echo "${git_ref}" | sed "${GIT_REF_SANITIZING_SED_REGEX}")"; then
    echo "Error: Couldn't sanitize Git ref to acceptable Docker tag format" >&2
    exit 1
fi

echo "${docker_tag}"
