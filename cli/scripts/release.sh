#!/usr/bin/env bash
# 2021-07-08 WATERMARK, DO NOT REMOVE - This script was generated from the Kurtosis Bash script template

set -euo pipefail   # Bash "strict mode"
script_dirpath="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"
root_dirpath="$(dirname "${script_dirpath}")"

# ==================================================================================================
#                                             Constants
# ==================================================================================================
RELEASE_CMD_NAME="kudet release"     # NOTE: kudet should be installed by every Kurtosis dev

# ==================================================================================================
#                                             Main Logic
# ==================================================================================================
if ! cd "${root_dirpath}"; then
  echo "Couldn't cd to the git root dirpath '${root_dirpath}'" >&2
  exit 1
if ! $(${RELEASE_CMD_NAME}); then
    echo "Error: Couldn't cut the release" >&2
    exit 1
fi
