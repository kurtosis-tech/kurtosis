#!/usr/bin/env bash
# 2021-07-08 WATERMARK, DO NOT REMOVE - This script was generated from the Kurtosis Bash script template

set -euo pipefail   # Bash "strict mode"
script_dirpath="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"
cli_module_dirpath="$(dirname "${script_dirpath}")"


# ==================================================================================================
#                                             Constants
# ==================================================================================================
source "${script_dirpath}/_constants.sh"

# ==================================================================================================
#                                             Main Logic
# ==================================================================================================
goarch="$(go env GOARCH)"
goos="$(go env GOOS)"

architecture_dirname="${GORELEASER_CLI_BUILD_ID}_${goos}_${goarch}"

if [ "${goarch}" == "${GO_ARCH_ENV_AMD64_VALUE}" ]; then
  goamd64="$(go env GOAMD64)"
  if [ "${goamd64}" == "" ]; then
    goamd64="${GO_DEFAULT_AMD64_ENV}"
  fi
  architecture_dirname="${architecture_dirname}_${goamd64}"
fi

cli_binary_filepath="${cli_module_dirpath}/${GORELEASER_OUTPUT_DIRNAME}/${architecture_dirname}/${CLI_BINARY_FILENAME}"

if ! [ -f "${cli_binary_filepath}" ]; then
    echo "Error: Expected a CLI binary to have been built by Goreleaser at '${cli_binary_filepath}' but none exists" >&2
    exit 1
fi

# The funky ${1+"${@}"} incantation is how you feed arguments exactly as-is to a child script in Bash
# ${*} loses quoting and ${@} trips set -e if no arguments are passed, so this incantation says, "if and only if
#  ${1} exists, evaluate ${@}"
"${cli_binary_filepath}" ${1+"${@}"}
