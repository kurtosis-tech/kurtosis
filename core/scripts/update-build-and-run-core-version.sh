#!/usr/bin/env bash
# 2021-07-08 WATERMARK, DO NOT REMOVE - This script was generated from the Kurtosis Bash script template

# This script is a script that is run during release which will update the version of Kurtosis stored in the build-and-run-core.sh script

set -euo pipefail   # Bash "strict mode"
script_dirpath="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"
root_dirpath="$(dirname "${script_dirpath}")"


# ==================================================================================================
#                                             Constants
# ==================================================================================================
GET_AUTOUPDATING_DOCKER_IMAGE_TAG_SCRIPT_FILENAME="get-autoupdating-docker-images-tag.sh"
TARGET_SCRIPT_RELATIVE_FILEPATH="testsuite_scripts/build-and-run-core.sh"

# Devtool for updating versions in a file (expected to be on the PATH)
UPDATE_VERSION_DEVTOOL_FILENAME="update-version-in-file.sh"

REPLACE_PATTERN="API_CONTAINER_VERSION=\"%s\""

# ==================================================================================================
#                                       Arg Parsing & Validation
# ==================================================================================================
show_helptext_and_exit() {
    echo "Usage: $(basename "${0}") new_kurtosis_version"
    echo ""
    echo "  new_kurtosis_version    The version of Kurtosis that should be stored in the build-and-run script"
    echo ""
    exit 1  # Exit with an error so that if this is accidentally called by CI, the script will fail
}

new_version="${1:-}"
if [ -z "${new_version}" ]; then
    echo "Error: no new version arg provided" >&2
    show_helptext_and_exit
fi



# ==================================================================================================
#                                             Main Logic
# ==================================================================================================
to_update_filepath="${root_dirpath}/${TARGET_SCRIPT_RELATIVE_FILEPATH}"
if ! bash "${UPDATE_VERSION_DEVTOOL_FILENAME}" "${to_update_filepath}" "${REPLACE_PATTERN}" "${new_version}"; then
    echo "Error: Couldn't replace the version in build-and-run file '${to_update_filepath}' to version '${new_version}' using pattern '${REPLACE_PATTERN}'" >&2
    exit 1
fi
