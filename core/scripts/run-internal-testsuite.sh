#!/usr/bin/env bash
# 2021-07-08 WATERMARK, DO NOT REMOVE - This script was generated from the Kurtosis Bash script template

set -euo pipefail   # Bash "strict mode"
script_dirpath="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"
root_dirpath="$(dirname "${script_dirpath}")"



# ==================================================================================================
#                                             Constants
# ==================================================================================================
source "${script_dirpath}/_constants.sh"

EXAMPLE_API_MICROSERVICE_IMAGE="${DOCKER_ORG}/example-microservices_api"
EXAMPLE_DATASTORE_MICROSERVICE_IMAGE="${DOCKER_ORG}/example-microservices_datastore"
INTERNAL_TESTSUITE_PARAMS_JSON='{
    "apiServiceImage" :"'${EXAMPLE_API_MICROSERVICE_IMAGE}'",
    "datastoreServiceImage": "'${EXAMPLE_DATASTORE_MICROSERVICE_IMAGE}'"
}'



# ==================================================================================================
#                                             Main Logic
# ==================================================================================================
if ! docker_tag="$("${script_dirpath}/${GET_DOCKER_IMAGES_TAG_SCRIPT_FILENAME}")"; then
    echo "Error: An error occurred getting the Docker tag for the images produced by this repo" >&2
    exit 1
fi

internal_testsuite_image="${DOCKER_ORG}/${INTERNAL_TESTSUITE_REPO}:${docker_tag}"
wrapper_filepath="${root_dirpath}/${WRAPPER_OUTPUT_REL_FILEPATH}"

# The funky ${1+"${@}"} incantation is how you you feed arguments exactly as-is to a child script in Bash
# ${*} loses quoting and ${@} trips set -e if no arguments are passed, so this incantation says, "if and only if 
#  ${1} exists, evaluate ${@}"
bash "${wrapper_filepath}" --custom-params "${INTERNAL_TESTSUITE_PARAMS_JSON}" ${1+"${@}"} "${internal_testsuite_image}"
