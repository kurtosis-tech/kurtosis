set -euo pipefail
script_dirpath="$(cd "$(dirname "${0}")" && pwd)"
root_dirpath="$(dirname "${script_dirpath}")"

GO_EXAMPLE_SUITE="kurtosistech/kurtosis-go-example"

"${root_dirpath}/build/kurtosis-core" "--test-suite-image=${GO_EXAMPLE_SUITE}" ${*}
