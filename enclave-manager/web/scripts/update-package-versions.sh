#!/usr/bin/env bash
# 2021-07-08 WATERMARK, DO NOT REMOVE - This script was generated from the Kurtosis Bash script template

set -euo pipefail   # Bash "strict mode"
script_dirpath="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"
root_dirpath="$(dirname "${script_dirpath}")"


# ==================================================================================================
#                                             Constants
# ==================================================================================================
LERNA_JSON_FILEPATH="lerna.json"
PACKAGE_JSON_FILEPATH="package.json"
PACKAGES_PACKAGE_JSON_FILEPATH="packages/*/package.json"
REPLACE_PATTERN="(\"version\": \")[0-9]+.[0-9]+.[0-9]+(\")"
REPLACE_DEP_PATTERN="(\"kurtosis.*?\": \")[0-9]+.[0-9]+.[0-9]+(\")"

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
to_update_abs_filepath="${root_dirpath}/${PACKAGE_JSON_FILEPATH}"
if ! sed -i -r "s/${REPLACE_PATTERN}/\1${new_version}\2/g" "${to_update_abs_filepath}"; then
    echo "Error: An error occurred setting new version '${new_version}' in root package.json file '${to_update_abs_filepath}' using pattern '${REPLACE_PATTERN}'" >&2
    exit 1
fi

to_update_abs_filepath="${root_dirpath}/${LERNA_JSON_FILEPATH}"
if ! sed -i -r "s/${REPLACE_PATTERN}/\1${new_version}\2/g" "${to_update_abs_filepath}"; then
    echo "Error: An error occurred setting new version '${new_version}' in lerna file '${to_update_abs_filepath}' using pattern '${REPLACE_PATTERN}'" >&2
    exit 1
fi

for to_update_rel_filepath in `ls ${root_dirpath}/${PACKAGES_PACKAGE_JSON_FILEPATH}`; do
  if ! sed -i -r "s/${REPLACE_PATTERN}/\1${new_version}\2/g" "${to_update_rel_filepath}"; then
      echo "Error: An error occurred setting new version '${new_version}' in package package.json file '${to_update_abs_filepath}' using pattern '${REPLACE_PATTERN}'" >&2
      exit 1
  fi
  if ! sed -i -r "s/${REPLACE_DEP_PATTERN}/\1${new_version}\2/g" "${to_update_rel_filepath}"; then
      echo "Error: An error occurred setting new version '${new_version}' in package package.json file '${to_update_abs_filepath}' using pattern '${REPLACE_PATTERN}'" >&2
      exit 1
  fi
done