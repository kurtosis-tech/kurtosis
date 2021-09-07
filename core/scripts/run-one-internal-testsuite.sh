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
EXAMPLE_API_MICROSERVICE_IMAGE="${DOCKER_ORG}/example-microservices_api"
EXAMPLE_DATASTORE_MICROSERVICE_IMAGE="${DOCKER_ORG}/example-microservices_datastore"
INTERNAL_TESTSUITE_PARAMS_JSON='{
    "apiServiceImage" :"'${EXAMPLE_API_MICROSERVICE_IMAGE}'",
    "datastoreServiceImage": "'${EXAMPLE_DATASTORE_MICROSERVICE_IMAGE}'"
}'


# ==================================================================================================
#                                       Arg Parsing & Validation
# ==================================================================================================
show_helptext_and_exit() {
    echo "Usage: $(basename "${0}") testsuite_lang_to_run [optional kurtosis.sh args....]"
    echo ""
    echo "  testsuite_lang_to_run   The language of the testsuite to run"
    echo ""
    exit 1  # Exit with an error so that if this is accidentally called by CI, the script will fail
}

testsuite_lang="${1:-}"
if [ -z "${testsuite_lang}" ]; then
    echo "Error: A testsuite lang to run must be provided" >&2
    show_helptext_and_exit
fi
shift 1 # All other args should be passed as-is to the kurtosis.sh script

# ==================================================================================================
#                                             Main Logic
# ==================================================================================================
if ! docker_tag="$("${script_dirpath}/${GET_DOCKER_IMAGES_TAG_SCRIPT_FILENAME}")"; then
    echo "Error: An error occurred getting the Docker tag for the images produced by this repo" >&2
    exit 1
fi

supported_langs_filepath="${root_dirpath}/${SUPPORTED_LANGS_FILENAME}"
if ! [ -f "${supported_langs_filepath}" ]; then
    echo "Error: Expected supported languages file at '${supported_langs_filepath}', but none was found" >&2
    exit 1
fi

is_lang_valid="false"
for lang in $(cat "${supported_langs_filepath}"); do
    if [ "${lang}" == "${testsuite_lang}" ]; then
        is_lang_valid="true"
        break
    fi
done
if ! "${is_lang_valid}"; then
    echo "Error: Testsuite lang '${testsuite_lang}' doesn't correspond to any of the supported langauges in '${supported_langs_filepath}'" >&2
    exit 1
fi

testsuite_image="${DOCKER_ORG}/${testsuite_lang}-${INTERNAL_TESTSUITE_IMAGE_SUFFIX}:${docker_tag}"
wrapper_filepath="${root_dirpath}/${WRAPPER_OUTPUT_REL_FILEPATH}"
# The funky ${1+"${@}"} incantation is how you you feed arguments exactly as-is to a child script in Bash
# ${*} loses quoting and ${@} trips 'set -e' if no arguments are passed, so this incantation says, "if and only if 
#  ${1} exists, evaluate ${@}"
bash "${wrapper_filepath}" --custom-params "${INTERNAL_TESTSUITE_PARAMS_JSON}" ${1+"${@}"} "${testsuite_image}"
