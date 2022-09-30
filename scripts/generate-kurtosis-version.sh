#!/usr/bin/env bash
# 2021-07-08 WATERMARK, DO NOT REMOVE - This script was generated from the Kurtosis Bash script template

set -euo pipefail   # Bash "strict mode"
script_dirpath="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"
root_dirpath="$(dirname "${script_dirpath}")"

# ==================================================================================================
#                                             Constants
# ==================================================================================================

KURTOSIS_VERSION_PACKAGE_DIR="kurtosis_version"
KURTOSIS_VERSION_PACKAGE_NAME="github.com/kurtosis-tech/kurtosis/kurtosis_version"
KURTOSIS_VERSION_PACKAGE_GOSUM_PATH="kurtosis_version/go.sum"

# ==================================================================================================
#                                             Main Logic
# ==================================================================================================

