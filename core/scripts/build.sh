#!/bin/bash

set -o errexit
set -o nounset
set -o pipefail


SCRIPTS_PATH=$(cd $(dirname "${BASH_SOURCE[0]}"); pwd)
KURTOSIS_PATH=$(dirname "${SCRIPTS_PATH}")

BUILD_DIR="build"
MAIN_DIR="ava_initializer"
MAIN_BINARY_OUTPUT_FILE="kurtosis"
MAIN_BINARY_OUTPUT_PATH="$KURTOSIS_PATH/$BUILD_DIR/$MAIN_BINARY_OUTPUT_FILE"


go build -o "$MAIN_BINARY_OUTPUT_PATH" "$KURTOSIS_PATH/$MAIN_DIR/"*.go

if [[ -f "$MAIN_BINARY_OUTPUT_PATH" ]]; then
        echo "Build Successful"
        echo "Built kurtosis binary to $MAIN_BINARY_OUTPUT_PATH"
        echo "Run $MAIN_BINARY_OUTPUT_PATH --help for usage."
else
        echo "Build failure"
fi