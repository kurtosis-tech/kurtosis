#!/usr/bin/env bash
# 2021-07-08 WATERMARK, DO NOT REMOVE - This script was generated from the Kurtosis Bash script template

set -euo pipefail   # Bash "strict mode"
script_dirpath="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"
root_dirpath="$(dirname "${script_dirpath}")"



# ==================================================================================================
#                                             Constants
# ==================================================================================================
# TODO Fill constants here with UPPER_SNAKE_CASE, noting that the only variables constants may use are:
# TODO  1) other constants (with the "${OTHER_CONSTANT}" syntax)
# TODO  2) script_dirpath/root_dirpath from above
DEFAULT_SOME_OPTIONAL_ARG_VALUE="A default value"   # TODO Replace with your own constants


# ==================================================================================================
#                                             Main Logic
# ==================================================================================================
# docker run --rm --privileged \
#     -v "${root_dirpath}:/go/src/github.com/user/repo" \
#     -v /var/run/docker.sock:/var/run/docker.sock \
#     -w /go/src/github.com/user/repo \
#     -e GITHUB_TOKEN \
#     -e DOCKER_USERNAME \
#     -e DOCKER_PASSWORD \
#     -e DOCKER_REGISTRY \
#     goreleaser/goreleaser \
#     build --rm-dist --snapshot --id wrapper-generator --single-target

docker run --rm --privileged \
    -v "${root_dirpath}:/go/src/github.com/user/repo" \
    -v /var/run/docker.sock:/var/run/docker.sock \
    -w /go/src/github.com/user/repo \
    -e GITHUB_TOKEN \
    -e DOCKER_USERNAME \
    -e DOCKER_PASSWORD \
    -e DOCKER_REGISTRY \
    goreleaser/goreleaser \
    release --rm-dist --snapshot
    # build --rm-dist --snapshot
