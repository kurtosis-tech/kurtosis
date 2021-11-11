#!/usr/bin/env bash
# 2021-07-08 WATERMARK, DO NOT REMOVE - This script was generated from the Kurtosis Bash script template

set -euo pipefail   # Bash "strict mode"
script_dirpath="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"
repls_dirpath="$(dirname "${script_dirpath}")"
root_dirpath="$(dirname "${repls_dirpath}")"


# ==================================================================================================
#                                             Constants
# ==================================================================================================
DOCKER_IMAGE_PREFIX="kurtosistech/"
DOCKER_IMAGE_SUFFIX="-interactive-repl"
DOCKERIGNORE_FILENAME=".dockerignore"

GET_DOCKER_TAG_SCRIPT_FILENAME="get-docker-images-tag.sh"

BUILD_DIRNAME="build"

REPL_DOCKERFILE_GENERATOR_MODULE_DIRNAME="dockerfile_generator"
REPL_DOCKERFILE_TEMPLATE_FILENAME="template.Dockerfile"
REPL_DOCKERFILE_GENERATOR_BINARY_OUTPUT_FILENAME="repl-dockerfile-generator"

REPL_IMAGES_DIRNAME="repl_images"

REPL_OUTPUT_DOCKERFILE_SUFFIX=".Dockerfile"



# ==================================================================================================
#                                             Main Logic
# ==================================================================================================
get_docker_tag_script_filepath="${root_dirpath}/scripts/${GET_DOCKER_TAG_SCRIPT_FILENAME}"
if ! docker_tag="$("${get_docker_tag_script_filepath}")"; then
    echo "Error: Couldn't get tag to give Docker images using script '${get_docker_tag_script_filepath}'" >&2
    exit 1
fi

# Build REPL Dockefile-generating binary
echo "Building REPL Dockerfile-generating binary..."
repl_dockerfile_generator_module_dirpath="${repls_dirpath}/${REPL_DOCKERFILE_GENERATOR_MODULE_DIRNAME}"
repl_dockerfile_generator_binary_filepath="${repls_dirpath}/${BUILD_DIRNAME}/${REPL_DOCKERFILE_GENERATOR_BINARY_OUTPUT_FILENAME}"
(
    if ! cd "${repl_dockerfile_generator_module_dirpath}"; then
        echo "Error: Couldn't cd to the REPL Dockerfile-generating module dirpath '${repl_dockerfile_generator_module_dirpath}'" >&2
        exit 1
    fi
    if ! go build -o "${repl_dockerfile_generator_binary_filepath}"; then
        echo "Error: Build of the REPL Dockerfile-generating binary failed" >&2
        exit 1
    fi
)
echo "REPL Dockerfile-generating binary built successfully"

# Now, use the built binary to generate REPL Dockerfiles and build Docker images
echo "Generating REPL Dockerfiles..."
repl_images_dirpath="${repls_dirpath}/${REPL_IMAGES_DIRNAME}"
build_dirpath="${repls_dirpath}/${BUILD_DIRNAME}"
for repl_image_dirpath in $(find "${repl_images_dirpath}" -type d -mindepth 1 -maxdepth 1); do
    repl_type="$(basename "${repl_image_dirpath}")"
    echo "Building Docker image for '${repl_type}' REPL..."
    repl_dockerfile_template_filepath="${repl_image_dirpath}/${REPL_DOCKERFILE_TEMPLATE_FILENAME}"
    if ! [ -f "${repl_dockerfile_template_filepath}" ]; then
        echo "Error: Tried to generate Dockerfile for REPL '${repl_type}' but no template file was found at path '${repl_dockerfile_template_filepath}'" >&2
        exit 1
    fi
    output_filepath="${build_dirpath}/${repl_type}${REPL_OUTPUT_DOCKERFILE_SUFFIX}"
    if ! "${repl_dockerfile_generator_binary_filepath}" "${repl_dockerfile_template_filepath}" "${output_filepath}" "${repl_type}"; then
        echo "Error: An error occurred rendering template for REPL '${repl_type}' at path '${repl_dockerfile_template_filepath}' to output filepath '${output_filepath}'" >&2
        exit 1
    fi

    dockerignore_filepath="${repl_image_dirpath}/${DOCKERIGNORE_FILENAME}"
    if ! [ -f "${dockerignore_filepath}" ]; then
        echo "Error: No ${DOCKERIGNORE_FILENAME} found at '${dockerignore_filepath}'; this is required so that build caching is effective" >&2
        exit 1
    fi

    docker_image_name="${DOCKER_IMAGE_PREFIX}${repl_type}${DOCKER_IMAGE_SUFFIX}:${docker_tag}"
    if ! docker build -t "${docker_image_name}" -f "${output_filepath}" "${repl_image_dirpath}"; then
        echo "Error: Build of '${repl_type}' REPL using Dockerfile '${output_filepath}' and build context '${repl_image_dirpath}' failed" >&2
        exit 1
    fi
    echo "Successfully built Docker image for '${repl_type}' REPL"
done
