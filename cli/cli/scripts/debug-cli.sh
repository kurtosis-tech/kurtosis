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

# Check and install Go Delve
# it should be installed if you are using Nix, but it's better to check first because the script could be executed from outside the Nix scope
echo "Checking for Delve..."
if ! dlv version ; then
  echo "Delve binary was no found, we recommend you whether to execute this script inside the Nix scope (if you are using Nix) or install the Delve binary with this cmd: 'go install github.com/go-delve/delve/cmd/dlv@latest'"
  exit 1
fi
echo "...it was found."

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
cli_arguments_and_flags=${1+"${@}"}
first_argument=$1
headless_val="true"
if [ "${first_argument}" == "dlv-terminal" ]; then
  headless_val="false"
  # The CLI's arguments start from the second position
  cli_arguments_and_flags=${2+"${@:2}"}
fi

# Split between program arguments and flags
cli_arguments=""
cli_flags=""
for argument_or_flag in ${cli_arguments_and_flags}
do
  # If it's a flag
  if [[ ${argument_or_flag} == -* ]]; then
    cli_flags="${cli_flags} ${argument_or_flag}"
  else
    cli_arguments="${cli_arguments} ${argument_or_flag}"
  fi
done

dlv --listen="127.0.0.1:${CLI_DEBUG_SERVER_PORT}" --headless="${headless_val}" --api-version=2 --check-go-version=false --only-same-user=false exec "${cli_binary_filepath}" ${cli_arguments} -- "--debug-mode" ${cli_flags}
