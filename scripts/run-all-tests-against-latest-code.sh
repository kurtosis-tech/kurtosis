#!/usr/bin/env bash
# 2021-07-08 WATERMARK, DO NOT REMOVE - This script was generated from the Kurtosis Bash script template

set -euo pipefail   # Bash "strict mode"
script_dirpath="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"
root_dirpath="$(dirname "${script_dirpath}")"


# ==================================================================================================
#                                             Constants
# ==================================================================================================
BUILD_SCRIPT_FILENAME="build.sh"

CLI_LAUNCH_PATH="${root_dirpath}/cli/cli/scripts/launch-cli.sh"

GOLANG_INTERNAL_TESTSUITES_TEST_SCRIPT_PATH="${root_dirpath}/internal_testsuites/golang/scripts/test.sh"
TYPESCRIPT_INTERNAL_TESTSUITES_TEST_SCRIPT_PATH="${root_dirpath}/internal_testsuites/typescript/scripts/test.sh"


TESTSUITE_CLUSTER_BACKEND_DOCKER="docker"
TESTSUITE_CLUSTER_BACKEND_MINIKUBE="minikube"

# By default, run testsuite against docker
DEFAULT_TESTSUITE_CLUSTER_BACKEND="${TESTSUITE_CLUSTER_BACKEND_DOCKER}"


# ==================================================================================================
#                                       Arg Parsing & Validation
# ==================================================================================================
show_helptext_and_exit() {
    echo "Usage: $(basename "${0}") cli_cluster_backend_arg"
    echo ""
    echo "  cli_cluster_backend_arg   Optional argument describing the cluster backend tests are running against. Must be one of '${TESTSUITE_CLUSTER_BACKEND_DOCKER}', '${TESTSUITE_CLUSTER_BACKEND_MINIKUBE}' (default: ${DEFAULT_TESTSUITE_CLUSTER_BACKEND})"
    echo ""
    exit 1  # Exit with an error so that if this is accidentally called by CI, the script will fail
}

testsuite_cluster_backend_arg="${1:-${DEFAULT_TESTSUITE_CLUSTER_BACKEND}}"
if [ "${testsuite_cluster_backend_arg}" != "${TESTSUITE_CLUSTER_BACKEND_DOCKER}" ] &&
   [ "${testsuite_cluster_backend_arg}" != "${TESTSUITE_CLUSTER_BACKEND_MINIKUBE}" ]; then
    echo "Error: unknown cluster provided to run tests against. Must be one of '${TESTSUITE_CLUSTER_BACKEND_DOCKER}', '${TESTSUITE_CLUSTER_BACKEND_MINIKUBE}'"
    show_helptext_and_exit
fi

# ==================================================================================================
#                                             Main Logic
# ==================================================================================================
# if the test suite is k8s we build & run images in k8s
# we start minikube if it isn't running
# we set the docker to be the one inside minikube
if [ "${testsuite_cluster_backend_arg}" == "${TESTSUITE_CLUSTER_BACKEND_MINIKUBE}" ]; then
    if ! minikube status; then
        if ! minikube start; then
            echo "Error starting Minikube" >&2
            exit 1
        fi
    fi
    if ! eval $(minikube docker-env); then
        echo "Error changing Docker image build environment to Minikube" >&2
        exit 1
    fi
fi

build_script_filepath="${script_dirpath}/${BUILD_SCRIPT_FILENAME}"
if ! bash "${script_dirpath}/${BUILD_SCRIPT_FILENAME}"; then
    echo "Error: Master buildscript '${build_script_filepath}' failed" >&2
    exit 1
fi

# stop existing engine
if ! bash "${CLI_LAUNCH_PATH}" engine stop; then
    echo "Error: Stopping the engine failed" >&2
    exit 1
fi

if [ "${testsuite_cluster_backend_arg}" == "${TESTSUITE_CLUSTER_BACKEND_MINIKUBE}" ]; then
  if ! bash "${CLI_LAUNCH_PATH}" cluster set "${TESTSUITE_CLUSTER_BACKEND_MINIKUBE}"; then
      echo "Error: setting cluster to '${TESTSUITE_CLUSTER_BACKEND_MINIKUBE}'" >&2
      exit 1
  fi
fi

if [ "${testsuite_cluster_backend_arg}" == "${TESTSUITE_CLUSTER_BACKEND_DOCKER}" ]; then
  if ! bash "${CLI_LAUNCH_PATH}" cluster set "${TESTSUITE_CLUSTER_BACKEND_DOCKER}"; then
      echo "Error: setting cluster to '${TESTSUITE_CLUSTER_BACKEND_DOCKER}'" >&2
      exit 1
  fi
fi

if ! bash "${CLI_LAUNCH_PATH}" engine start; then
    echo "Error: Starting the engine failed" >&2
    exit 1
fi

# if minikube run engine gateway
gateway_pid=""
if [ "${testsuite_cluster_backend_arg}" == "${TESTSUITE_CLUSTER_BACKEND_MINIKUBE}" ]; then
    "${CLI_LAUNCH_PATH}" "gateway" &
    gateway_pid="${!}"
    echo "Running gateway with pid '${gateway_pid}'"
fi

if ! bash "${GOLANG_INTERNAL_TESTSUITES_TEST_SCRIPT_PATH}" "${testsuite_cluster_backend_arg}"; then
    echo "Error: Build script '${GOLANG_INTERNAL_TESTSUITES_TEST_SCRIPT_PATH}' failed" >&2
    # kill the gateway pid if its k8s if failure
    if [ "${testsuite_cluster_backend_arg}" == "${TESTSUITE_CLUSTER_BACKEND_MINIKUBE}" ]; then
        kill "${gateway_pid}"
    fi
    exit 1
fi

if ! bash "${TYPESCRIPT_INTERNAL_TESTSUITES_TEST_SCRIPT_PATH}" "${testsuite_cluster_backend_arg}"; then
    echo "Error: Build script '${TYPESCRIPT_INTERNAL_TESTSUITES_TEST_SCRIPT_PATH}' failed" >&2
    # kill the gateway pid if its k8s if failure
    if [ "${testsuite_cluster_backend_arg}" == "${TESTSUITE_CLUSTER_BACKEND_MINIKUBE}" ]; then
        kill "${gateway_pid}"
    fi
    exit 1
fi

# kill the gateway pid if its k8s
if [ "${testsuite_cluster_backend_arg}" == "${TESTSUITE_CLUSTER_BACKEND_MINIKUBE}" ]; then
    kill "${gateway_pid}"
fi
