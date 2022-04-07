#!/usr/bin/env bash
# 2021-07-08 WATERMARK, DO NOT REMOVE - This script was generated from the Kurtosis Bash script template

set -euo pipefail   # Bash "strict mode"
script_dirpath="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"
lang_root_dirpath="$(dirname "${script_dirpath}")"



# ==================================================================================================
#                                             Constants
# ==================================================================================================
PARALLELISM=2
TIMEOUT="3m"   # This must be Go-parseable timeout


# ==================================================================================================
#                                             Main Logic
# ==================================================================================================
cd "${lang_root_dirpath}"
go build ./...
# The count=1 disables caching for the testsuite (which we want, because the engine server might have changed even though the test code didn't)
CGO_ENABLED=0 go test ./... -p "${PARALLELISM}" -count=1 -timeout "${TIMEOUT}"
