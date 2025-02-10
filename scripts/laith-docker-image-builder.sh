#!/usr/bin/env bash
# 2021-07-08 WATERMARK, DO NOT REMOVE - This script was generated from the Kurtosis Bash script template

set -euo pipefail   # Bash "strict mode"
set -x

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

BUILDKITD_TOML=buildkitd.toml
REGISTRY_NAME=k3d-registry
# K3D_NETWORK_NAME=k3d-kurtosis
REGISTRY_CONTAINER_NAME=$REGISTRY_NAME
REGISTRY_PREFIX=""
registry_internal_port=$(docker inspect $REGISTRY_CONTAINER_NAME | jq -r '.[0].Config.Labels."k3s.registry.port.internal"')
if [ $? -eq 0 ]; then
  REGISTRY_PREFIX="${REGISTRY_NAME}:${registry_internal_port}"
fi

#[registry."$REGISTRY_PREFIX"]
cat > "./$BUILDKITD_TOML" <<EOF
debug = true
trace = true
insecure-entitlements = [ "network.host", "security.insecure" ]

[registry."$REGISTRY_PREFIX"]
http = true
insecure = true

EOF


uname_arch=$(uname -m)
architecture="amd64"

if [ "$uname_arch" == "x86_64" ] || [ "$uname_arch" == "amd64" ]; then
    architecture="amd64"
elif [ "$uname_arch" == "aarch64" ] || [ "$uname_arch" == "arm64" ]; then
    architecture="arm64"
fi

#buildx_platform_arg='linux/arm64/v8,linux/amd64'
buildx_platform_arg='linux/amd64'
push_flag='--push'
echo "Building docker image for architecture '${buildx_platform_arg}' with flag '${push_flag}'"

# Build Docker image

## Start by making sure the builder and the context do not already exist. If that's the case remove them

## Check the network doesn't already exist
network_name="${REGISTRY_NAME}-network"
NETWORK_ALREADY_EXISTS=false
if docker network inspect "${network_name}" &>/dev/null; then
  echo "WARNING: network ${network_name} already exists, skipping creation. You may have to manually disconnect containers"
  echo " and remove the network with 'docker network rm ${network_name}' if this script fails"
  NETWORK_ALREADY_EXISTS=true
fi

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

## Create network
if [ "${NETWORK_ALREADY_EXISTS}" == "true" ]; then
  echo "Network ${network_name} already exists, skipping creation."
elif ! docker network create "${network_name}" &>/dev/null; then
  echo "Failed creating network ${network_name}. Try creating it manually with 'docker network create ${network_name}' before re-running this script"
  exit 1
fi

# If the network already exists, check if the containers are already connected and if not, connect them
if [ "${NETWORK_ALREADY_EXISTS}" == "true" ]; then
  network_containers=$(docker network inspect "${network_name}" --format='{{range .Containers}}{{printf "%s\n" .Name}}{{end}}')
  for container_name in ${REGISTRY_CONTAINER_NAME}; do
    if ! echo "${network_containers}" | grep -qw "${container_name}"; then
      echo "Container ${container_name} is not connected to network ${network_name}, connecting it now"
      docker network connect "${network_name}" "${container_name}"
      continue
    fi
    echo "Container ${container_name} is already connected to network ${network_name}"
  done
fi

# docker network connect "${network_name}" "${REGISTRY_CONTAINER_NAME}" &>/dev/null
# if [ "$?" != "0" ]; then
#   if
#   echo "WARNING: Failed connecting network ${network_name} to ${REGISTRY_CONTAINER_NAME}. Try connecting it manually with 'docker network connect ${network_name} ${REGISTRY_CONTAINER_NAME}' before re-running this script"
#   exit 1
# fi

## Create Docker context and buildx builder
if ! docker context create "${docker_buildx_context}" &>/dev/null; then
  echo "Error: Docker context creation for buildx failed" >&2
  exit 1
fi

echo "Using toml: ${BUILDKITD_TOML}"
cat ${BUILDKITD_TOML}
#if ! docker buildx create --config ./${BUILDKITD_TOML} --use --name "${kurtosis_docker_builder}" "${docker_buildx_context}" &>/dev/null; then
if ! docker buildx create --driver-opt network=${network_name} --config ./${BUILDKITD_TOML} --use --name "${kurtosis_docker_builder}" "${docker_buildx_context}" ; then
  echo "Error: Docker context switch for buildx failed" >&2d
  exit 1
fi

image_tags_concatenated="-t k3d-registry:5000/kurtosistech/files-artifacts-expander:f7984a-dirty"
dockerfile_filepath=/home/laith/code/randamu/workbench/kurtosis/core/files_artifacts_expander/Dockerfile
dockerfile_dirpath=/home/laith/code/randamu/workbench/kurtosis/core/files_artifacts_expander

## Actually build the Docker image
push_flag="--push"
docker_buildx_cmd="docker buildx build --allow network.host --allow security.insecure ${push_flag} --platform ${buildx_platform_arg} ${image_tags_concatenated} -f ${dockerfile_filepath} ${dockerfile_dirpath}"
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
