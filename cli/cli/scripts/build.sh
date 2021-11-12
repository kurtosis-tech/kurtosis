#!/usr/bin/env bash
# 2021-07-08 WATERMARK, DO NOT REMOVE - This script was generated from the Kurtosis Bash script template

set -euo pipefail   # Bash "strict mode"
script_dirpath="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"
cli_module_dirpath="$(dirname "${script_dirpath}")"
root_dirpath="$(dirname "${cli_module_dirpath}")"



# ==================================================================================================
#                                             Constants
# ==================================================================================================
source "${script_dirpath}/_constants.sh"

DEFAULT_SHOULD_PUBLISH_ARG="false"

GET_VERSION_SCRIPT_FILENAME="get-docker-images-tag.sh"

# ==================================================================================================
#                                       Arg Parsing & Validation
# ==================================================================================================
show_helptext_and_exit() {
    echo "Usage: $(basename "${0}") [should_publish_arg]"
    echo ""
    echo "  should_publish_arg  Whether the build artifacts should be published (default: ${DEFAULT_SHOULD_PUBLISH_ARG})"
    echo ""
    exit 1  # Exit with an error so that if this is accidentally called by CI, the script will fail
}

should_publish_arg="${1:-"${DEFAULT_SHOULD_PUBLISH_ARG}"}"
if [ "${should_publish_arg}" != "true" ] && [ "${should_publish_arg}" != "false" ]; then
    echo "Error: Invalid should-publish arg '${should_publish_arg}'" >&2
    show_helptext_and_exit
fi

# ==================================================================================================
#                                             Main Logic
# ==================================================================================================
get_version_script_filepath="${root_dirpath}/scripts/${GET_VERSION_SCRIPT_FILENAME}"
if ! version="$("${get_version_script_filepath}")"; then
    echo "Error: Couldn't get version using script '${get_version_script_filepath}'" >&2
    exit 1
fi

# Now that we have the REPL Dockerfiles, build the CLI
# vvvvvvvv Goreleaser variables vvvvvvvvvvvvvvvvvvv
export CLI_BINARY_FILENAME \
export VERSION="${version}"
if "${should_publish_arg}"; then
    # This environment variable will be set ONLY when publishing, in the CI environment
    # See the CI config for details on how this gets set
    export FURY_TOKEN
fi
# ^^^^^^^^ Goreleaser variables ^^^^^^^^^^^^^^^^^^^

# Build a CLI binary (compatible with the current OS & arch) so that we can run interactive & testing locally via the launch-cli.sh script
(
    if ! cd "${cli_module_dirpath}"; then
        echo "Error: Couldn't cd to CLI module dirpath '${cli_module_dirpath}'" >&2
        exit 1
    fi
    # TODO TODO IF PUBLISHING, THEN PASS IN DIFFERENT ARGS
    if ! goreleaser build --rm-dist --snapshot --id "${GORELEASER_CLI_BUILD_ID}" --single-target; then
        echo "Error: Couldn't build the CLI binary for the current OS/arch" >&2
        exit 1
    fi

    # TODO TODO TODO HOOK PUBLISHING BACK UP
    # # Build all the Docker images
    # if "${should_publish_arg}"; then
    #     goreleaser_release_extra_args=""
    # else
    #     goreleaser_release_extra_args="--snapshot"
    # fi
    # if ! goreleaser release --rm-dist --skip-announce ${goreleaser_release_extra_args}; then
    #     echo "Error: Goreleaser release of all binaries & Docker images failed" >&2
    #     exit 1
    # fi
)

# Now that we have a CLI built from source, start the version of the engine that the CLI uses
goarch="$(go env GOARCH)"
goos="$(go env GOOS)"
cli_binary_filepath="${cli_module_dirpath}/${GORELEASER_OUTPUT_DIRNAME}/${GORELEASER_CLI_BUILD_ID}_${goos}_${goarch}/${CLI_BINARY_FILENAME}"
if ! [ -f "${cli_binary_filepath}" ]; then
    echo "Error: Expected a CLI binary to have been built by Goreleaser at '${cli_binary_filepath}' but none exists" >&2
    exit 1
fi
