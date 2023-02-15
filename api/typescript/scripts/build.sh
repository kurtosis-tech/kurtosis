#!/usr/bin/env bash
# 2021-07-08 WATERMARK, DO NOT REMOVE - This script was generated from the Kurtosis Bash script template

set -euo pipefail   # Bash "strict mode"
script_dirpath="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"
lang_root_dirpath="$(dirname "${script_dirpath}")"


# ==================================================================================================
#                                             Main Logic
# ==================================================================================================
cd "${lang_root_dirpath}"
yarn install
CGO_ENABLED=0 yarn test
yarn build
