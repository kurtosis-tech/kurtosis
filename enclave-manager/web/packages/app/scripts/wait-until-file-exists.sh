#!/usr/bin/env bash
# 2021-07-08 WATERMARK, DO NOT REMOVE - This script was generated from the Kurtosis Bash script template

set -euo pipefail   # Bash "strict mode"

# ==================================================================================================
#                                       Arg Parsing & Validation
# ==================================================================================================
show_helptext_and_exit() {
    echo "Usage: $(basename "${0}") file_name"
    echo ""
    echo "  file_name         The file to wait to exist"
    echo ""
    exit 1  # Exit with an error so that if this is accidentally called by CI, the script will fail
}

file_name="${1:-}"

if [ -z "${file_name}" ]; then
    echo "Error: No file to wait for provided" >&2
    show_helptext_and_exit
fi

# ==================================================================================================
#                                             Main Logic
# ==================================================================================================
# Block until the given file appears or the given timeout is reached.
# Exit status is 0 iff the file exists.
wait_file() {
  local file="$1"; shift
  local wait_seconds="${1:-30}"; shift # 30 seconds as default timeout
  test $wait_seconds -lt 1 && echo 'At least 1 second is required' && return 1

  until test $((wait_seconds--)) -eq 0 -o -e "$file" ; do sleep 1; done

  test $wait_seconds -ge 0 # equivalent: let ++wait_seconds
}

wait_file "$file_name" 30
exit 0
