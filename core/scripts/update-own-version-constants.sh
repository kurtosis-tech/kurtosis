#!/usr/bin/env bash

set -euo pipefail   # Bash "strict mode"
script_dirpath="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"
root_dirpath="$(dirname "${script_dirpath}")"



# ==================================================================================================
#                                             Constants
# ==================================================================================================
API_DIRNAME="api"
SUPPORTED_LANGS_FILENAME="supported-languages.txt"
UPDATE_VERSION_IN_FILE_SCRIPT_FILENAME="update-version-in-file.sh" # From devtools; expected to be on PATH

# Per-language filepath that needs updating to update the constant, RELATIVE TO THE LANGUAGE ROOT
declare -A CONSTANT_FILE_RELATIVE_FILEPATHS

# Per-language patterns matching the constant line, which will be used for updating the version
declare -A CONSTANT_PATTERNS

# Golang
CONSTANT_FILE_RELATIVE_FILEPATHS["golang"]="lib/kurtosis_api_version_const/kurtosis_api_version_const.go"
CONSTANT_PATTERNS["golang"]="KurtosisApiVersion = \"%s\""

# Typescript
CONSTANT_FILE_RELATIVE_FILEPATHS["typescript"]="src/lib/kurtosis_api_version_const.ts"
CONSTANT_PATTERNS["typescript"]="KURTOSIS_API_VERSION: string = \"%s\""


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
echo "Updating the constants containing this library's version for all supported languages..."
api_dirpath="${root_dirpath}/${API_DIRNAME}"
supported_langs_filepath="${api_dirpath}/${SUPPORTED_LANGS_FILENAME}"
for lang in $(cat "${supported_langs_filepath}"); do
    constant_file_rel_filepath="${CONSTANT_FILE_RELATIVE_FILEPATHS["${lang}"]}"
    if [ -z "${constant_file_rel_filepath}" ]; then
        echo "Error: No relative filepath to a constant file that needs replacing was found for language '${lang}'; this script needs to be updated with this information" >&2
        exit 1
    fi

    constant_file_abs_filepath="${api_dirpath}/${lang}/${constant_file_rel_filepath}"

    pattern="${CONSTANT_PATTERNS["${lang}"]}"
    if [ -z "${pattern}" ]; then
        echo "Error: No replacement pattern was found for language '${lang}'; this script needs to be updated with this information" >&2
        exit 1
    fi

    if ! "${UPDATE_VERSION_IN_FILE_SCRIPT_FILENAME}" "${constant_file_abs_filepath}" "${pattern}" "${new_version}"; then
        echo "Error: An error occurred setting new version '${new_version}' in '${lang}' constants file '${to_update_abs_filepath}' using pattern '${replacement_pattern}'" >&2
        exit 1
    fi
done
echo "Successfully updated the constants containing this library's version have been updated for all supported languages"
