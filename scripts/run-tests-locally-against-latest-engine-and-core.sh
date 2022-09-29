#!/usr/bin/env bash
# 2021-07-08 WATERMARK, DO NOT REMOVE - This script was generated from the Kurtosis Bash script template

set -euo pipefail   # Bash "strict mode"
script_dirpath="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"
root_dirpath="$(dirname "${script_dirpath}")"
cli_launch_path="${root_dirpath}/cli/cli/scripts/launch-cli.sh"
internal_test_suite_build_script_path="${root_dirpath}/internal_testsuites/scripts/build.sh"



# ==================================================================================================
#                                             Constants
# ==================================================================================================
BUILD_SCRIPT_RELATIVE_FILEPATHS=(
    "scripts/create-server-images-and-populate-env.sh"
    "cli/scripts/build.sh"
)

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
    echo "  cli_cluster_backend_arg   Optional argument describing the cluster backend tests are running against. Must be one of 'docker', 'minikube' (default: ${DEFAULT_TESTSUITE_CLUSTER_BACKEND})"
    echo ""
    exit 1  # Exit with an error so that if this is accidentally called by CI, the script will fail
}

testsuite_cluster_backend_arg="${1:-${DEFAULT_TESTSUITE_CLUSTER_BACKEND}}"
if [ "${testsuite_cluster_backend_arg}" != "${TESTSUITE_CLUSTER_BACKEND_DOCKER}" ] &&
   [ "${testsuite_cluster_backend_arg}" != "${TESTSUITE_CLUSTER_BACKEND_MINIKUBE}" ]; then
    echo "Error: unknown cluster provided to run tests against. Must be one of 'docker', 'minikube'"
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
          echo "Error starting minikube" >&2
          exit 1
      fi
    fi
    if ! eval $(minikube docker-env); then
        echo "Error changing docker environment to minikube" >&2
        exit 1
    fi
fi

# stop existing engine
if ! bash "${cli_launch_path}" engine stop; then
    echo "Error: Stopping the engine failed" >&2
    exit 1
fi

# if k8s we set cluster to minikube
if ! bash "${cli_launch_path}" cluster set minikube; then
    echo "Setting the cluster to minikube failed" >&2
    exit 1
fi

source ${root_dirpath}/_kurtosis_servers.env

if ! bash "${cli_launch_path}" engine restart --version ${CORE_ENGINE_VERSION_TAG}; then
    echo "Restarting the engine failed" >&2
    exit 1
fi

# if minikube run engine gateway
gateway_pid=""
if [ "${testsuite_cluster_backend_arg}" == "${TESTSUITE_CLUSTER_BACKEND_MINIKUBE}" ]; then
  "${cli_launch_path}" "gateway" &
  gateway_pid="${!}"
fi

if ! bash "${internal_test_suite_build_script_path}" "${testsuite_cluster_backend_arg}"; then
    echo "Error: Build script '${internal_test_suite_build_script_path}' failed" >&2
    exit 1
fi

# kill the gateway pid if its k8s
if [ "${testsuite_cluster_backend_arg}" == "${TESTSUITE_CLUSTER_BACKEND_MINIKUBE}" ]; then
  "kill" "${gateway_pid}"
fi