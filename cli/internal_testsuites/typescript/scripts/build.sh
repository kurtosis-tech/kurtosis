#!/usr/bin/env bash
# 2021-07-08 WATERMARK, DO NOT REMOVE - This script was generated from the Kurtosis Bash script template

set -euo pipefail   # Bash "strict mode"
script_dirpath="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"
root_dirpath="$(dirname "${script_dirpath}")"


# Ignore these tests if building in kubernetes
if "$TESTING_KUBERNETES"; then
  ignore_patterns="/build/testsuite/(network_partition_test|network_soft_partition_test|service_pause_test)"
else
  ignore_patterns=""
fi


# ==================================================================================================
#                                             Main Logic
# ==================================================================================================
cd "${root_dirpath}"
rm -rf build
yarn install
yarn build --testPathIgnorePatterns=${ignore_patterns}
yarn test
