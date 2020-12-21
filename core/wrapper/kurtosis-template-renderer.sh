set -euo pipefail
script_dirpath="$(cd "$(dirname "${BASH_SOURCE[0]}")"; pwd)"

# ============================================================================================
#                                      Constants
# ============================================================================================
# The directory where Kurtosis will store files it uses in between executions, e.g. access tokens
KURTOSIS_DIRPATH="${HOME}/.kurtosis"

KURTOSIS_CORE_TAG="KURTOSISCOREVERSION"
KURTOSIS_DOCKERHUB_ORG="kurtosistech"
INITIALIZER_IMAGE="${KURTOSIS_DOCKERHUB_ORG}/kurtosis-core_initializer:${KURTOSIS_CORE_TAG}"
API_IMAGE="${KURTOSIS_DOCKERHUB_ORG}/kurtosis-core_api:${KURTOSIS_CORE_TAG}"

POSITIONAL_ARG_DEFINITION_FRAGMENTS=2

# ============================================================================================
#                                      Arg Parsing
# ============================================================================================
# Default values
testsuite_image=""

# Arg format - EITHER flag arg:
#   -flag1}}}--flag2}}}-flag3}}}-etc}}}variable_to_store_result_in}}}Helptext message}}}default_value
# OR positional arg:
#   variable_to_store_result_in}}}Helptext message
ARG_DEFINITIONS=(
    "-h}}}--help}}}show_help}}}false}}}false}}}Prints this help message"
    "-p}}}--parallelism}}}parallelism}}}4}}}false}}}Number of tests to run concurrently"
)


# Parse arg definitions into a) a show_help eval string and b) an arg-parsing while loop eval string
help_function_lines=()
for arg_definition in "${ARG_DEFINITIONS[@]}"; do
    arg_definition_delim_replaced="$(echo "${arg_definition}" | sed $'s/}}}/\\\n/g')"
    readarray -t fragments < <(echo -n "${arg_definition_delim_replaced}")

    flags=()
    fragments_without_flags=()
    for fragment in "${fragments[@]}"; do
        if [ "${fragment}" =~ ^-.* ]; then
            flags+=("${fragment}")
        else
            fragments_without_flags+=("${fragment}")
        fi
    done

    # Flag argument case
    if [ "${#flags[@]}" -gt 0 ]; then
        # TODO
        help_function_line=""
        for flag in "${flags[@]}"; do
    # Positional argument case
    else 
        if [ "${#fragments_without_flags[@]}" -ne "${POSITIONAL_ARG_DEFINITION_FRAGMENTS}" ]; then
            echo "Found positional argument definition that didn't have ${POSITIONAL_ARG_DEFINITION_FRAGMENTS} fragments, the required number for positional arguments: ${arg_definition}"
            exit 1
        fi
        arg_variable="${fragments_without_flags[0]}"
        helptext="${fragments_without_flags[1]}"

        help_function_line="${arg_variable}"$'\t'"${helptext}"
    fi


    help_function_lines+=("${help_function_line}")



    positional
    if [ "${#fragments_without_flags[@]}" -eq 2 ]; then
    elif [ "${#fragments_without_flags[@]}" -eq 3 ]; then
        # TODO
    else
        echo "Found argument definition ${arg_definition} that didn't match the length required for either flag or positional arguments" >&2
        exit 1
    fi
done

exit 99

show_help_eval_str='function show_help() { echo ""; echo $(basename ${0})

eval "function show_help() {



POSITIONAL=()
while [ ${#} -gt 0 ]; do
    key="${1}"
    case "${key}" in
        -h|--help)




# ============================================================================================
#                                    Arg Validation
# ============================================================================================
if [ -z "${testsuite_image}" ]; then
    echo "No testsuite image provided




BUILD_ACTION="build"
RUN_ACTION="run"
BOTH_ACTION="all"
HELP_ACTION="help"
show_help() {
    echo "${0} <action> [<extra 'docker run' args...>]"
    echo ""
    echo "  This script will optionally build your Kurtosis testsuite into a Docker image and/or run it via a call to 'docker run'"
    echo ""
    echo "  To select behaviour, choose from the following actions:"
    echo "    help    Displays this messages"
    echo "    build   Executes only the build step, skipping the run step"
    echo "    run     Executes only the run step, skipping the build step"
    echo "    all     Executes both build and run steps"
    echo ""
    echo "  To modify how your suite is run, you can set Kurtosis environment variables using the '--env' flag to 'docker run' like so:"
    echo "    ${0} all --env PARALLELISM=4"
    echo ""
    echo "  To see all the environment variables Kurtosis accepts, add the '--env SHOW_HELP=true' flag"
    echo ""
}

if [ "${#}" -eq 0 ]; then
    show_help
    exit 1     # Exit with error code so we dont't get spurious CI passes
fi

action="${1:-}"
shift 1

do_build=true
do_run=true
case "${action}" in
    ${HELP_ACTION})
        show_help
        exit 0
        ;;
    ${BUILD_ACTION})
        do_build=true
        do_run=false
        ;;
    ${RUN_ACTION})
        do_build=false
        do_run=true
        ;;
    ${BOTH_ACTION})
        do_build=true
        do_run=true
        ;;
    *)
        echo "Error: First argument must be one of '${HELP_ACTION}', '${BUILD_ACTION}', '${RUN_ACTION}', or '${BOTH_ACTION}'" >&2
        exit 1
        ;;
esac

# ====================== MAIN LOGIC =======================================================
git_branch="$(git rev-parse --abbrev-ref HEAD)"
docker_tag="$(echo "${git_branch}" | sed 's,[/:],_,g')"

root_dirpath="$(dirname "${script_dirpath}")"
if "${do_build}"; then
    echo "Running unit tests..."

    # TODO Extract this go-specific logic out into a separate script so we can copy/paste the build_and_run.sh between various languages
    if ! go test "${root_dirpath}/..."; then
        echo "Tests failed!"
        exit 1
    else
        echo "Tests succeeded"
    fi

    echo "Building ${SUITE_IMAGE} Docker image..."
    docker build -t "${SUITE_IMAGE}:${docker_tag}" -f "${root_dirpath}/testsuite/Dockerfile" "${root_dirpath}"
fi

if "${do_run}"; then
    # Kurtosis needs a Docker volume to store its execution data in
    # To learn more about volumes, see: https://docs.docker.com/storage/volumes/
    sanitized_image="$(echo "${SUITE_IMAGE}" | sed 's/[^a-zA-Z0-9_.-]/_/g')"
    suite_execution_volume="$(date +%Y-%m-%dT%H.%M.%S)_${sanitized_image}_${docker_tag}"
    docker volume create "${suite_execution_volume}"

    mkdir -p "${KURTOSIS_DIRPATH}"

    # ======================================= Custom Docker environment variables ========================================================
    # NOTE: Replace these with whatever custom properties your service needs
    api_service_image="${KURTOSIS_DOCKERHUB_ORG}/example-microservices_api"
    datastore_service_image="${KURTOSIS_DOCKERHUB_ORG}/example-microservices_datastore"
    # Docker only allows you to have spaces in the variable if you escape them or use a Docker env file
    custom_env_vars_json_flag="CUSTOM_ENV_VARS_JSON={\"API_SERVICE_IMAGE\":\"${api_service_image}\",\"DATASTORE_SERVICE_IMAGE\":\"${datastore_service_image}\"}"
    # ====================================== End custom Docker environment variables =====================================================

    docker run \
        `# The Kurtosis initializer runs inside a Docker container, but needs to access to the Docker engine; this is how to do it` \
        `# For more info, see the bottom of: http://jpetazzo.github.io/2015/09/03/do-not-use-docker-in-docker-for-ci/` \
        --mount "type=bind,source=/var/run/docker.sock,target=/var/run/docker.sock" \
        \
        `# Because the Kurtosis initializer runs inside Docker but needs to persist & read files on the host filesystem between execution,` \
        `#  the container expects the Kurtosis directory to be bind-mounted at the special "/kurtosis" path` \
        --mount "type=bind,source=${KURTOSIS_DIRPATH},target=/kurtosis" \
        \
        `# The Kurtosis initializer image requires the volume for storing suite execution data to be mounted at the special "/suite-execution" path` \
        --mount "type=volume,source=${suite_execution_volume},target=/suite-execution" \
        \
        `# A JSON map of custom environment variable bindings that should be set when running the testsuite container` \
        `# IMPORTANT: Docker only allows spaces here if they're backslash-escaped!` \
        --env "${custom_env_vars_json_flag}" \
        \
        `# Tell the initializer which test suite image to use` \
        --env "TEST_SUITE_IMAGE=${SUITE_IMAGE}:${docker_tag}" \
        \
        `# Tell the initializer the name of the volume to store data in, so it can mount it on new Docker containers it creates` \
        --env "SUITE_EXECUTION_VOLUME=${suite_execution_volume}" \
        \
        `# The initializer needs a special Kurtosis API image to operate` \
        `# The release channel here should match the release channel of the initializer itself` \
        --env "KURTOSIS_API_IMAGE=${API_IMAGE}" \
        \
        `# Extra Docker arguments that will be passed as-is to 'docker run'` \
        `# In Bash, this is how you feed arguments exactly as-is to a child script (since ${*} loses quoting and ${@} trips set -e if no arguments are passed)` \
        `# It basically says, "if and only if ${1} exists, evaluate ${@}"` \
        ${1+"${@}"} \
        \
        "${INITIALIZER_IMAGE}"
fi
