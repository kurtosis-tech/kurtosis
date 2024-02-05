#!/usr/bin/env bash
# 2021-07-08 WATERMARK, DO NOT REMOVE - This script was generated from the Kurtosis Bash script template

set -euo pipefail   # Bash "strict mode"
script_dirpath="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"
root_dirpath="$(dirname "${script_dirpath}")"



# ==================================================================================================
#                                             Constants
# ==================================================================================================
DEFAULT_DEBUG_IMAGE=false
SET_ARCH_AUTOMATICALLY=""
DEFAULT_SKIP_DOCKER_IMAGE_BUILDING=false

BUILD_SCRIPT_RELATIVE_FILEPATHS=(
    "launcher/scripts/build.sh"
    "server/scripts/build.sh"
)

# ==================================================================================================
#                                       Arg Parsing & Validation
# ==================================================================================================
show_helptext_and_exit() {
    echo "Usage: $(basename "${0}") debug_image..."
    echo ""
    echo "  debug_image   Whether images should contains the debug server and run in debug mode, this will use the Dockerfile.debug image to build the container (configured for the engine server so far)"
    echo ""
    exit 1  # Exit with an error so that if this is accidentally called by CI, the script will fail
}

# Raw parsing of the arguments
debug_image="${1:-"${DEFAULT_DEBUG_IMAGE}"}"
if [ "${debug_image}" != "true" ] && [ "${debug_image}" != "false" ]; then
    echo "Error: Invalid debug_image arg: '${debug_image}'" >&2
    show_helptext_and_exit
fi

# ==================================================================================================
#                                             Main Logic
# ==================================================================================================
for build_script_rel_filepath in "${BUILD_SCRIPT_RELATIVE_FILEPATHS[@]}"; do
    build_script_abs_filepath="${root_dirpath}/${build_script_rel_filepath}"
    if ! bash "${build_script_abs_filepath}" "${DEFAULT_SKIP_DOCKER_IMAGE_BUILDING}" "${SET_ARCH_AUTOMATICALLY}" "${debug_image}"; then
        echo "Error: Build script '${build_script_abs_filepath}' failed" >&2
        exit 1
    fi
done
