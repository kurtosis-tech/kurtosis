set -euo pipefail
script_dirpath="$(cd "$(dirname "${BASH_SOURCE[0]}")"; pwd)"
root_dirpath="$(dirname "${script_dirpath}")"

# ====================== CONSTANTS =======================================================
DOCKER_ORG="kurtosistech"
REPO_BASE="kurtosis-core"
API_REPO="${REPO_BASE}_api"
INITIALIZER_REPO="${REPO_BASE}_initializer"
GO_EXAMPLE_SUITE_IMAGE="${DOCKER_ORG}/kurtosis-go-example:develop"
KURTOSIS_DIRPATH="$HOME/.kurtosis"

BUILD_DIRPATH="${root_dirpath}/build"
WRAPPER_GENERATOR_DIRPATH="${root_dirpath}/wrapper_generator"
WRAPPER_GENERATOR_FILEPATH="${BUILD_DIRPATH}/wrapper-generator"
WRAPPER_TEMPLATE_FILEPATH="${WRAPPER_GENERATOR_DIRPATH}/kurtosis.template.sh"
WRAPPER_FILEPATH="${BUILD_DIRPATH}/kurtosis.sh"


BUILD_ACTION="build"
RUN_ACTION="run"
BOTH_ACTION="all"
HELP_ACTION="help"

# ====================== ARG PARSING =======================================================
show_help() {
    echo "${0} <action> [<extra kurtosis.sh args...>]"
    echo ""
    echo "  This script will optionally build your Kurtosis testsuite into a Docker image and/or run it via a call to the kurtosis.sh wrapper script"
    echo ""
    echo "  To select behaviour, choose from the following actions:"
    echo "    help    Displays this messages"
    echo "    build   Executes only the build step, skipping the run step"
    echo "    run     Executes only the run step, skipping the build step"
    echo "    all     Executes both build and run steps"
    echo ""
    echo "  To see the flags the kurtosis.sh script accepts, add the '--help' flag"
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
# Captures the first of tag > branch > commit
git_ref="$(git describe --tags --exact-match 2> /dev/null || git symbolic-ref -q --short HEAD || git rev-parse --short HEAD)"
docker_tag="$(echo "${git_ref}" | sed 's,[/:],_,g')"

initializer_image="${DOCKER_ORG}/${INITIALIZER_REPO}:${docker_tag}"
api_image="${DOCKER_ORG}/${API_REPO}:${docker_tag}"

initializer_log_filepath="$(mktemp)"
api_log_filepath="$(mktemp)"
if "${do_build}"; then
    echo "Generating wrapper script..."
    mkdir -p "${BUILD_DIRPATH}"
    go build -o "${WRAPPER_GENERATOR_FILEPATH}" "${WRAPPER_GENERATOR_DIRPATH}/main.go"

    # If we're building a tag, then we need to generate the wrapper script to pull the 'X.Y' Docker tag (rather than X.Y.Z, since X.Y is the only
    #  tag Docker images get published with)
    kurtosis_core_version="${docker_tag}"
    if [[ "${git_ref}" =~ ^[0-9]+\.[0-9]+\.[0-9]+$ ]]; then
        kurtosis_core_version="$(echo "${git_ref}" | egrep -o '^[0-9]+\.[0-9]+')"
    fi
    "${WRAPPER_GENERATOR_FILEPATH}" -kurtosis-core-version "${kurtosis_core_version}" -template "${WRAPPER_TEMPLATE_FILEPATH}" -output "${WRAPPER_FILEPATH}"
    echo "Successfully generated wrapper script"

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
    # --------------------- Kurtosis Go environment variables ---------------------
    api_service_image="${DOCKER_ORG}/example-microservices_api"
    datastore_service_image="${DOCKER_ORG}/example-microservices_datastore"
    # Docker only allows you to have spaces in the variable if you escape them or use a Docker env file
    go_suite_env_vars_json='{
        "API_SERVICE_IMAGE" :"'${api_service_image}'",
        "DATASTORE_SERVICE_IMAGE": "'${datastore_service_image}'"
    }'
    # --------------------- End Kurtosis Go environment variables ---------------------

    # The generated wrapper will come hardcoded the correct version of the initializer/API images
    bash "${WRAPPER_FILEPATH}" --custom-env-vars "${go_suite_env_vars_json}" "${@}" "${GO_EXAMPLE_SUITE_IMAGE}"
fi
