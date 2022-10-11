#!/usr/bin/env bash
# 2021-07-08 WATERMARK, DO NOT REMOVE - This script was generated from the Kurtosis Bash script template

set -euo pipefail   # Bash "strict mode"
script_dirpath="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"
root_dirpath="$(dirname "${script_dirpath}")"

# ==================================================================================================
#                                             Constants
# ==================================================================================================

KURTOSIS_VERSION_PACKAGE_DIR="kurtosis_version"
KURTOSIS_VERSION_GO_FILE="kurtosis_version.go"
KURTOSIS_VERSION_PACKAGE_NAME="github.com/kurtosis-tech/kurtosis/kurtosis_version"
KURTOSIS_VERSION_PACKAGE_GOSUM_PATH="go.sum"


# ==================================================================================================
#                                       Arg Parsing & Validation
# ==================================================================================================
show_helptext_and_exit() {
    echo "Usage: $(basename "${0}") new_version"
    echo ""
    echo "  new_version     The version to be generate the version constants with, otherwise uses 'kudet get-docker-tag'"
    echo ""
    exit 1  # Exit with an error so that if this is accidentally called by CI, the script will fail
}


# ==================================================================================================
#                                             Main Logic
# ==================================================================================================
new_version="${1:-}"

if [ -z "${new_version}" ]; then
  if ! new_version="$(kudet get-docker-tag)"; then
    echo "Error: No new version provided and couldn't generate one" >&2
    show_helptext_and_exit
  fi
fi

if ! rm "-rf" "${root_dirpath}/${KURTOSIS_VERSION_PACKAGE_DIR}"; then
  echo "Failed to remove '${root_dirpath}/${KURTOSIS_VERSION_PACKAGE_DIR}'"
  exit 1
fi

if ! mkdir "${root_dirpath}/${KURTOSIS_VERSION_PACKAGE_DIR}" ; then
  echo "Failed to create directory '${root_dirpath}/${KURTOSIS_VERSION_PACKAGE_DIR}'" >&2
  exit 1
fi

if ! cd "${root_dirpath}/${KURTOSIS_VERSION_PACKAGE_DIR}"; then
  echo "Failed to cd into directory '${root_dirpath}/${KURTOSIS_VERSION_PACKAGE_DIR}'" >&2
  exit 1
fi

if ! go "mod" "init" "${KURTOSIS_VERSION_PACKAGE_NAME}" ; then
  echo "Failed to create package '${KURTOSIS_VERSION_PACKAGE_NAME}'" >&2
  exit 1
fi

go_sum_abs_path="${root_dirpath}/${KURTOSIS_VERSION_PACKAGE_DIR}/${KURTOSIS_VERSION_PACKAGE_GOSUM_PATH}"
if ! touch "${go_sum_abs_path}" ; then
  echo "Failed to create go sum '${go_sum_abs_path}'" >&2
  exit 1
fi

kurtosis_version_go_file_abs_path="${root_dirpath}/${KURTOSIS_VERSION_PACKAGE_DIR}/${KURTOSIS_VERSION_GO_FILE}"
echo "package ${KURTOSIS_VERSION_PACKAGE_DIR}" > "${kurtosis_version_go_file_abs_path}"
echo "" >> "${kurtosis_version_go_file_abs_path}"
echo "const (" >> "${kurtosis_version_go_file_abs_path}"
echo "  // !!!!!!!!!!!!!!!!!! DO NOT MODIFY THIS! IT WILL BE UPDATED AUTOMATICALLY DURING THE BUILD PROCESS !!!!!!!!!!!!!!!" >> "${kurtosis_version_go_file_abs_path}"
echo "	KurtosisVersion = \"${new_version}\"" >> "${kurtosis_version_go_file_abs_path}"
echo "  // !!!!!!!!!!!!!!!!!! DO NOT MODIFY THIS! IT WILL BE UPDATED AUTOMATICALLY DURING THE BUILD PROCESS !!!!!!!!!!!!!!!" >> "${kurtosis_version_go_file_abs_path}"
echo ")" >> "${kurtosis_version_go_file_abs_path}"

