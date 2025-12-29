#!/usr/bin/env bash
# 2021-07-08 WATERMARK, DO NOT REMOVE - This script was generated from the Kurtosis Bash script template

set -euo pipefail   # Bash "strict mode"

# ==================================================================================================
#                                             Main Logic
# ==================================================================================================

if ! git status > /dev/null; then
  echo "Error: This command was run from outside a git repo" >&2
  exit 1
fi

commit_sha="$(git rev-parse --short=6 HEAD)"
suffix="$(git diff --quiet || echo '-dirty')"
echo "${commit_sha}${suffix}"