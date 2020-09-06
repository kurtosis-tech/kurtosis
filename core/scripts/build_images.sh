#!/bin/bash

set -euo pipefail
script_dirpath="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"

# Constants
DOCKER_ORG="kurtosistech"
REPO_BASE="kurtosis-core"
API_REPO="${REPO_BASE}_api"
INITIALIZER_REPO="${REPO_BASE}_initializer"
LATEST_TAG="latest"
INITIALIZER_IMAGE_TAG="${DOCKER_ORG}/${INITIALIZER_REPO}:${LATEST_TAG}"
API_IMAGE_TAG="${DOCKER_ORG}/${API_REPO}:${LATEST_TAG}"

root_dirpath="$(dirname "${script_dirpath}")"

# To cut down on build time, we build our images in parallel
initializer_log_filepath="$(mktemp)"
api_log_filepath="$(mktemp)"

echo "Launching builds of initializer & API images in parallel threads..."
docker build -t "${INITIALIZER_IMAGE_TAG}" "${root_dirpath}" -f "${root_dirpath}/initializer/Dockerfile" 2>&1 > "${initializer_log_filepath}" &
initializer_build_pid="${!}"
docker build -t "${API_IMAGE_TAG}" "${root_dirpath}" -f "${root_dirpath}/api_container/Dockerfile" 2>&1 > "${api_log_filepath}" &
api_build_pid="${!}"
echo "Build threads launched successfully"

echo "Waiting for build threads to exit... (initializer PID: ${initializer_build_pid}, API PID: ${api_build_pid})"
builds_succeeded=true
if ! wait "${initializer_build_pid}"; then
    builds_succeeded=false
fi
if ! wait "${api_build_pid}"; then
    builds_succeeded=false
fi
echo "Build threads exited"

echo ""
echo "===================== Initializer Image Build Logs =========================="
cat "${initializer_log_filepath}"

echo ""
echo "========================= API Image Build Logs =============================="
cat "${api_log_filepath}"

echo ""
if "${builds_succeeded}"; then
    echo "Build SUCCEEDED"
    exit 0
else
    echo "Build FAILED"
    exit 1
fi
