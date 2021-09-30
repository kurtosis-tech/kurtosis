#!/usr/bin/env bash
# 2021-07-08 WATERMARK, DO NOT REMOVE - This script was generated from the Kurtosis Bash script template

set -euo pipefail   # Bash "strict mode"
script_dirpath="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"
root_dirpath="$(dirname "${script_dirpath}")"



# ==================================================================================================
#                                             Constants
# ==================================================================================================
source "${script_dirpath}/_constants.sh"

GET_DOCKER_IMAGES_TAG_SCRIPT_FILENAME="get-docker-images-tag.sh"

WRAPPER_GENERATOR_DIRNAME="wrapper_generator"
WRAPPER_GENERATOR_BINARY_OUTPUT_FILENAME="wrapper-generator"
WRAPPER_TEMPLATE_REL_FILEPATH="${WRAPPER_GENERATOR_DIRNAME}/kurtosis.template.sh"

WRAPPER_SCRIPT_GENERATOR_GORELEASER_BUILD_ID="wrapper-generator"

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

if ! docker_image_tag="$(bash "${script_dirpath}/${GET_DOCKER_IMAGES_TAG_SCRIPT_FILENAME}")"; then
    echo "Error: Couldn't get the Docker images tag" >&2
    exit 1
fi

# TODO GENERATE JAVASCRIPT REPL

# These variables are used by Goreleaser
export API_IMAGE \
    DOCKER_ORG \
    INTERNAL_TESTSUITE_IMAGE_SUFFIX \
    JAVASCRIPT_REPL_IMAGE \
    CLI_BINARY_FILENAME
export DOCKER_IMAGE_TAG="${docker_image_tag}"
if "${should_publish_arg}"; then
    export GEMFURY_PUBLISH_TOKEN
fi

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

# Build a CLI binary, compatible with the current OS & arch, so that we can run interactive & testing locally
if ! goreleaser build --rm-dist --snapshot --id "${GORELEASER_CLI_BUILD_ID}" --single-target; then
    echo "Error: Couldn't build the wrapper script-generating binary" >&2
    exit 1
fi
