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

GO_MODULE="github.com/kurtosis-tech/kurtosis-core"

API_CONTAINER_DIRNAME="api_container"
API_CONTAINER_API_DIRNAME="api"
API_CONTAINER_API_BINDINGS_DIRNAME="bindings"
API_CONTAINER_API_BINDINGS_GO_PKG="${GO_MODULE}/${API_CONTAINER_DIRNAME}/${API_CONTAINER_API_DIRNAME}/${API_CONTAINER_API_BINDINGS_DIRNAME}"

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
    # NOTE: When multiple people start developing on this, we won't be able to rely on using the user's local protoc because they might differ. We'll need to standardize by:
    #  1) Using protoc inside the API container Dockerfile to generate the output Go files (standardizes the output files for Docker)
    #  2) Using the user's protoc to generate the output Go files on the local machine, so their IDEs will work
    #  3) Tying the protoc inside the Dockerfile and the protoc on the user's machine together using a protoc version check
    #  4) Adding the locally-generated Go output files to .gitignore
    #  5) Adding the locally-generated Go output files to .dockerignore (since they'll get generated inside Docker)
    echo "Generating API container code from protobufs..."
    api_container_dirpath="${root_dirpath}/${API_CONTAINER_DIRNAME}"
    api_protobufs_input_dirpath="${api_container_dirpath}/${API_CONTAINER_API_DIRNAME}"
    api_protobufs_output_dirpath="${api_protobufs_input_dirpath}/${API_CONTAINER_API_BINDINGS_DIRNAME}"
    if [ "${api_protobufs_output_dirpath}/" != "/" ]; then
        if ! find ${api_protobufs_output_dirpath} -name '*.go' -delete; then
            echo "Error: An error occurred removing the protobuf-generated code" >&2
            exit 1
        fi
    else
        echo "Error: lib core generated API code dirpath must not be empty!" >&2
        exit 1
    fi
    for protobuf_filepath in $(find "${api_protobufs_input_dirpath}" -name "*.proto"); do
        protobuf_filename="$(basename "${protobuf_filepath}")"
        if ! protoc \
                -I="${api_protobufs_input_dirpath}" \
                --go_out="plugins=grpc:${api_protobufs_output_dirpath}" \
                `# Rather than specify the go_package in source code (which means all consumers of these protobufs would get it),` \
                `#  we specify the go_package here per https://developers.google.com/protocol-buffers/docs/reference/go-generated` \
                `# See also: https://github.com/golang/protobuf/issues/1272` \
                --go_opt="M${protobuf_filename}=${API_CONTAINER_API_BINDINGS_GO_PKG};$(basename "${API_CONTAINER_API_BINDINGS_GO_PKG}")" \
                "${protobuf_filepath}"; then
            echo "Error: An error occurred generating lib core files from protobuf file: ${protobuf_filepath}" >&2
            exit 1
        fi
    done
    echo "Successfully generated API container code from protobufs"

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
