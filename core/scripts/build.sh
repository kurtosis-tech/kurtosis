#!/usr/bin/env bash
# 2021-07-08 WATERMARK, DO NOT REMOVE - This script was generated from the Kurtosis Bash script template

set -euo pipefail   # Bash "strict mode"
script_dirpath="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"
root_dirpath="$(dirname "${script_dirpath}")"



# ==================================================================================================
#                                             Constants
# ==================================================================================================
source "${script_dirpath}/_constants.sh"

INITIALIZER_DIRNAME="initializer"

GET_DOCKER_IMAGES_TAG_SCRIPT_FILENAME="get-docker-images-tag.sh"

WRAPPER_GENERATOR_DIRNAME="wrapper_generator"
WRAPPER_GENERATOR_BINARY_OUTPUT_FILENAME="wrapper-generator"
WRAPPER_TEMPLATE_REL_FILEPATH="${WRAPPER_GENERATOR_DIRNAME}/kurtosis.template.sh"

WRAPPER_SCRIPT_GENERATOR_GORELEASER_BUILD_ID="wrapper-generator"

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
    INITIALIZER_IMAGE \
    INTERNAL_TESTSUITE_IMAGE \
    WRAPPER_GENERATOR_BINARY_OUTPUT_FILENAME \
    JAVASCRIPT_REPL_IMAGE \
    CLI_BINARY_FILENAME
export DOCKER_IMAGE_TAG="${docker_image_tag}"

# We want to run goreleaser from the root
cd "${root_dirpath}"

go test ./...

# TODO TODO TODO REMOVE ALL THIS WHEN THE CLI WILL DO EVERYTHING
# Generate the wrapper script
if ! goreleaser build --rm-dist --snapshot --id "${WRAPPER_SCRIPT_GENERATOR_GORELEASER_BUILD_ID}" --single-target; then
    echo "Error: Couldn't build the wrapper script-generating binary" >&2
    exit 1
fi
if ! "${root_dirpath}/${GORELEASER_OUTPUT_DIRNAME}/${WRAPPER_GENERATOR_BINARY_OUTPUT_FILENAME}" \
        -kurtosis-core-version "${docker_image_tag}" \
        -template "${WRAPPER_TEMPLATE_REL_FILEPATH}" \
        -output "${WRAPPER_OUTPUT_REL_FILEPATH}"; then
    echo "Error: Failed to generate wrapper script" >&2
    exit 1
fi

# Build all the Docker images
if ! goreleaser release --rm-dist --snapshot; then
    echo "Error: Goreleaser build of all binaries & Docker images failed" >&2
    exit 1
fi

# Build a CLI binary, compatible with the current OS & arch, so that we can run Interactive locally
if ! goreleaser build --rm-dist --snapshot --id "${GORELEASER_CLI_BUILD_ID}" --single-target; then
    echo "Error: Couldn't build the wrapper script-generating binary" >&2
    exit 1
fi
