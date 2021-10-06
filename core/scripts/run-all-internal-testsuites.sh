#!/usr/bin/env bash
# 2021-07-08 WATERMARK, DO NOT REMOVE - This script was generated from the Kurtosis Bash script template

set -euo pipefail   # Bash "strict mode"
script_dirpath="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"
root_dirpath="$(dirname "${script_dirpath}")"



# ==================================================================================================
#                                             Constants
# ==================================================================================================
source "${script_dirpath}/_constants.sh"

SUPPORTED_LANGS_FILENAME="supported-languages.txt"
RUN_ONE_TESTSUITE_SCRIPT_FILENAME="run-one-internal-testsuite.sh"

# ==================================================================================================
#                                             Main Logic
# ==================================================================================================
if ! docker_tag="$("${script_dirpath}/${GET_FIXED_DOCKER_IMAGES_TAG_SCRIPT_FILENAME}")"; then
    echo "Error: An error occurred getting the Docker tag for the images produced by this repo" >&2
    exit 1
fi

supported_langs_filepath="${root_dirpath}/${SUPPORTED_LANGS_FILENAME}"
if ! [ -f "${supported_langs_filepath}" ]; then
    echo "Error: Expected supported languages file at '${supported_langs_filepath}', but none was found" >&2
    exit 1
fi

run_one_testsuite_script_filepath="${script_dirpath}/${RUN_ONE_TESTSUITE_SCRIPT_FILENAME}"
had_failures="false"
for lang in $(cat "${supported_langs_filepath}"); do
    echo "Running internal testsuite for lang '${lang}'..."
    # The funky ${1+"${@}"} incantation is how you you feed arguments exactly as-is to a child script in Bash
    # ${*} loses quoting and ${@} trips set -e if no arguments are passed, so this incantation says, "if and only if 
    #  ${1} exists, evaluate ${@}"
    if ! bash "${run_one_testsuite_script_filepath}" "${lang}" ${1+"${@}"}; then
        echo "Error: Internal testsuite for lang '${lang}' failed!" >&2
        had_failures="true"
    fi
    echo "Internal testsuite for lang '${lang}' succeeded"
done

if "${had_failures}"; then
    echo "Error: One or more testsuites failed" >&2
    exit 1
fi
echo "All testsuites completed successfully!"
