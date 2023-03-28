#!/usr/bin/env bash
# 2021-07-08 WATERMARK, DO NOT REMOVE - This script was generated from the Kurtosis Bash script template

set -euo pipefail   # Bash "strict mode"

# ==================================================================================================
#                                             Main Logic
# ==================================================================================================

show_helptext_and_exit() {
    echo "Usage: $(basename "${0}") push_to_registry_container dockerfile_filepath image_tags..."
    echo ""
    echo "  push_to_registry_container     Whether images should be pushed to container registry or just loaded locally"
    echo "  dockerfile_filepath            The absolute path to the Dockerfile to use for this build"
    echo "  image_tags                     The tags for this image. Multiple values can be passed"
    echo ""
    exit 1  # Exit with an error so that if this is accidentally called by CI, the script will fail
}

# Raw parsing of the arguments
push_to_registry_container="${1:-}"
if [ "${push_to_registry_container}" != "true" ] && [ "${push_to_registry_container}" != "false" ]; then
    echo "Error: Invalid push_to_registry_container arg: '${push_to_registry_container}'" >&2
    show_helptext_and_exit
fi

dockerfile_filepath="${2:-}"
if [ ! -f "${dockerfile_filepath}" ]; then
    echo "Error: Invalid dockerfile_filepath arg: '${dockerfile_filepath}'" >&2
    show_helptext_and_exit
fi

image_tags="${*:3}"
if [ -z "${image_tags}" ]; then
    echo "Error: Invalid image_tags arg: '${image_tags}'" >&2
    show_helptext_and_exit
fi

# Argument processing
if "${push_to_registry_container}"; then
  buildx_platform_arg='linux/arm64/v8,linux/amd64'
  push_flag='--push'
else
  buildx_platform_arg='linux/amd64' # TODO: infer the local arch if that's reasonable
  push_flag='--load'
fi
echo "Building docker image for architecture '${buildx_platform_arg}' with flag '${push_flag}'"

dockerfile_dirpath="$(dirname "${dockerfile_filepath}")"
echo "Docker file located at '${dockerfile_filepath}' in directory '${dockerfile_dirpath}'"

image_tags_concatenated=""
for image_tag in ${image_tags}; do
  echo "Image will be tagged with: '${image_tag}'"
  image_tags_concatenated=" -t ${image_tag}${image_tags_concatenated} "
done

# Build Docker image

## Start by making sure the builder and the context do not already exist. If that's the case remove them
kurtosis_docker_builder="kurtosis-docker-builder"
docker_buildx_context='kurtosis-docker-builder-context'
if docker buildx inspect "${kurtosis_docker_builder}" &>/dev/null; then
  echo "Removing docker buildx builder ${kurtosis_docker_builder} as it seems to already exist"
  if ! docker buildx rm ${kurtosis_docker_builder} &>/dev/null; then
    echo "Failed removing docker buildx builder ${kurtosis_docker_builder}. Try removing it manually with 'docker buildx rm ${kurtosis_docker_builder}' before re-running this script"
    exit 1
  fi
fi
if docker context inspect "${docker_buildx_context}" &>/dev/null; then
  echo "Removing docker context ${docker_buildx_context} as it seems to already exist"
  if ! docker context rm ${docker_buildx_context} &>/dev/null; then
    echo "Failed removing docker context ${docker_buildx_context}. Try removing it manually with 'docker context rm ${docker_buildx_context}' before re-running this script"
    exit 1
  fi
fi

## Create Docker context and buildx builder
if ! docker context create "${docker_buildx_context}" &>/dev/null; then
  echo "Error: Docker context creation for buildx failed" >&2
  exit 1
fi
if ! docker buildx create --use --name "${kurtosis_docker_builder}" "${docker_buildx_context}" &>/dev/null; then
  echo "Error: Docker context switch for buildx failed" >&2d
  exit 1
fi

## Actually build the Docker image
docker_buildx_cmd="docker buildx build ${push_flag} --platform ${buildx_platform_arg} ${image_tags_concatenated} -f ${dockerfile_filepath} ${dockerfile_dirpath}"
echo "Running the following docker buildx command:"
echo "${docker_buildx_cmd}"
if ! eval "${docker_buildx_cmd}"; then
  echo "Error: Docker build failed" >&2
  exit 1
fi

# Cleanup context and buildx runner
echo "Cleaning up remaining resources"
if ! docker buildx rm "${kurtosis_docker_builder}" &>/dev/null; then
  echo "Warn: Failed removing the buildx builder '${kurtosis_docker_builder}'. Try manually removing it with 'docker buildx rm ${kurtosis_docker_builder}'" >&2
  exit 1
fi
if ! docker context rm "${docker_buildx_context}" &>/dev/null; then
  echo "Warn: Failed removing the buildx context '${docker_buildx_context}'. Try manually removing it with 'docker context rm ${docker_buildx_context}'" >&2
  exit 1
fi
echo "Successfully built docker image"
