#!/usr/bin/env bash
# 2021-07-08 WATERMARK, DO NOT REMOVE - This script was generated from the Kurtosis Bash script template

set -euo pipefail   # Bash "strict mode"
script_dirpath="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"
lang_root_dirpath="$(dirname "${script_dirpath}")"



# ==================================================================================================
#                                             Constants
# ==================================================================================================
PARALLELISM=2
DOCKER_TIMEOUT="3m"   # This must be Go-parseable timeout
KUBERNETES_TIMEOUT="8m" # K8S takes longer than docker

TESTSUITE_CLUSTER_BACKEND_DOCKER="docker"
TESTSUITE_CLUSTER_BACKEND_KUBERNETES="kubernetes"

TEST_IS_RUNNING_ON_CIRCLE_CI="true"
TEST_IS_NOT_RUNNING_ON_CIRCLE_CI="false"

# By default, run testsuite against docker
DEFAULT_TESTSUITE_CLUSTER_BACKEND="${TESTSUITE_CLUSTER_BACKEND_DOCKER}"
DEFAULT_IS_RUNNING_ON_CIRCLE_CI="${TEST_IS_NOT_RUNNING_ON_CIRCLE_CI}"

# ==================================================================================================
#                                       Arg Parsing & Validation
# ==================================================================================================
show_helptext_and_exit() {
    echo "Usage: $(basename "${0}") cli_cluster_backend_arg"
    echo ""
    echo "  cli_cluster_backend_arg   Optional argument describing the cluster backend tests are running against. Must be one of 'docker', 'kubernetes' (default: ${DEFAULT_TESTSUITE_CLUSTER_BACKEND})"
    echo "  circle_ci_arg             Optional argument that allows for test splitting on Circle CI. Must be one of 'true' or 'false'"
    echo ""
    exit 1  # Exit with an error so that if this is accidentally called by CI, the script will fail
}

testsuite_cluster_backend_arg="${1:-${DEFAULT_TESTSUITE_CLUSTER_BACKEND}}"
if [ "${testsuite_cluster_backend_arg}" != "${TESTSUITE_CLUSTER_BACKEND_DOCKER}" ] &&
   [ "${testsuite_cluster_backend_arg}" != "${TESTSUITE_CLUSTER_BACKEND_KUBERNETES}" ]; then
    echo "Error: unknown cluster provided to run tests against. Must be one of 'docker', 'kubernetes'"
    show_helptext_and_exit
fi

testsuite_is_running_on_circleci=${2:-${DEFAULT_IS_RUNNING_ON_CIRCLE_CI}}
if [ "${testsuite_is_running_on_circleci}" != "${TEST_IS_RUNNING_ON_CIRCLE_CI}" ] &&
   [ "${testsuite_is_running_on_circleci}" != "${TEST_IS_NOT_RUNNING_ON_CIRCLE_CI}" ]; then
    echo "Error: unknown value for whether the test is running against circleci. Must be one of 'false', 'true'"
    show_helptext_and_exit
fi

# ==================================================================================================
#                                             Main Logic
# ==================================================================================================

cd "${lang_root_dirpath}"
go build ./...

# The count=1 disables caching for the testsuite (which we want, because the engine server might have changed even though the test code didn't)
if [ "${testsuite_cluster_backend_arg}" == "${TESTSUITE_CLUSTER_BACKEND_KUBERNETES}" ]; then
    # TODO This should be removed! Go Kurtosis tests should be completely agnostic to the backend they're running against
    # The only reason this exists is because, some Kurtosis feature doesn't work on Kubernetes so we have to know to skip
    #  those tests
    # K8S is also slower than docker, so they have different timeouts
    if [ "${testsuite_is_running_on_circleci}" == "${TEST_IS_RUNNING_ON_CIRCLE_CI}" ]; then
        CGO_ENABLED=0 go test -v $(go list -tags kubernetes ./...| circleci tests split) -p "${PARALLELISM}" -count=1 -timeout "${KUBERNETES_TIMEOUT}" -tags kubernetes
    else
        CGO_ENABLED=0 go test ./... -p "${PARALLELISM}" -count=1 -timeout "${KUBERNETES_TIMEOUT}" -tags kubernetes
    fi
else
    if [ "${testsuite_is_running_on_circleci}" == "${TEST_IS_RUNNING_ON_CIRCLE_CI}" ]; then
        CGO_ENABLED=0 go test -v $(go list ./... | circleci tests split) -p "${PARALLELISM}" -count=1 -timeout "${DOCKER_TIMEOUT}"
    else
        CGO_ENABLED=0 go test ./... -p "${PARALLELISM}" -count=1 -timeout "${DOCKER_TIMEOUT}"
    fi
fi

