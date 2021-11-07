#!/usr/bin/env bash
# 2021-07-08 WATERMARK, DO NOT REMOVE - This script was generated from the Kurtosis Bash script template

set -euo pipefail   # Bash "strict mode"
script_dirpath="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"
root_dirpath="$(dirname "${script_dirpath}")"



# ==================================================================================================
#                                             Constants
# ==================================================================================================
RELEASE_SCRIPT_FILENAME="release-repo.sh"     # NOTE: Must be on the path; comes from devtools repo

UPDATE_OWN_VERSION_CONSTS_SCRIPT_FILENAME="update-own-version-constants.sh"
UPDATE_PACKAGE_VERSION_SCRIPT_FILENAME="update-package-versions.sh"


# ==================================================================================================
#                                             Main Logic
# ==================================================================================================
if ! bash "${RELEASE_SCRIPT_FILENAME}" "${root_dirpath}" "${script_dirpath}/${UPDATE_OWN_VERSION_CONSTS_SCRIPT_FILENAME}" "${script_dirpath}/${UPDATE_PACKAGE_VERSION_SCRIPT_FILENAME}"; then
    echo "Error: Couldn't cut the release" >&2
    exit 1
fi
