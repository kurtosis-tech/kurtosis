#!/usr/bin/env bash
set -euo pipefail # Bash "strict mode"
script_dirpath="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"
root_dirpath="$(dirname "${script_dirpath}")"

# ==================================================================================================
#                                             Constants
# ==================================================================================================
BUILD_DIRNAME="build"
IMAGE_ORG_AND_REPO="kurtosistech/kurtosis-engine-server"
ENGINE_DIRNAME="engine"
MAIN_BINARY_OUTPUT_FILE="kurtosis-engine.bin"
MAIN_BINARY_OUTPUT_PATH="${root_dirpath}/${BUILD_DIRNAME}/${MAIN_BINARY_OUTPUT_FILE}"
MAIN_GO_FILEPATH="${root_dirpath}/${ENGINE_DIRNAME}/main.go"

GET_DOCKER_TAG_SCRIPT_FILENAME="get-docker-image-tag.sh"

# =============================================================================
#                                 Main Code
# =============================================================================
# Checks if dockerignore file is in the root path
if ! [ -f "${root_dirpath}"/.dockerignore ]; then
  echo "Error: No .dockerignore file found in root '${root_dirpath}'; this is required so Docker caching is enabled and the engine image builds remain quick" >&2
  exit 1
fi

# Test code
echo "Running unit tests..."
if ! go test "${root_dirpath}/..."; then
  echo "Tests failed!"
  exit 1
fi
echo "Tests succeeded"

# Build Go code
echo "Building Kurtosis Engine Server code '${MAIN_GO_FILEPATH}'..."
if ! GOOS=linux go build -o "${MAIN_BINARY_OUTPUT_PATH}" "${MAIN_GO_FILEPATH}"; then
  echo "Error: Code build of the Kurtosis Engine Server failed" >&2
  exit 1
fi

# Generate Docker image tag
get_docker_image_tag_script_filepath="${script_dirpath}/${GET_DOCKER_TAG_SCRIPT_FILENAME}"
if ! docker_tag="$(bash "${get_docker_image_tag_script_filepath}")"; then
    echo "Error: Couldn't get the Docker image tag" >&2
    exit 1
fi

# Build Docker image
dockerfile_filepath="${root_dirpath}/${ENGINE_DIRNAME}/Dockerfile"
image_name="${IMAGE_ORG_AND_REPO}:${docker_tag}"
echo "Building Kurtosis Engine Server into a Docker image named '${image_name}'..."
if ! docker build -t "${image_name}" -f "${dockerfile_filepath}" "${root_dirpath}"; then
  echo "Error: Docker build of the Kurtosis Engine Server failed" >&2
  exit 1
fi
echo "Successfully built Docker image '${image_name}' containing the Kurtosis Engine Server"
