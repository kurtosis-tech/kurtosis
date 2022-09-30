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
KURTOSIS_VERSION_PACKAGE_GOSUM_PATH="go.sum"


# ==================================================================================================
#                                       Arg Parsing & Validation
# ==================================================================================================
show_helptext_and_exit() {
    echo "Usage: $(basename "${0}")"
    exit 1
}


# ==================================================================================================
#                                             Main Logic
# ==================================================================================================

if ![--d "${root_dirpath}/${KURTOSIS_VERSION_PACKAGE_DIR}"]; then
  if ! bash "mkdir" "${root_dirpath}/${KURTOSIS_VERSION_PACKAGE_DIR}" ; then
    echo "Failed to create directory '${root_dirpath}/${KURTOSIS_VERSION_PACKAGE_DIR}'" >&2
    exit 1
  fi
fi

if ! bash "cd" "${root_dirpath}/${KURTOSIS_VERSION_PACKAGE_DIR}"; then
  echo "Failed to cd into directory '${root_dirpath}/${KURTOSIS_VERSION_PACKAGE_DIR}'" >&2
fi

if ! bash "go" "mod" "init" "${KURTOSIS_VERSION_PACKAGE_NAME}" ; then
  echo "Failed to create package '${KURTOSIS_VERSION_PACKAGE_NAME}'" >&2
  exit 1
fi

go_sum_path="${root_dirpath}/${KURTOSIS_VERSION_PACKAGE_DIR}/${KURTOSIS_VERSION_PACKAGE_GOSUM_PATH}"
if ! bash "go" "touch" "${go_sum_path}" ; then
  echo "Failed to create package '${go_sum_path}'" >&2
  exit 1
fi

if ! docker_tag="$(kudet get-docker-tag)"; then
    echo "Error: Couldn't get the Docker image tag" >&2
    exit 1
fi

