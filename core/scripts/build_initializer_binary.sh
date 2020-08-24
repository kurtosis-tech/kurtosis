#!/bin/bash

set -euo pipefail

BUILD_DIR="build"
MAIN_BINARY_OUTPUT_FILE="kurtosis-core"

script_dirpath=$(cd $(dirname "${BASH_SOURCE[0]}"); pwd)
root_dirpath=$(dirname "${script_dirpath}")

binary_output_filepath="${root_dirpath}/${BUILD_DIR}/${MAIN_BINARY_OUTPUT_FILE}"

echo "Running unit tests..."
go test "${root_dirpath}"/...

echo "Building..."
if go build -o "${binary_output_filepath}" "${root_dirpath}/todo_rename_new_initializer/main.go"; then
        echo "Build Successful"
        echo "Built initializer binary at ${binary_output_filepath}"
        echo "Run '${binary_output_filepath} --help' for usage."
else
        echo "Build failure"
        exit 1
fi
