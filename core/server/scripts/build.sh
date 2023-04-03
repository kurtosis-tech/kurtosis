#!/usr/bin/env bash
# 2021-07-08 WATERMARK, DO NOT REMOVE - This script was generated from the Kurtosis Bash script template

set -euo pipefail   # Bash "strict mode"
script_dirpath="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"
server_root_dirpath="$(dirname "${script_dirpath}")"
git_repo_dirpath="$(dirname "$(dirname "${server_root_dirpath}")")"
# ==================================================================================================
#                                             Constants
# ==================================================================================================
source "${script_dirpath}/_constants.env"

BUILD_DIRNAME="build"

DEFAULT_SKIP_DOCKER_IMAGE_BUILDING=false

MAIN_DIRNAME="api_container"
MAIN_GO_FILEPATH="${server_root_dirpath}/${MAIN_DIRNAME}/main.go"
MAIN_BINARY_OUTPUT_FILENAME="api-container"
MAIN_BINARY_OUTPUT_FILEPATH="${server_root_dirpath}/${BUILD_DIRNAME}/${MAIN_BINARY_OUTPUT_FILENAME}"

# =============================================================================
#                                 Main Code
# =============================================================================
skip_docker_image_building="${1:-"${DEFAULT_SKIP_DOCKER_IMAGE_BUILDING}"}"
if [ "${skip_docker_image_building}" != "true" ] && [ "${skip_docker_image_building}" != "false" ]; then
    echo "Error: Invalid skip-docker-image-building arg '${skip_docker_image_building}'" >&2
fi

# Checks if dockerignore file is in the root path
if ! [ -f "${server_root_dirpath}"/.dockerignore ]; then
  echo "Error: No .dockerignore file found in server root '${server_root_dirpath}'; this is required so Docker caching is enabled and the image builds remain quick" >&2
  exit 1
fi

# Test code
echo "Running unit tests..."
if ! cd "${server_root_dirpath}"; then
  echo "Couldn't cd to the server root dirpath '${server_root_dirpath}'" >&2
  exit 1
fi
if ! CGO_ENABLED=0 go test "./..."; then
  echo "Tests failed!" >&2
  exit 1
fi
echo "Tests succeeded"

# Build binary for packaging inside an Alpine Linux image
echo "Building server main.go '${MAIN_GO_FILEPATH}'..."
if ! CGO_ENABLED=1 GOOS=linux GOARCH=amd64 CC="x86_64-linux-musl-gcc" CXX="x86_64-linux-musl-g++" go build -o "${MAIN_BINARY_OUTPUT_FILEPATH}" "${MAIN_GO_FILEPATH}"; then
  echo "Error: An error occurred building the server code" >&2
  exit 1
fi
echo "Successfully built server code"

# Generate Docker image tag
if ! cd "${git_repo_dirpath}"; then
  echo "Error: Couldn't cd to the git root dirpath '${git_repo_dirpath}'" >&2
  exit 1
fi
if ! docker_tag="$(./scripts/get-docker-tag.sh)"; then
    echo "Error: Couldn't get the Docker image tag" >&2
    exit 1
fi

# Build Docker image
if "${skip_docker_image_building}"; then
  echo "Not building docker image as requested"
  exit 0
fi

dockerfile_filepath="${server_root_dirpath}/Dockerfile"
image_name="${IMAGE_ORG_AND_REPO}:${docker_tag}"
load_not_push_image=false
docker_build_script_cmd="${git_repo_dirpath}/scripts/docker-image-builder.sh ${load_not_push_image} ${dockerfile_filepath} ${image_name}"
if ! eval "${docker_build_script_cmd}"; then
  echo "Error: Docker build failed" >&2
fi
