set -euo pipefail
script_dirpath="$(cd "$(dirname "${0}")" && pwd)"
root_dirpath="$(dirname "${script_dirpath}")"

GO_EXAMPLE_SUITE="kurtosistech/kurtosis-go-example"

# TODO TODO TODO

# The Go suite is designed to take in the nginxdemo/hello image - we only pass it in here as a demonstration of custom environment variables
"${root_dirpath}/build/kurtosis-core" "--test-suite-image=${GO_EXAMPLE_SUITE}" '--custom-env-vars-json={"GO_EXAMPLE_SERVICE_IMAGE":"nginxdemos/hello"}' ${*}
