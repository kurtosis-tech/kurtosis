#!/usr/bin/env bash
# 2021-07-08 WATERMARK, DO NOT REMOVE - This script was generated from the Kurtosis Bash script template

set -euo pipefail   # Bash "strict mode"
script_dirpath="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"
cli_module_dirpath="$(dirname "${script_dirpath}")"
root_dirpath="$(dirname "$(dirname "${cli_module_dirpath}")")"



# ==================================================================================================
#                                             Constants
# ==================================================================================================
source "${script_dirpath}/_constants.sh"

DEFAULT_SHOULD_PUBLISH_ARG="false"
DEFAULT_SHOULD_BUILD_ALL_ARG="false"

# ==================================================================================================
#                                       Arg Parsing & Validation
# ==================================================================================================
show_helptext_and_exit() {
    echo "Usage: $(basename "${0}") [should_build_all_arg] [should_publish_arg]"
    echo ""
    echo "  should_build_all_arg  Whether the build artifacts for all pair of OS/arch (default: ${DEFAULT_SHOULD_BUILD_ALL_ARG})"
    echo "  should_publish_arg    Whether the build artifacts should be published (default: ${DEFAULT_SHOULD_PUBLISH_ARG})"
    echo ""
    exit 1  # Exit with an error so that if this is accidentally called by CI, the script will fail
}

should_build_all_arg="${1:-"${DEFAULT_SHOULD_BUILD_ALL_ARG}"}"
if [ "${should_build_all_arg}" != "true" ] && [ "${should_build_all_arg}" != "false" ]; then
    echo "Error: Invalid should-build-all arg '${should_build_all_arg}'" >&2
    show_helptext_and_exit
fi

should_publish_arg="${2:-"${DEFAULT_SHOULD_PUBLISH_ARG}"}"
if [ "${should_publish_arg}" != "true" ] && [ "${should_publish_arg}" != "false" ]; then
    echo "Error: Invalid should-publish arg '${should_publish_arg}'" >&2
    show_helptext_and_exit
fi

# ==================================================================================================
#                                             Main Logic
# ==================================================================================================
# Generate Docker image tag
if ! cd "${root_dirpath}"; then
  echo "Error: Couldn't cd to the git root dirpath '${server_root_dirpath}'" >&2
  exit 1
fi
if ! version="$(./scripts/get-docker-tag.sh)"; then
    echo "Error: Couldn't get the version using get-docker-tag.sh" >&2
    exit 1
fi

(
    if ! cd "${cli_module_dirpath}"; then
        echo "Error: Couldn't cd to the CLI module directory in preparation for running Go generate & tests" >&2
        exit 1
    fi
    if ! go generate "./..."; then
        echo "Error: Go generate failed" >&2
        exit 1
    fi
    if ! CGO_ENABLED=1 go test "./..."; then
        echo "Error: Go tests failed" >&2
        exit 1
    fi
)

# vvvvvvvv Goreleaser variables vvvvvvvvvvvvvvvvvvv
export CLI_BINARY_FILENAME \
export VERSION="${version}"
if "${should_publish_arg}"; then
    # These environment variables will be set ONLY when publishing, in the CI environment
    # See the CI config for details on how these get set
    export FURY_TOKEN="${GEMFURY_PUBLISH_TOKEN}"
    export GITHUB_TOKEN="${KURTOSISBOT_GITHUB_TOKEN}"
fi
# ^^^^^^^^ Goreleaser variables ^^^^^^^^^^^^^^^^^^^

# Build a CLI binary (compatible with the current OS & arch) so that we can run interactive & testing locally via the launch-cli.sh script
(
    if ! cd "${cli_module_dirpath}"; then
        echo "Error: Couldn't cd to CLI module dirpath '${cli_module_dirpath}'" >&2
        exit 1
    fi
    if "${should_publish_arg}"; then
        goreleaser_verb_and_flags="release --rm-dist"
    elif "${should_build_all_arg}" ; then
        goreleaser_verb_and_flags="release --rm-dist --snapshot"
    else
        goreleaser_verb_and_flags="build --rm-dist --snapshot --single-target"
    fi
    if ! GORELEASER_CURRENT_TAG=$(cat $root_dirpath/version.txt) goreleaser ${goreleaser_verb_and_flags}; then
        echo "Error: Couldn't build the CLI binary for the current OS/arch" >&2
        exit 1
    fi
)

# Final verification
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
