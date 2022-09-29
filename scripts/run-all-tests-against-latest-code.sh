#!/usr/bin/env bash
# 2021-07-08 WATERMARK, DO NOT REMOVE - This script was generated from the Kurtosis Bash script template

set -euo pipefail   # Bash "strict mode"
script_dirpath="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"
root_dirpath="$(dirname "${script_dirpath}")"


# ==================================================================================================
#                                             Constants
# ==================================================================================================

BUILD_SCRIPT_RELATIVE_FILEPATHS=(
  "core/server/scripts/build.sh"
  "core/files_artifacts_expander/scripts/build.sh"
  "cli/scripts/build.sh"
)

ENGINE_SERVER_BUILDSCRIPT_PATH="${root_dirpath}/engine/server/scripts/build.sh"
SHOULD_RUN_UNITTESTS="false"

RUN_PRE_RELEASE_SCRIPTS_SCRIPT_PATH="${script_dirpath}/run-pre-release-scripts.sh"
CLI_LAUNCH_PATH="${root_dirpath}/cli/cli/scripts/launch-cli.sh"
INTERNAL_TESTSUITES_BUILDSCRIPT_PATH="${root_dirpath}/internal_testsuites/scripts/build.sh"


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

if ! bash "${RUN_PRE_RELEASE_SCRIPTS_SCRIPT_PATH}"; then
  echo "Error: Error running pre release scripts '${RUN_PRE_RELEASE_SCRIPTS_SCRIPT_PATH}' failed" >&2
  exit 1
fi

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


if ! bash "${ENGINE_SERVER_BUILDSCRIPT_PATH}" "${SHOULD_RUN_UNITTESTS}"; then
  echo "Error: Build script '${ENGINE_SERVER_BUILDSCRIPT_PATH}' failed" >&2
  exit 1
fi

for build_script_rel_filepath in "${BUILD_SCRIPT_RELATIVE_FILEPATHS[@]}"; do
    build_script_abs_filepath="${root_dirpath}/${build_script_rel_filepath}"
    if ! bash "${build_script_abs_filepath}"; then
        echo "Error: Build script '${build_script_abs_filepath}' failed" >&2
        exit 1
    fi
done

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

# stop existing engine
if ! bash "${CLI_LAUNCH_PATH}" engine stop; then
    echo "Error: Stopping the engine failed" >&2
    exit 1
fi

if ! bash "${CLI_LAUNCH_PATH}" engine restart --version ${CORE_ENGINE_VERSION_TAG}; then
    echo "Restarting the engine failed" >&2
    exit 1
fi

# if minikube run engine gateway
gateway_pid=""
if [ "${testsuite_cluster_backend_arg}" == "${TESTSUITE_CLUSTER_BACKEND_MINIKUBE}" ]; then
  "${CLI_LAUNCH_PATH}" "gateway" &
  gateway_pid="${!}"
  echo "Running gateway with pid '${gateway_pid}'"
fi

if ! bash "${INTERNAL_TESTSUITES_BUILDSCRIPT_PATH}" "${testsuite_cluster_backend_arg}"; then
    echo "Error: Build script '${INTERNAL_TESTSUITES_BUILDSCRIPT_PATH}' failed" >&2
    kill "${gateway_pid}"
    exit 1
fi

# kill the gateway pid if its k8s
if [ "${testsuite_cluster_backend_arg}" == "${TESTSUITE_CLUSTER_BACKEND_MINIKUBE}" ]; then
  kill "${gateway_pid}"
fi