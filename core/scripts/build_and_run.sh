set -euo pipefail
script_dirpath="$(cd "$(dirname "${BASH_SOURCE[0]}")"; pwd)"

# ====================== CONSTANTS =======================================================
DOCKER_ORG="kurtosistech"
REPO_BASE="kurtosis-core"
API_REPO="${REPO_BASE}_api"
INITIALIZER_REPO="${REPO_BASE}_initializer"
GO_EXAMPLE_SUITE_IMAGE="kurtosistech/kurtosis-go-example:develop"

# ====================== ARG PARSING =======================================================
show_help() {
    echo "${0}:"
    echo "  -h      Displays this message"
    echo "  -b      Executes only the build step, skipping the run step"
    echo "  -r      Executes only the run step, skipping the build step"
    echo "  -d      Extra args to pass to 'docker run' (e.g. '--env MYVAR=somevalue')"
}

do_build=true
do_run=true
extra_docker_args=""
while getopts "brd:" opt; do
    case "${opt}" in
        h)
            show_help
            exit 0
            ;;
        b)
            do_run=false
            ;;
        r)
            do_build=false
            ;;
        d)
            extra_docker_args="${OPTARG}"
            ;;
    esac
done

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
    echo "Build threads launched successfully"

    echo "Waiting for build threads to exit... (initializer PID: ${initializer_build_pid}, API PID: ${api_build_pid})"
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
    go_suite_execution_volume="go-example-suite_${docker_tag}_$(date +%s)"
    docker volume create "${go_suite_execution_volume}"
    docker run \
        --mount "type=bind,source=/var/run/docker.sock,target=/var/run/docker.sock" \
        --mount "type=volume,source=${go_suite_execution_volume},target=/suite-execution" \
        --env 'CUSTOM_ENV_VARS_JSON={"GO_EXAMPLE_SERVICE_IMAGE":"nginxdemos/hello"}' \
        --env "TEST_SUITE_IMAGE=${GO_EXAMPLE_SUITE_IMAGE}" \
        --env "KURTOSIS_API_IMAGE=${api_image}" \
        --env "SUITE_EXECUTION_VOLUME=${go_suite_execution_volume}" \
        ${extra_docker_args:-} \
        "${initializer_image}"
fi
