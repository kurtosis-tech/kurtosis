#!/bin/bash

set -o errexit
set -o nounset
set -o pipefail

BUILD_PREFIX="${PREFIX:-$(pwd)/build}"
MAIN_BINARY="kurtosis"

KURTOSIS_PKG=github.com/gmarchetti/kurtosis
KURTOSIS_PATH="$GOPATH/src/$KURTOSIS_PKG"

go build -o "$PREFIX/$MAIN_BINARY" "$KURTOSIS_PKG/main/"*.go

if [[ -f "$PREFIX/$MAIN_BINARY" ]]; then
        echo "Build Successful"
else
        echo "Build failure"
fi