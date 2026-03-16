#!/usr/bin/env bash

set -euo pipefail

script_dirpath="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"
root_dirpath="$(dirname "${script_dirpath}")"
cli_dirpath="${root_dirpath}/cli/cli"
output_binary="${root_dirpath}/.tmp/kurtosis"
install_path="${INSTALL_PATH:-/usr/local/bin/kurtosis}"
build_images="${BUILD_IMAGES:-true}"
debug_image="${DEBUG_IMAGE:-false}"
podman_mode="${PODMAN_MODE:-false}"
reset_kurtosis="${RESET_KURTOSIS_AFTER_BUILD:-true}"

mkdir -p "$(dirname "${output_binary}")"

echo "Generating Kurtosis version constants..."
"${script_dirpath}/generate-kurtosis-version.sh"

if [[ "${build_images}" == "true" ]]; then
  echo "Building matching Kurtosis engine Docker image..."
  "${root_dirpath}/engine/scripts/build.sh" "${debug_image}" "${podman_mode}"

  echo "Building matching Kurtosis core Docker image..."
  "${root_dirpath}/core/scripts/build.sh" "${debug_image}" "${podman_mode}"
fi

echo "Building Kurtosis CLI with go build..."
(
  cd "${cli_dirpath}"
  go build -o "${output_binary}" .
)

echo "Installing Kurtosis CLI to '${install_path}'..."
cp "${output_binary}" "${install_path}"

if [[ "${reset_kurtosis}" == "true" ]]; then
  echo "Stopping Kurtosis engine..."
  "${install_path}" engine stop || true

  echo "Removing Kurtosis-managed Docker containers..."
  docker ps -aq --filter "label=com.kurtosistech.app-id=kurtosis" | xargs -r docker rm -f

  echo "Removing Kurtosis infrastructure containers by name..."
  docker ps -aq --filter "name=kurtosis-engine" | xargs -r docker rm -f
  docker ps -aq --filter "name=kurtosis-reverse-proxy" | xargs -r docker rm -f
  docker ps -aq --filter "name=kurtosis-logs-collector" | xargs -r docker rm -f
  docker ps -aq --filter "name=kurtosis-api--" | xargs -r docker rm -f
fi

echo "Build complete."
echo "Installed binary: ${install_path}"
