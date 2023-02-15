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
    exit 1
}


# ==================================================================================================
#                                             Main Logic
# ==================================================================================================
new_version="${1:-}"

if [ -z "${new_version}" ]; then
    if ! cd "${root_dirpath}"; then
        echo "Error: Couldn't cd to the root of this repo, '${root_dirpath}', which is required to get the Git tag" >&2
        show_helptext_and_exit
    fi
    if ! new_version="$(kudet get-docker-tag)"; then
        echo "Error: No new version provided and couldn't generate one" >&2
        show_helptext_and_exit
    fi
fi

kurtosis_version_go_file_abs_path="${root_dirpath}/${KURTOSIS_VERSION_PACKAGE_DIR}/${KURTOSIS_VERSION_GO_FILE}"
cat << EOF > "${kurtosis_version_go_file_abs_path}"
package ${KURTOSIS_VERSION_PACKAGE_DIR}

const (
  // !!!!!!!!!!!!!!!!!! DO NOT MODIFY THIS! IT WILL BE UPDATED AUTOMATICALLY DURING THE BUILD PROCESS !!!!!!!!!!!!!!!
	KurtosisVersion = "${new_version}"
  // !!!!!!!!!!!!!!!!!! DO NOT MODIFY THIS! IT WILL BE UPDATED AUTOMATICALLY DURING THE BUILD PROCESS !!!!!!!!!!!!!!!
)
EOF
