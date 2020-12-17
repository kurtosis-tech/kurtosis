set -euo pipefail
script_dirpath="$(cd "$(dirname "${BASH_SOURCE[0]}")"; pwd)"

# ====================== CONSTANTS =======================================================
DOCKER_ORG="kurtosistech"
REPO_BASE="kurtosis-core"
API_REPO="${REPO_BASE}_api"
INITIALIZER_REPO="${REPO_BASE}_initializer"
GO_EXAMPLE_SUITE_IMAGE="kurtosistech/kurtosis-go-example:develop"
KURTOSIS_DIRPATH="$HOME/.kurtosis"

BUILD_ACTION="build"
RUN_ACTION="run"
BOTH_ACTION="all"
HELP_ACTION="help"

# ====================== ARG PARSING =======================================================
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
    exit 1   # Exit as error so we don't get spurious passes in CI
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

initializer_image="${DOCKER_ORG}/${INITIALIZER_REPO}:${docker_tag}"
api_image="${DOCKER_ORG}/${API_REPO}:${docker_tag}"

initializer_log_filepath="$(mktemp)"
api_log_filepath="$(mktemp)"
if "${do_build}"; then
    echo "Launching builds of initializer & API images in parallel threads..."
    docker build -t "${initializer_image}" -f "${root_dirpath}/initializer/Dockerfile" "${root_dirpath}" 2>&1 > "${initializer_log_filepath}" &
    initializer_build_pid="${!}"
    docker build -t "${api_image}" -f "${root_dirpath}/api_container/Dockerfile" "${root_dirpath}" 2>&1 > "${api_log_filepath}" &
    api_build_pid="${!}"
    echo "Build threads launched successfully:"
    echo " - Initializer thread PID: ${initializer_build_pid}"
    echo " - Initializer logs: ${initializer_log_filepath}"
    echo " - API thread PID: ${api_build_pid}"
    echo " - API logs: ${api_log_filepath}"

    echo "Waiting for build threads to exit..."
    builds_succeeded=true
    if ! wait "${initializer_build_pid}"; then
        builds_succeeded=false
    fi
    if ! wait "${api_build_pid}"; then
        builds_succeeded=false
    fi
    echo "Build threads exited"

    echo ""
    echo "===================== Initializer Image Build Logs =========================="
    cat "${initializer_log_filepath}"

    echo ""
    echo "========================= API Image Build Logs =============================="
    cat "${api_log_filepath}"

    echo ""
    if ! "${builds_succeeded}"; then
        echo "Build FAILED"
        exit 1
    fi
    echo "Build SUCCEEDED"
fi

if "${do_run}"; then
    mkdir -p "${KURTOSIS_DIRPATH}"
    go_suite_execution_volume="go-example-suite_${docker_tag}_$(date +%s)"
    docker volume create "${go_suite_execution_volume}"

    # --------------------- Kurtosis Go environment variables ---------------------
    api_service_image="${KURTOSIS_DOCKERHUB_ORG}/example-microservices_api"
    datastore_service_image="${KURTOSIS_DOCKERHUB_ORG}/example-microservices_datastore"
    # Docker only allows you to have spaces in the variable if you escape them or use a Docker env file
    go_suite_env_vars_json="CUSTOM_ENV_VARS_JSON={\"API_SERVICE_IMAGE\":\"${api_service_image}\",\"DATASTORE_SERVICE_IMAGE\":\"${datastore_service_image}\"}"
    # --------------------- End Kurtosis Go environment variables ---------------------

    docker run \
        --mount "type=bind,source=/var/run/docker.sock,target=/var/run/docker.sock" \
        --mount "type=bind,source=${KURTOSIS_DIRPATH},target=/kurtosis" \
        --mount "type=volume,source=${go_suite_execution_volume},target=/suite-execution" \
        --env "CUSTOM_ENV_VARS_JSON=${go_suite_env_vars_json}" \
        --env "TEST_SUITE_IMAGE=${GO_EXAMPLE_SUITE_IMAGE}" \
        --env "KURTOSIS_API_IMAGE=${api_image}" \
        --env "SUITE_EXECUTION_VOLUME=${go_suite_execution_volume}" \
        `# In Bash, this is how you feed arguments exactly as-is to a child script (since ${*} loses quoting and ${@} trips set -e if no arguments are passed)` \
        `# It basically says, "if and only if ${1} exists, evaluate ${@}"` \
        ${1+"${@}"} \
        "${initializer_image}"
fi
