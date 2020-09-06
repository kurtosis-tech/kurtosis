#!/bin/bash

set -euo pipefail
script_dirpath="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"

# Constants
DOCKER_ORG="kurtosistech"
REPO_BASE="kurtosis-core"
API_REPO="${REPO_BASE}_api"
INITIALIZER_REPO="${REPO_BASE}_initializer"
LATEST_TAG="latest"

root_dirpath="$(dirname "${script_dirpath}")"

initializer_image_tag="${DOCKER_ORG}/${INITIALIZER_REPO}:${LATEST_TAG}"
docker build -t "${initializer_image_tag}" "${root_dirpath}" -f "${root_dirpath}/initializer/Dockerfile"

api_image_tag="${DOCKER_ORG}/${API_REPO}:${LATEST_TAG}"
docker build -t "${api_image_tag}" "${root_dirpath}" -f "${root_dirpath}/api_container/Dockerfile"
