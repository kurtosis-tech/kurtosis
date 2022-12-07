#!/usr/bin/env bash

set -eu pipefail
if [ -z "${CLI_BINARY_FILENAME}" ]; then
    echo "Environment variable CLI_BINARY_FILENAME must not be empty" >&2
    exit 1
fi
COMPLETIONS_BUILD_DIR="completions/"
COMPLETIONS_BINARY_DIR="${COMPLETIONS_BUILD_DIR}/cli/"
COMPLETIONS_SCRIPT_DIR="${COMPLETIONS_BUILD_DIR}/scripts/"
CLI_BINARY_PATH="${COMPLETIONS_BINARY_DIR}/${CLI_BINARY_FILENAME}"

rm -rf "${COMPLETIONS_BUILD_DIR}"
mkdir -p "${COMPLETIONS_BINARY_DIR}"
mkdir -p "${COMPLETIONS_SCRIPT_DIR}"

go build -o "${CLI_BINARY_PATH}" main.go

for sh in bash zsh fish; do
	"${CLI_BINARY_PATH}" completion "$sh" > "${COMPLETIONS_SCRIPT_DIR}/kurtosis.$sh"
done
