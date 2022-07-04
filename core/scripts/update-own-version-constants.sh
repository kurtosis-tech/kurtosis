#!/usr/bin/env bash

set -euo pipefail   # Bash "strict mode"
script_dirpath="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"
root_dirpath="$(dirname "${script_dirpath}")"



# ==================================================================================================
#                                             Constants
# ==================================================================================================
UPDATE_VERSION_IN_FILE_CMD_NAME="kudet update-version-in-file" # NOTE: kudet should be installed by every Kurtosis dev

API_DIRNAME="api"
API_SUPPORTED_LANGS_REL_FILEPATH="${API_DIRNAME}/supported-languages.txt"

# Relative to root of repo
declare -A REL_FILEPATH_UPDATE_PATTERNS
REL_FILEPATH_UPDATE_PATTERNS["api/golang/kurtosis_core_version/kurtosis_core_version.go"]="KurtosisCoreVersion = \"%s\""
REL_FILEPATH_UPDATE_PATTERNS["api/typescript/src/kurtosis_core_version/kurtosis_core_version.ts"]="KURTOSIS_CORE_VERSION: string = \"%s\""
REL_FILEPATH_UPDATE_PATTERNS["launcher/api_container_launcher/api_container_launcher.go"]="DefaultVersion = \"%s\""

# ==================================================================================================
#                                       Arg Parsing & Validation
# ==================================================================================================
show_helptext_and_exit() {
    echo "Usage: $(basename "${0}") new_version"
    echo ""
    echo "  new_version     The version of this repo that is about to be released"
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
echo "Updating own-version constants..."
for rel_filepath in "${!REL_FILEPATH_UPDATE_PATTERNS[@]}"; do
    replace_pattern="${REL_FILEPATH_UPDATE_PATTERNS["${rel_filepath}"]}"
    constant_file_abs_filepath="${root_dirpath}/${rel_filepath}"
    if ! $("${UPDATE_VERSION_IN_FILE_CMD_NAME}") "${constant_file_abs_filepath}" "${replace_pattern}" "${new_version}"; then
        echo "Error: An error occurred setting new version '${new_version}' in constants file '${constant_file_abs_filepath}' using pattern '${replace_pattern}'" >&2
        exit 1
    fi
    echo "Successfully updated '${constant_file_abs_filepath}'"
done
echo "Successfully updated all own-version constants"
