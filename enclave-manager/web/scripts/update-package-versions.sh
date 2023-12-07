#!/usr/bin/env bash
# 2021-07-08 WATERMARK, DO NOT REMOVE - This script was generated from the Kurtosis Bash script template

set -euo pipefail   # Bash "strict mode"
script_dirpath="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"
root_dirpath="$(dirname "${script_dirpath}")"


# ==================================================================================================
#                                             Constants
# ==================================================================================================


# ==================================================================================================
#                                       Arg Parsing & Validation
# ==================================================================================================
show_helptext_and_exit() {
    echo "Usage: $(basename "${0}") new_version"
    echo ""
    echo "  new_version         The new version that the lerna configuration should contain"
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
pushd "${root_dirpath}"
yarn install
yarn version ${new_version}
popd
to_update_abs_filepath="${root_dirpath}/${PACKAGE_JSON_FILEPATH}"
if ! sed -i -r "s/${REPLACE_PATTERN}/\1${new_version}\2/g" "${to_update_abs_filepath}"; then
    echo "Error: An error occurred setting new version '${new_version}' in constants file '${to_update_abs_filepath}' using pattern '${REPLACE_PATTERN}'" >&2
    exit 1
fi