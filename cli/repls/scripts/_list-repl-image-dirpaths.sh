#!/usr/bin/env bash
# 2021-07-08 WATERMARK, DO NOT REMOVE - This script was generated from the Kurtosis Bash script template

set -euo pipefail   # Bash "strict mode"
script_dirpath="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"
repls_dirpath="$(dirname "${script_dirpath}")"

# ==================================================================================================
#                                             Constants
# ==================================================================================================
REPL_IMAGES_DIRNAME="repl_images"


# ==================================================================================================
#                                             Main Logic
# ==================================================================================================
repl_images_dirpath="${repls_dirpath}/${REPL_IMAGES_DIRNAME}"
find "${repl_images_dirpath}" -mindepth 1 -maxdepth 1 -type d
