#!/usr/bin/env bash
# 2021-07-08 WATERMARK, DO NOT REMOVE - This script was generated from the Kurtosis Bash script template

set -euo pipefail   # Bash "strict mode"
script_dirpath="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"
root_dirpath="$(dirname "${script_dirpath}")"
# Tests to ignore for Kubernetes
ignore_patterns="/build/testsuite/(network_partition_test|network_soft_partition_test|service_pause_test)"

# ==================================================================================================
#                                             Main Logic
# ==================================================================================================
cd "${root_dirpath}"
rm -rf build
yarn install
yarn build
yarn test --testPathIgnorePatterns=${ignore_patterns}
