#!/usr/bin/env bash
# 2021-07-08 WATERMARK, DO NOT REMOVE - This script was generated from the Kurtosis Bash script template

set -euo pipefail   # Bash "strict mode"
script_dirpath="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"
root_dirpath="$(dirname "${script_dirpath}")"


# ==================================================================================================
#                                             Constants
# ==================================================================================================
SUPPORTED_LANGS_FILENAME="supported-languages.txt"

RUN_ONE_TESTSUITE_SCRIPT_REL_FILEPATH="scripts/run-one-internal-testsuite.sh"

# If a lang has this value for its error keyword, there is no error keyword to be searched for 
NO_ERR_KEYWORD_KEY="NONE"

# This contains a mapping of supported_lang -> keyword_in_testsuite_output_indicating_an_error
declare -A PER_LANG_ERROR_KEYWORD
PER_LANG_ERROR_KEYWORD["golang"]="ERRO"

DOCKER_CONTAINER_LS_CONTAINER_ID_COL_HEADER="CONTAINER ID"

KURTOSIS_ENGINE_CONTAINER_NAME_FRAGMENT="kurtosis-engine"

# ==================================================================================================
#                                       Arg Parsing & Validation
# ==================================================================================================
show_helptext_and_exit() {
    echo "Usage: $(basename "${0}") kurtosis_client_id kurtosis_client_secret"
    echo ""
    echo "  kurtosis_client_id     The Kurtosis client ID that will be used when running the testsuites"
    echo "  kurtosis_client_secret The Kurtosis client secret that will be used when running the testsuites"
    echo ""
    exit 1  # Exit with an error so that if this is accidentally called by CI, the script will fail
}

# NOTE: As of 2021-11-15, this is actually unused!!!!
client_id="${1:-}"
client_secret_DO_NOT_EVER_LOG="${2:-}"

if [ -z "${client_id}" ]; then
    echo "Error: no client ID arg provided" >&2
    show_helptext_and_exit
fi
if [ -z "${client_secret_DO_NOT_EVER_LOG}" ]; then
    echo "Error: no client secret arg provided" >&2
    show_helptext_and_exit
fi

# ==================================================================================================
#                                             Main Logic
# ==================================================================================================
supported_langs_abs_filepath="${root_dirpath}/${SUPPORTED_LANGS_FILENAME}"
run_one_testsuite_script_abs_filepath="${root_dirpath}/${RUN_ONE_TESTSUITE_SCRIPT_REL_FILEPATH}"

had_failures="false"
for lang in $(cat "${supported_langs_abs_filepath}"); do
    echo "Validating '${lang}' testsuite..."
    if ! output_filepath="$(mktemp)"; then
        echo "Error: Couldn't create a temp file to store the output of the '${lang}' testsuite in" >&2
        had_failures="true"
        continue
    fi

    bash "${run_one_testsuite_script_abs_filepath}" "${lang}" --client-id "${client_id}" --client-secret "${client_secret_DO_NOT_EVER_LOG}" | tee "${output_filepath}"

    if ! [ -v "PER_LANG_ERROR_KEYWORD[${lang}]" ]; then
        echo "Error: Language '${lang}' doesn't have an error keyword defined in this script; this script needs to be updated" >&2
        had_failures="true"
        continue
    fi
    err_keyword="${PER_LANG_ERROR_KEYWORD["${lang}"]}"

      # Grep exits with 0 if one or more lines match, so we fail the build if the err keyword is detected
      # This helps us catch errors that might show up in the testsuite logs but not get propagated to the actual exit codes
    if grep "${err_keyword}" "${output_filepath}"; then
        echo "Error: Detected error keyword '${err_keyword}' in '${lang}' testsuite output logfile" >&2
        had_failures="true"
        continue
    fi

    # Finally, verify that no containers besides the Kurtosis engine were left hanging around (i.e. that Kurtosis cleans up after itself)
    if [ -n "$(docker container ls | grep -v "${DOCKER_CONTAINER_LS_CONTAINER_ID_COL_HEADER}" | grep -v "${KURTOSIS_ENGINE_CONTAINER_NAME_FRAGMENT}")" ]; then
        echo "Error: Kurtosis left one or more containers hanging around after executing the '${lang}' testsuite; this is a Kurtosis bug!" >&2
        had_failures="true"
        continue
    fi

    echo "Successfully validated '${lang}' testsuite"
done

if "${had_failures}"; then
    echo "Error: One or more testsuites had failures" >&2
    exit 1
fi
