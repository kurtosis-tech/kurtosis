set -euo pipefail
script_dirpath="$(cd "$(dirname "${0}")" && pwd)"
root_dirpath="$(dirname "${script_dirpath}")"

INITIALIZER_IMAGE="kurtosistech/kurtosis-core_initializer"
GO_EXAMPLE_SUITE_IMAGE="kurtosistech/kurtosis-go-example:develop"

go_suite_execution_volume="suite-execution_go-example-suite_$(date +%s)"

docker volume create "${go_suite_execution_volume}"

# TODO feed in Go example suite image
docker run \
    --mount "type=bind,source=/var/run/docker.sock,target=/var/run/docker.sock" \
    --mount "type=volume,source=${go_suite_execution_volume},target=/suite-execution" \
    --env 'CUSTOM_ENV_VARS_JSON={"GO_EXAMPLE_SERVICE_IMAGE":"nginxdemos/hello"}' \
    --env "TEST_SUITE_IMAGE=${GO_EXAMPLE_SUITE_IMAGE}" \
    --env "SUITE_EXECUTION_VOLUME=${go_suite_execution_volume}" \
    ${*:-} \
    "${INITIALIZER_IMAGE}"
