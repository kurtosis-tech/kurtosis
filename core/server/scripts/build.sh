#!/usr/bin/env bash
# 2021-07-08 WATERMARK, DO NOT REMOVE - This script was generated from the Kurtosis Bash script template

set -euo pipefail   # Bash "strict mode"
script_dirpath="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"
server_root_dirpath="$(dirname "${script_dirpath}")"
git_repo_dirpath="$(dirname "$(dirname "${server_root_dirpath}")")"
uname_arch=$(uname -m)
# ==================================================================================================
#                                             Constants
# ==================================================================================================
source "${script_dirpath}/_constants.env"

if [ "$uname_arch" == "x86_64" ] || [ "$uname_arch" == "amd64" ]; then
    DEFAULT_ARCHITECTURE_TO_BUILD="amd64"
elif [ "$uname_arch" == "aarch64" ] || [ "$uname_arch" == "arm64" ]; then
    DEFAULT_ARCHITECTURE_TO_BUILD="arm64"
fi

MAIN_DIRNAME="api_container"
MAIN_GO_FILEPATH="${server_root_dirpath}/${MAIN_DIRNAME}/main.go"
MAIN_BINARY_OUTPUT_FILENAME="api-container"
MAIN_BINARY_OUTPUT_FILEPATH="${server_root_dirpath}/${BUILD_DIRNAME}/${MAIN_BINARY_OUTPUT_FILENAME}"

DELVE_BINARY_FILENAME="dlv"
DELVE_BINARY_OUTPUT_FILEPATH="${server_root_dirpath}/${BUILD_DIRNAME}/${DELVE_BINARY_FILENAME}"

# ==================================================================================================
#                                       Arg Parsing & Validation
# ==================================================================================================
show_helptext_and_exit() {
    echo "Usage: $(basename "${0}") skip_docker_image_building, architecture_to_build, debug_image..."
    echo ""
    echo "  skip_docker_image_building     Whether build the Docker image"
    echo "  architecture_to_build          The desired architecture for the project's binaries"
    echo "  debug_image                    Whether images should contains the debug server and run in debug mode, this will use the Dockerfile.debug image to build the container"
    echo ""
    exit 1  # Exit with an error so that if this is accidentally called by CI, the script will fail
}


skip_docker_image_building="${1:-"${DEFAULT_SKIP_DOCKER_IMAGE_BUILDING}"}"
if [ "${skip_docker_image_building}" != "true" ] && [ "${skip_docker_image_building}" != "false" ]; then
    echo "Error: Invalid skip-docker-image-building arg '${skip_docker_image_building}'" >&2
fi

architecture_to_build="${2:-"${DEFAULT_ARCHITECTURE_TO_BUILD}"}"
if [ "${architecture_to_build}" != "amd64" ] && [ "${architecture_to_build}" != "arm64" ]; then
    echo "Error: Invalid architecture-to-build arg '${architecture_to_build}'" >&2
fi

debug_image="${3:-"${DEFAULT_DEBUG_IMAGE}"}"
if [ "${debug_image}" != "true" ] && [ "${debug_image}" != "false" ]; then
    echo "Error: Invalid debug_image arg: '${debug_image}'" >&2
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
gc_flags=""
if "${debug_image}"; then
  gc_flags="all=-N -l"
fi
if ! CGO_ENABLED=0 GOOS=linux GOARCH=${architecture_to_build} go build -gcflags="${gc_flags}" -o "${MAIN_BINARY_OUTPUT_FILEPATH}.${architecture_to_build}" "${MAIN_GO_FILEPATH}"; then
  echo "Error: An error occurred building the server code" >&2
  exit 1
fi
echo "Successfully built server code"

# TODO move this to its own file because it's duplicated in the engine's build script
# Install and copy delve (the server debugger: https://github.com/go-delve/delve) to include it into the build folder so it can be picked up from the Docker image file
if "${debug_image}"; then
  # Checks if the binary already exist
  delve_exist="false"
  dlv_go_folder_bin_filepath=~/go/bin/linux_"${architecture_to_build}"/"${DELVE_BINARY_FILENAME}"
  dlv_bin_in_project_filepath="${DELVE_BINARY_OUTPUT_FILEPATH}.${architecture_to_build}"
  # Check if it is in the build folder and skip if it is there
  if ! [ -f "$dlv_bin_in_project_filepath" ]; then
    # Check if it already exist in the goland folder
    if ! [ -f "$dlv_go_folder_bin_filepath" ]; then
      echo "Installing delve..."
      if ! GOOS=linux GOARCH=${architecture_to_build} go install "${DELVE_SERVER_REPOSITORY_AND_VERSION}"; then
        echo "Error: An error occurred installing the delve binary" >&2
        exit 1
      fi
      echo "...delve successfully installed."
    fi
    # Now checks if the binary has been created
    if ! [ -f "$dlv_go_folder_bin_filepath" ]; then
        echo "Error: expected to have dlv binary in ${dlv_go_folder_bin_filepath} but it is not there."
        exit 1
    fi
    if ! cp "$dlv_go_folder_bin_filepath" "$dlv_bin_in_project_filepath" ; then
      echo "Error: something failed while copying the delve binary from the GO bin folder to the project's build folder"
      exit 1
    fi
  fi
fi

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

dockerfile_filepath="${server_root_dirpath}/${DOCKER_IMAGE_FILENAME}"

# Using the Docker debug image if it was requested
if "${debug_image}"; then
  dockerfile_filepath="${server_root_dirpath}/${DOCKER_DEBUG_IMAGE_FILENAME}"
fi

image_name="${IMAGE_ORG_AND_REPO}:${docker_tag}"
# specifying that this is a image for debugging
if "${debug_image}"; then
  image_name="${image_name}-${DOCKER_DEBUG_IMAGE_NAME_SUFFIX}"
fi

load_not_push_image=false
docker_build_script_cmd="${git_repo_dirpath}/scripts/docker-image-builder.sh ${load_not_push_image} ${dockerfile_filepath} ${image_name}"
if ! eval "${docker_build_script_cmd}"; then
  echo "Error: Docker build failed" >&2
fi
