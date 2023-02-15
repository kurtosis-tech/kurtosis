#!/usr/bin/env bash
# 2021-07-08 WATERMARK, DO NOT REMOVE - This script was generated from the Kurtosis Bash script template

set -euo pipefail   # Bash "strict mode"
script_dirpath="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"
root_dirpath="$(dirname "${script_dirpath}")"
git_root_dirpath="$(dirname ""$(dirname "${root_dirpath}")"")"

# ==================================================================================================
#                                             Constants
# ==================================================================================================
TESTSUITE_CLUSTER_BACKEND_DOCKER="docker"
TESTSUITE_CLUSTER_BACKEND_MINIKUBE="minikube"

# By default, run testsuite against docker
DEFAULT_TESTSUITE_CLUSTER_BACKEND="${TESTSUITE_CLUSTER_BACKEND_DOCKER}"

# Pattern for tests to ignore when running against kubernetes
# Four tests are ignored
# network_partition_test,network_soft_partition_test - Networking partitioning is not implemented in kubernetes
# service_pause_test - Service pausing not implemented in Kubernetes
# stream_log_test - The centralized logs feature is not implemented in Kubernetes yet
KUBERNETES_TEST_IGNORE_PATTERNS="/build/testsuite/(network_partition_starlark|network_partition_test|network_soft_partition_test|service_pause_test|stream_log_test|search_logs_test)"

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

testsuite_cluster_backend_arg="${1:-"${DEFAULT_TESTSUITE_CLUSTER_BACKEND}"}"
if [ "${testsuite_cluster_backend_arg}" != "${TESTSUITE_CLUSTER_BACKEND_DOCKER}" ] &&
   [ "${testsuite_cluster_backend_arg}" != "${TESTSUITE_CLUSTER_BACKEND_MINIKUBE}" ]; then
    echo "Error: unknown cluster provided to run tests against. Must be one of 'docker', 'minikube'"
    show_helptext_and_exit
fi

# ==================================================================================================
#                                             Main Logic
# ==================================================================================================

rm -rf node_modules
echo "Building the kurtosis-sdk in typescript as this script depends on it"
typescript_sdk_build_path="${git_root_dirpath}/api/typescript/scripts/build.sh"
if ! "${typescript_sdk_build_path}"; then
    echo "Error: SDK buildscript at '${typescript_sdk_build_path}' failed" >&2
    exit 1
fi

cd "${root_dirpath}"
rm -rf build
yarn install
yarn build

if [ "${testsuite_cluster_backend_arg}" == "${TESTSUITE_CLUSTER_BACKEND_MINIKUBE}" ]; then
    # TODO This should be removed! Typescript Kurtosis tests should be completely agnostic to the backend they're running against
    # The only reason this exists is because, as of 2022-10-28, network partitioning doesn't work on Kubernetes so we have to know to skip
    #  those tests
    yarn test --testPathIgnorePatterns=${KUBERNETES_TEST_IGNORE_PATTERNS}
else
    yarn test
fi
