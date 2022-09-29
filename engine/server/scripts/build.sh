#!/usr/bin/env bash
# 2021-07-08 WATERMARK, DO NOT REMOVE - This script was generated from the Kurtosis Bash script template

set -euo pipefail # Bash "strict mode"
script_dirpath="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"
server_root_dirpath="$(dirname "${script_dirpath}")"
git_repo_dirpath="$(dirname "$(dirname "${server_root_dirpath}")")"

# ==================================================================================================
#                                             Constants
# ==================================================================================================
source "${script_dirpath}/_constants.env"

BUILD_DIRNAME="build"

MAIN_DIRNAME="engine"
MAIN_GO_FILEPATH="${server_root_dirpath}/${MAIN_DIRNAME}/main.go"
MAIN_BINARY_OUTPUT_FILENAME="kurtosis-engine"
MAIN_BINARY_OUTPUT_FILEPATH="${server_root_dirpath}/${BUILD_DIRNAME}/${MAIN_BINARY_OUTPUT_FILENAME}"

SHOULD_RUN_TEST_TRUE="true"
SHOULD_RUN_TEST_FALSE="false"
DEFAULT_SHOULD_RUN_TEST="${SHOULD_RUN_TEST_TRUE}"
# ==================================================================================================
#                                       Arg Parsing & Validation
# ==================================================================================================
show_helptext_and_exit() {
    echo "Usage: $(basename "${0}") run_tests"
    echo ""
    echo "  should_run_test   Optional argument to decide whether to run tests. Must be one of 'true' 'false' (default: ${DEFAULT_SHOULD_RUN_TEST})"
    echo ""
    exit 1  # Exit with an error so that if this is accidentally called by CI, the script will fail
}

should_run_test_arg="${1:-${DEFAULT_SHOULD_RUN_TEST}}"
if [ "${should_run_test_arg}" != "${SHOULD_RUN_TEST_TRUE}" ] &&
   [ "${should_run_test_arg}" != "${SHOULD_RUN_TEST_FALSE}" ]; then
    echo "Error: unknown value for whether tests should be run. Should be one of 'true' or 'false'"
    show_helptext_and_exit
fi

# =============================================================================
#                                 Main Code
# =============================================================================
# Checks if dockerignore file is in the root path
if ! [ -f "${server_root_dirpath}"/.dockerignore ]; then
  echo "Error: No .dockerignore file found in server root '${server_root_dirpath}'; this is required so Docker caching is enabled and the image builds remain quick" >&2
  exit 1
fi

# Test code
if ! cd "${server_root_dirpath}"; then
  echo "Couldn't cd to the server root dirpath '${server_root_dirpath}'" >&2
  exit 1
fi

if [ "${should_run_test_arg}" == "${SHOULD_RUN_TEST_TRUE}" ] ; then
  echo "Running unit tests..."
  if ! CGO_ENABLED=0 go test "./..."; then
    echo "Tests failed!" >&2
    exit 1
  fi
  echo "Tests succeeded"
fi

# Build binary for packaging inside an Alpine Linux image
echo "Building server main.go '${MAIN_GO_FILEPATH}'..."
if ! CGO_ENABLED=0 GOOS=linux GOARCH=amd64 go build -o "${MAIN_BINARY_OUTPUT_FILEPATH}" "${MAIN_GO_FILEPATH}"; then
  echo "Error: An error occurred building the server code" >&2
  exit 1
fi
echo "Successfully built server code"

# Generate Docker image tag
if ! cd "${git_repo_dirpath}"; then
  echo "Error: Couldn't cd to the git root dirpath '${server_root_dirpath}'" >&2
  exit 1
fi
if ! docker_tag="$(kudet get-docker-tag)"; then
    echo "Error: Couldn't get the Docker image tag" >&2
    exit 1
fi

# Build Docker image
dockerfile_filepath="${server_root_dirpath}/Dockerfile"
image_name="${IMAGE_ORG_AND_REPO}:${docker_tag}"
echo "Building server into a Docker image named '${image_name}'..."
if ! docker build -t "${image_name}" -f "${dockerfile_filepath}" "${server_root_dirpath}"; then
  echo "Error: Docker build of the server failed" >&2
  exit 1
fi
echo "Successfully built Docker image '${image_name}' containing the server"
