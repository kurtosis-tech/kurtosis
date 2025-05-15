#!/usr/bin/env bash
# 2021-07-08 WATERMARK, DO NOT REMOVE - This script was generated from the Kurtosis Bash script template

set -euo pipefail   # Bash "strict mode"
script_dirpath="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"
root_dirpath="$(dirname "${script_dirpath}")"

set +u
# Check if it's running inside a nix shell env so there is no need to check versions
if [ -z "${IN_NIX_SHELL}" ]; then
    if ! bash "${script_dirpath}/versions_check.sh"; then
    exit 1
    fi
fi
set -u

# ==================================================================================================
#                                             Constants
# ==================================================================================================
# These scripts will be executed always
MANDATORY_BUILD_SCRIPT_RELATIVE_FILEPATHS=(
  "scripts/go-lint-all.sh"
  "scripts/go-tidy-all.sh"
  "scripts/generate-kurtosis-version.sh"
  "cli/scripts/build.sh"
)

# for regular builds
BUILD_SCRIPT_RELATIVE_FILEPATHS=(
    "container-engine-lib/scripts/build.sh"
    "contexts-config-store/scripts/build.sh"
    "grpc-file-transfer/scripts/build.sh"
    "name_generator/scripts/build.sh"
    "api/scripts/build.sh"
    "metrics-library/scripts/build.sh"
    "enclave-manager/scripts/build.sh"
    "engine/scripts/build.sh"
    "core/scripts/build.sh"
)

# projects with debug mode enabled
BUILD_DEBUG_SCRIPT_RELATIVE_FILEPATHS=(
    "engine/scripts/build.sh"
    "core/scripts/build.sh"
)

DEFAULT_DEBUG_IMAGE="false"
DEFAULT_PODMAN_MODE="false"

# ==================================================================================================
#                                       Arg Parsing & Validation
# ==================================================================================================
show_helptext_and_exit() {
    echo "Usage: $(basename "${0}") debug_image podman_mode..."
    echo ""
    echo "  debug_image   Whether images should contains the debug server and run in debug mode, this will use the Dockerfile.debug image to build the container (configured for the engine server and the APIC server so far)"
    echo "  podman_mode   Whether images should be built with podman instead of docker. Use if you are developing Kurtosis on Podman cluster type"
    echo ""
    exit 1  # Exit with an error so that if this is accidentally called by CI, the script will fail
}

# Raw parsing of the arguments
debug_image="${1:-"${DEFAULT_DEBUG_IMAGE}"}"
if [ "${debug_image}" != "true" ] && [ "${debug_image}" != "false" ]; then
    echo "Error: Invalid debug_image arg: '${debug_image}'" >&2
    show_helptext_and_exit
fi

podman_mode="${2:-"${DEFAULT_PODMAN_MODE}"}"
if [ "${podman_mode}" != "true" ] && [ "${podman_mode}" != "false" ]; then
    echo "Error: Invalid podman_mode arg: '${podman_mode}'" >&2
    show_helptext_and_exit
fi

# ==================================================================================================
#                                             Main Logic
# ==================================================================================================
# run mandatory scripts first
for build_script_rel_filepath in "${MANDATORY_BUILD_SCRIPT_RELATIVE_FILEPATHS[@]}"; do
    build_script_abs_filepath="${root_dirpath}/${build_script_rel_filepath}"
    if ! bash "${build_script_abs_filepath}"; then
        echo "Error: Build script '${build_script_abs_filepath}' failed" >&2
        exit 1
    fi
done

# then run the remaining scripts
build_script_rel_filepaths=("${BUILD_SCRIPT_RELATIVE_FILEPATHS[@]}")
if "${debug_image}"; then
  build_script_rel_filepaths=("${BUILD_DEBUG_SCRIPT_RELATIVE_FILEPATHS[@]}")
fi

for build_script_rel_filepath in "${build_script_rel_filepaths[@]}"; do
    build_script_abs_filepath="${root_dirpath}/${build_script_rel_filepath}"
    if ! bash "${build_script_abs_filepath}" "${debug_image}" "${podman_mode}" ; then
        echo "Error: Build script '${build_script_abs_filepath}' failed" >&2
        exit 1
    fi
done
