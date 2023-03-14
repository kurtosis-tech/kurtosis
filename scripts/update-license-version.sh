#!/usr/bin/env bash
# 2021-07-08 WATERMARK, DO NOT REMOVE - This script was generated from the Kurtosis Bash script template

set -euo pipefail   # Bash "strict mode"
script_dirpath="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"
root_dirpath="$(dirname "${script_dirpath}")"


# ==================================================================================================
#                                             Constants
# ==================================================================================================
PACKAGE_JSON_FILEPATH="LICENSE.md"
REPLACE_PATTERN_VERSION="[0-9]+.[0-9]+.[0-9]+"
REPLACE_PATTERN_YEAR="[0-9]{4} (Kurtosis Technologies, Inc.)"
REPLACE_PATTERN_CHANGE_DATE="(Change Date:          ).+$"

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

current_year=$(date +"%Y")

license_change_date=$(date -d "+4 years" +"%Y-%m-%d")

# ==================================================================================================
#                                             Main Logic
# ==================================================================================================
to_update_abs_filepath="${root_dirpath}/${PACKAGE_JSON_FILEPATH}"
if ! sed -i -r "s/${REPLACE_PATTERN_VERSION}/${new_version}/g" "${to_update_abs_filepath}"; then
    echo "Error: An error occurred setting new version '${new_version}' in constants file '${constant_file_abs_filepath}' using pattern '${REPLACE_PATTERN_VERSION}'" >&2
    exit 1
fi

if ! sed -i -r "s/${REPLACE_PATTERN_YEAR}/${current_year} \1/g" "${to_update_abs_filepath}"; then
    echo "Error: An error occurred setting year '${current_year}' in constants file '${constant_file_abs_filepath}' using pattern '${REPLACE_PATTERN_YEAR}'" >&2
    exit 1
fi

if ! sed -i -r "s/${REPLACE_PATTERN_CHANGE_DATE}/\1${license_change_date}/g" "${to_update_abs_filepath}"; then
    echo "Error: An error occurred setting change date '${license_change_date}' in constants file '${constant_file_abs_filepath}' using pattern '${REPLACE_PATTERN_CHANGE_DATE}'" >&2
    exit 1
fi
