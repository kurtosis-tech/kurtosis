#!/usr/bin/env bash
# 2021-07-08 WATERMARK, DO NOT REMOVE - This script was generated from the Kurtosis Bash script template

set -euo pipefail   # Bash "strict mode"
script_dirpath="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"
root_dirpath="$(dirname "${script_dirpath}")"


# ==================================================================================================
#                                             Constants
# ==================================================================================================
PACKAGE_JSON_FILEPATH="typescript/package.json"
REPLACE_PATTERN="\"version\": \"%s\""

# ==================================================================================================
#                                       Arg Parsing & Validation
# ==================================================================================================
show_helptext_and_exit() {
    echo "Usage: $(basename "${0}") new_version"
    echo ""
    echo "  new_version         The new version that the package files should contain"
    echo ""
    exit 1  # Exit with an error so that if this is accidentally called by CI, the script will fail
}

new_version="${1:-}"

if [ -z "${new_version}" ]; then
    echo "Error: No new version provided" >&2
    show_helptext_and_exit
fi

# ==================================================================================================
#                                             Main Logic
# ==================================================================================================
to_update_abs_filepath="${root_dirpath}/${PACKAGE_JSON_FILEPATH}"
if ! $(kudet update-version-in-file "${to_update_abs_filepath}" "${REPLACE_PATTERN}" "${new_version}"); then
    echo "Error: An error occurred setting new version '${new_version}' in constants file '${constant_file_abs_filepath}' using pattern '${replace_pattern}'" >&2
    exit 1
fi