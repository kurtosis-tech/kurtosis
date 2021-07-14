set -euo pipefail
script_dirpath="$(cd "$(dirname "${BASH_SOURCE[0]}")"; pwd)"
root_dirpath="$(dirname "${script_dirpath}")"

# ====================== CONSTANTS =======================================================
DOCKER_ORG="kurtosistech"
REPO_BASE="kurtosis-core"
API_REPO="${REPO_BASE}_api"
INITIALIZER_REPO="${REPO_BASE}_initializer"
# NOTE: We build against a specific version of the Golang testsuite (rather than an evergreen 'develop' version) to minimize the circular dependencies 
# going on (since Kurt Libs depends on this repo depends on Kurt Lib). 
# However, this *does* mean that we'll be testing Kurt Core against a probably-outdated version of Kurt Libs - we're alright doing this, 
# because some bit of sanity-checking is better than none and we'll test Kurt Core against the latest Kurt Libs when we upgrade the Core version in Libs
# TODO The ideal would be having an extensive testsuite, specific to this repo, that runs all the this-repos-specific tests so that we don't have to depend on Kurt Libs anymore
GO_EXAMPLE_SUITE_IMAGE="${DOCKER_ORG}/kurtosis-golang-example:1.28.0"
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
    echo "$(basename "${0}") <action> [<kurtosis.sh args...>]"
    echo ""
    echo "  This script will optionally a) generate a kurtosis.sh script + build your testsuite into a Docker image, and/or b) call down to the generated kurtosis.sh script to run the testsuite"
    echo ""
    echo "  To select this script's behaviour, choose from the following actions:"
    echo ""
    echo "    help    Displays this messages"
    echo "    build   Executes only the kurtosis.sh generation and Docker build steps, skipping the run step"
    echo "    run     Executes only the call to kurtosis.sh, skipping the build step"
    echo "    all     Executes both build and run steps"
    echo ""
    echo "  To see the args the kurtosis.sh script accepts for the 'run' phase, call '$(basename ${0}) all --help'"
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
docker_tag="$(echo "${git_ref}" | sed 's,[/:],_,g')"    # Sanitize git ref to be acceptable Docker tag format

# If we're building a tag of X.Y.Z, then we need to actually build the Docker images and generate the wrapper script with tag X.Y so that users will
#  get patch updates transparently
if [[ "${docker_tag}" =~ ^[0-9]+\.[0-9]+\.[0-9]+$ ]]; then
    docker_tag="$(echo "${docker_tag}" | egrep -o '^[0-9]+\.[0-9]+')"
fi

initializer_image="${DOCKER_ORG}/${INITIALIZER_REPO}:${docker_tag}"
api_image="${DOCKER_ORG}/${API_REPO}:${docker_tag}"

initializer_log_filepath="$(mktemp)"
api_log_filepath="$(mktemp)"
if "${do_build}"; then
    echo "Running tests..."
    if ! go test "${root_dirpath}/..."; then
        echo 'Tests failed!'
        exit 1
    fi
    echo "Tests completed"

    if ! [ -f "${root_dirpath}"/.dockerignore ]; then
        echo "Error: No .dockerignore file found in root; this is required so Docker caching works properly" >&2
        exit 1
    fi

    echo "Generating wrapper script..."
    mkdir -p "${BUILD_DIRPATH}"
    go build -o "${WRAPPER_GENERATOR_FILEPATH}" "${WRAPPER_GENERATOR_DIRPATH}/main.go"
    "${WRAPPER_GENERATOR_FILEPATH}" -kurtosis-core-version "${docker_tag}" -template "${WRAPPER_TEMPLATE_FILEPATH}" -output "${WRAPPER_FILEPATH}"
    echo "Successfully generated wrapper script"

    echo "Launching builds of initializer & API images in parallel threads..."
    docker build --progress=plain -t "${initializer_image}" -f "${root_dirpath}/initializer/Dockerfile" "${root_dirpath}" > "${initializer_log_filepath}" 2>&1 &
    initializer_build_pid="${!}"
    docker build --progress=plain -t "${api_image}" -f "${root_dirpath}/api_container/Dockerfile" "${root_dirpath}" > "${api_log_filepath}" 2>&1 &
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
    echo "Pulling latest version of example Go testsuite image..."
    if ! docker pull "${GO_EXAMPLE_SUITE_IMAGE}"; then
        echo "WARN: An error occurred pulling the latest version of the example Go testsuite image (${GO_EXAMPLE_SUITE_IMAGE}); you may be running an out-of-date version" >&2
    else
        echo "Successfully pulled latest version of example Go testsuite image"
    fi

    # --------------------- Kurtosis Go environment variables ---------------------
    api_service_image="${DOCKER_ORG}/example-microservices_api"
    datastore_service_image="${DOCKER_ORG}/example-microservices_datastore"
    go_suite_params_json='{
        "apiServiceImage" :"'${api_service_image}'",
        "datastoreServiceImage": "'${datastore_service_image}'",
        "isKurtosisCoreDevMode": true
    }'
    # --------------------- End Kurtosis Go environment variables ---------------------
    # The funky ${1+"${@}"} incantation is how you you feed arguments exactly as-is to a child script in Bash
    # ${*} loses quoting and ${@} trips set -e if no arguments are passed, so this incantation says, "if and only if 
    #  ${1} exists, evaluate ${@}"
    bash "${WRAPPER_FILEPATH}" --custom-params "${go_suite_params_json}" ${1+"${@}"} "${GO_EXAMPLE_SUITE_IMAGE}"
fi
