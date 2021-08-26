#!/usr/bin/env bash
# 2021-07-08 WATERMARK, DO NOT REMOVE - This script was generated from the Kurtosis Bash script template

set -euo pipefail   # Bash "strict mode"
script_dirpath="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"
root_dirpath="$(dirname "${script_dirpath}")"


# ==================================================================================================
#                                             Constants
# ==================================================================================================
source "${script_dirpath}/_constants.sh"

DEFAULT_DOCKER_BUILD_ARGS="--progress=plain"

# ==================================================================================================
#                                             Main Logic
# ==================================================================================================
if ! docker_tag="$("${script_dirpath}/${GET_DOCKER_IMAGES_TAG_SCRIPT_FILENAME}")"; then
    echo "Error: An error occurred getting the Docker tag for the images produced by this repo" >&2
    exit 1
fi

# TODO Upgrade this to mount a Go build cache volume, so that all the containers don't need to re-compile everything
docker_build_args="${DEFAULT_DOCKER_BUILD_ARGS}"

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

build_dirpath="${root_dirpath}/${BUILD_DIRNAME}"
if ! mkdir -p "${build_dirpath}"; then
    echo "Error: Couldn't create build directory '${build_dirpath}'" >&2
    exit 1
fi

# ----------------------------- Interactive ---------------------------------
echo "Building Kurtosis interactive CLI & Javascript REPL image..."
cli_binary_output_filepath="${root_dirpath}/${CLI_BINARY_OUTPUT_REL_FILEPATH}"
if ! go build -o "${cli_binary_output_filepath}" "${root_dirpath}/${CLI_DIRPATH}/main.go"; then
    echo "Error: Failed to build the CLI binary" >&2
    exit 1
fi
javascript_repl_image="${JAVASCRIPT_REPL_IMAGE}:${docker_tag}"
if ! docker build ${docker_build_args} -f "${root_dirpath}/${JAVASCRIPT_REPL_DIRNAME}/Dockerfile" -t "${javascript_repl_image}" "${root_dirpath}/${JAVASCRIPT_REPL_DIRNAME}"; then
    echo "Error: Failed to build the Javascript REPL image" >&2
    exit 1
fi
echo "Successfully build Kurtosis interactive CLI & Javascript REPL image"

# -------------------------- Testing Framework ------------------------------
echo "Generating wrapper script..."
wrapper_generator_binary_output_filepath="${root_dirpath}/${WRAPPER_GENERATOR_BINARY_OUTPUT_REL_FILEPATH}"
if ! go build -o "${wrapper_generator_binary_output_filepath}" "${root_dirpath}/${WRAPPER_GENERATOR_DIRNAME}/main.go"; then
    echo "Error: Failed to build the wrapper script-generating binary" >&2
    exit 1
fi
wrapper_template_filepath="${root_dirpath}/${WRAPPER_TEMPLATE_REL_FILEPATH}"
wrapper_output_filepath="${root_dirpath}/${WRAPPER_OUTPUT_REL_FILEPATH}"
if ! "${wrapper_generator_binary_output_filepath}" -kurtosis-core-version "${docker_tag}" -template "${wrapper_template_filepath}" -output "${wrapper_output_filepath}"; then
    echo "Error: Failed to generate the wrapper script" >&2
    exit 1
fi
echo "Successfully generated wrapper script"

initializer_log_filepath="$(mktemp)"
api_log_filepath="$(mktemp)"
internal_testsuite_log_filepath="$(mktemp)"

echo "Launching builds of initializer, API, & internal testsuite images in parallel threads..."
docker build ${docker_build_args} -t "${INITIALIZER_IMAGE}:${docker_tag}" -f "${root_dirpath}/${INITIALIZER_DIRNAME}/Dockerfile" "${root_dirpath}" > "${initializer_log_filepath}" 2>&1 &
initializer_build_pid="${!}"
docker build ${docker_build_args} -t "${API_IMAGE}:${docker_tag}" -f "${root_dirpath}/api_container/Dockerfile" "${root_dirpath}" > "${api_log_filepath}" 2>&1 &
api_build_pid="${!}"
docker build ${docker_build_args} -t "${INTERNAL_TESTSUITE_IMAGE}:${docker_tag}" -f "${root_dirpath}/internal_testsuite/Dockerfile" "${root_dirpath}" > "${internal_testsuite_log_filepath}" 2>&1 &
internal_testsuite_build_pid="${!}"
echo "Build threads launched successfully:"
echo " - Initializer thread PID: ${initializer_build_pid}"
echo " - Initializer logs: ${initializer_log_filepath}"
echo " - API thread PID: ${api_build_pid}"
echo " - API logs: ${api_log_filepath}"
echo " - Internal testsuite thread PID: ${internal_testsuite_build_pid}"
echo " - Internal testsuite logs: ${internal_testsuite_log_filepath}"

echo "Waiting for build threads to exit..."
builds_succeeded=true
if ! wait "${initializer_build_pid}"; then
    builds_succeeded=false
fi
if ! wait "${api_build_pid}"; then
    builds_succeeded=false
fi
if ! wait "${internal_testsuite_build_pid}"; then
    builds_succeeded=false
fi
echo "Build threads exited"

echo ""
echo "========================= Initializer Image Build Logs ================================"
cat "${initializer_log_filepath}"

echo ""
echo "============================ API Image Build Logs ====================================="
cat "${api_log_filepath}"

echo ""
echo "===================== Internal Testsuite Image Build Logs ============================="
cat "${internal_testsuite_log_filepath}"

echo ""
if ! "${builds_succeeded}"; then
    echo "Build FAILED"
    exit 1
fi
echo "Build SUCCEEDED"
