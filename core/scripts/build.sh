#!/bin/bash

set -o errexit
set -o nounset
set -o pipefail


SCRIPTS_PATH=$(cd $(dirname "${BASH_SOURCE[0]}"); pwd)
KURTOSIS_PATH=$(dirname "${SCRIPTS_PATH}")

BUILD_DIR="build"
MAIN_BINARY_OUTPUT_FILE="kurtosis"
MAIN_BINARY_OUTPUT_PATH="${KURTOSIS_PATH}/${BUILD_DIR}/${MAIN_BINARY_OUTPUT_FILE}"

echo "Running unit tests..."
if ! go test "${KURTOSIS_PATH}/..."; then
    echo "Tests failed!"
    exit 1
fi
echo "Tests succeeded"
