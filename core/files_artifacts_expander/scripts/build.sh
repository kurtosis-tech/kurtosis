#
# Copyright (c) 2022 - present Kurtosis Technologies Inc.
# All Rights Reserved.
#

set -euo pipefail   # Bash "strict mode"
script_dirpath="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"
expander_root_dirpath="$(dirname "${script_dirpath}")"

# ==================================================================================================
#                                             Constants
# ==================================================================================================
source "${script_dirpath}/_constants.env"

BUILD_DIRNAME="build"

MAIN_GO_FILEPATH="${expander_root_dirpath}/main.go"
MAIN_BINARY_OUTPUT_FILENAME="files-artifacts-expander"
MAIN_BINARY_OUTPUT_FILEPATH="${expander_root_dirpath}/${BUILD_DIRNAME}/${MAIN_BINARY_OUTPUT_FILENAME}"

# =============================================================================
#                                 Main Code
# =============================================================================
# Checks if dockerignore file is in the root path
if ! [ -f "${expander_root_dirpath}"/.dockerignore ]; then
  echo "Error: No .dockerignore file found in files artifacts expander root '${expander_root_dirpath}'; this is required so Docker caching is enabled and the image builds remain quick" >&2
  exit 1
fi

# Test code
echo "Running unit tests..."
if ! cd "${expander_root_dirpath}"; then
  echo "Couldn't cd to the files artifacts expander root dirpath '${server_root_dirpath}'" >&2
  exit 1
fi
if ! CGO_ENABLED=0 go test "./..."; then
  echo "Tests failed!" >&2
  exit 1
fi
echo "Tests succeeded"

# Build binary for packaging inside an Alpine Linux image
echo "Building files artifacts expander main.go '${MAIN_GO_FILEPATH}'..."
if ! CGO_ENABLED=0 GOOS=linux GOARCH=amd64 go build -o "${MAIN_BINARY_OUTPUT_FILEPATH}" "${MAIN_GO_FILEPATH}"; then
  echo "Error: An error occurred building the files artifacts expander code" >&2
  exit 1
fi
echo "Successfully built files artifacts expander code"

# Generate Docker image tag
get_docker_image_tag_script_filepath="${script_dirpath}/${GET_DOCKER_IMAGE_TAG_SCRIPT_FILENAME}"
if ! docker_tag="$(bash "${get_docker_image_tag_script_filepath}")"; then
    echo "Error: Couldn't get the Docker image tag" >&2
    exit 1
fi

# Build Docker image
dockerfile_filepath="${expander_root_dirpath}/Dockerfile"
image_name="${IMAGE_ORG_AND_REPO}:${docker_tag}"
echo "Building files artifacts expander into a Docker image named '${image_name}'..."
if ! docker build -t "${image_name}" -f "${dockerfile_filepath}" "${expander_root_dirpath}"; then
  echo "Error: Docker build of the files artifacts expander failed" >&2
  exit 1
fi
echo "Successfully built Docker image '${image_name}' containing the files artifacts expander"

