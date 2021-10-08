#!/usr/bin/env bash
# 2021-07-08 WATERMARK, DO NOT REMOVE - This script was generated from the Kurtosis Bash script template

set -euo pipefail   # Bash "strict mode"
script_dirpath="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"
root_dirpath="$(dirname "${script_dirpath}")"



# ==================================================================================================
#                                             Constants
# ==================================================================================================
source "${script_dirpath}/_constants.sh"
GET_AUTOUPDATING_DOCKER_IMAGES_TAG_SCRIPT_FILENAME="get-autoupdating-docker-images-tag.sh"
DEFAULT_SHOULD_PUBLISH_ARG="false"


# ==================================================================================================
#                                       Arg Parsing & Validation
# ==================================================================================================
show_helptext_and_exit() {
    echo "Usage: $(basename "${0}") [should_publish_arg]"
    echo ""
    echo "  should_publish_arg  Whether the build artifacts should be published (default: ${DEFAULT_SHOULD_PUBLISH_ARG})"
    echo ""
    exit 1  # Exit with an error so that if this is accidentally called by CI, the script will fail
}

should_publish_arg="${1:-"${DEFAULT_SHOULD_PUBLISH_ARG}"}"
if [ "${should_publish_arg}" != "true" ] && [ "${should_publish_arg}" != "false" ]; then
    echo "Error: Invalid should-publish arg '${should_publish_arg}'" >&2
    show_helptext_and_exit
fi

# ==================================================================================================
#                                             Main Logic
# ==================================================================================================
build_dirpath="${root_dirpath}/${BUILD_DIRNAME}"
if ! mkdir -p "${build_dirpath}"; then
    echo "Error: Couldn't create build output dir '${build_dirpath}'" >&2
    exit 1
fi

if ! fixed_docker_image_tag="$(bash "${script_dirpath}/${GET_FIXED_DOCKER_IMAGES_TAG_SCRIPT_FILENAME}")"; then
    echo "Error: Couldn't get fixed Docker image tag" >&2
    exit 1
fi
if ! autoupdating_docker_image_tag="$(bash "${script_dirpath}/${GET_AUTOUPDATING_DOCKER_IMAGES_TAG_SCRIPT_FILENAME}")"; then
    echo "Error: Couldn't get autoupdating Docker image tag" >&2
    exit 1
fi

# These variables are used by Goreleaser
export API_IMAGE
export FIXED_DOCKER_IMAGE_TAG="${fixed_docker_image_tag}"
export AUTOUPDATING_DOCKER_IMAGE_TAG="${autoupdating_docker_image_tag}"

# We want to run goreleaser from the root
cd "${root_dirpath}"

go test ./...

# Build all the Docker images
if "${should_publish_arg}"; then
    goreleaser_release_extra_args=""
else
    goreleaser_release_extra_args="--snapshot"
fi
if ! goreleaser release --rm-dist --skip-announce ${goreleaser_release_extra_args}; then
    echo "Error: Goreleaser release of all binaries & Docker images failed" >&2
    exit 1
fi
