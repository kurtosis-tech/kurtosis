#!/usr/bin/env bash

set -euo pipefail   # Bash "strict mode"
script_dirpath="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"
root_dirpath="$(dirname "${script_dirpath}")"



# ==================================================================================================
#                                             Constants
# ==================================================================================================
API_DIRNAME="api"
API_SUPPORTED_LANGS_REL_FILEPATH="${API_DIRNAME}/supported-languages.txt"

# Relative to root of repo
declare -A REL_FILEPATH_UPDATE_PATTERNS
REL_FILEPATH_UPDATE_PATTERNS["launcher/api_container_launcher/api_container_launcher.go"]="defaultImageVersionTag = \"%s\""
REL_FILEPATH_UPDATE_PATTERNS["${API_DIRNAME}/golang/lib/kurtosis_api_version_const/kurtosis_api_version_const.go"]="KurtosisApiVersion = \"%s\""
REL_FILEPATH_UPDATE_PATTERNS["${API_DIRNAME}/typescript/src/lib/kurtosis_api_version_const.ts"]="KURTOSIS_API_VERSION: string = \"%s\""


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
# Verify that we're updating an own-version constant in every API directory
api_langs_updated_filepath="$(mktemp)"
for rel_filepath in "${!REL_FILEPATH_UPDATE_PATTERNS[@]}"; do
    root_dirname="$(echo "${rel_filepath}" | cut -d'/' -f1)"
    if [ "${root_dirname}" != "${API_DIRNAME}" ]; then
        continue
    fi

    lang_dirname="$(echo "${rel_filepath}" | cut -d'/' -f2)"
    echo "${lang_dirname}" >> "${api_langs_updated_filepath}"
done
api_supported_langs_abs_filepath="${root_dirpath}/${API_SUPPORTED_LANGS_REL_FILEPATH}"
if ! [ -f "${api_supported_langs_abs_filepath}" ]; then
    echo "Error: No API supported-languages file found at '${api_supported_langs_abs_filepath}', which is necessary for verifying we've updated every language's constant" >&2
    exit 1
fi
if ! langs_with_unupdated_consts="$(diff <(sort "${api_langs_updated_filepath}") <(sort "${root_dirpath}/${API_SUPPORTED_LANGS_REL_FILEPATH}"))"; then
    echo "Error: Couldn't generate a list of langs with unupdated own-version constants" >&2
    exit 1
fi
if [ -n "${langs_with_unupdated_consts}" ]; then
    echo "Error: The following languages don't have an own-version constant file getting updated in this script; this script needs to be updated " >&2
    echo "${langs_with_unupdated_consts}" >&2
    exit 1
fi

echo "Updating own-version constants..."
for rel_filepath in "${!REL_FILEPATH_UPDATE_PATTERNS[@]}"; do
    replace_pattern="${!REL_FILEPATH_UPDATE_PATTERNS["${rel_filepath}"]}"
    constant_file_abs_filepath="${root_dirpath}/${rel_filepath}"
    if ! "${UPDATE_VERSION_IN_FILE_SCRIPT_FILENAME}" "${constant_file_abs_filepath}" "${replace_pattern}" "${new_version}"; then
        echo "Error: An error occurred setting new version '${new_version}' in constants file '${constant_file_abs_filepath}' using pattern '${replace_pattern}'" >&2
        exit 1
    fi
    echo "Successfully updated '${constant_file_abs_filepath}'"
done
echo "Successfully updated all own-version constants"
