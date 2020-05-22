#!/bin/bash

set -o errexit
set -o nounset
set -o pipefail

PREFIX="${PREFIX:-$(pwd)/build}"
MAIN_BINARY="kurtosis"

go build -o "$PREFIX/$MAIN_BINARY" "$GECKO_PATH/main/"*.go

if [[ -f "$PREFIX/$MAIN_BINARY" ]]; then
        echo "Build Successful"
else
        echo "Build failure"
fi