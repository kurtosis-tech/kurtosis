#!/bin/bash

set -euo pipefail
script_dirpath="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"

DOCKER_ORG="kurtosistech"
REPO_BASE="kurtosis-core"
API_REPO="${REPO_BASE}_api"
LATEST_TAG="latest"

root_dirpath="$(dirname "${script_dirpath}")"

tag="${DOCKER_ORG}/${API_REPO}:${LATEST_TAG}"
docker build -t "${tag}" "${root_dirpath}" -f "${root_dirpath}/api_container/Dockerfile"
