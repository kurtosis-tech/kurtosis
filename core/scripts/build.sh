#!/bin/bash

set -o errexit
set -o nounset
set -o pipefail


SCRIPTS_PATH=$(cd $(dirname "${BASH_SOURCE[0]}"); pwd)
KURTOSIS_PATH=$(dirname "${SCRIPTS_PATH}")

echo "Running unit tests..."
if ! go test "${KURTOSIS_PATH}/..."; then
    echo "Tests failed!"
    exit 1
fi
echo "Tests succeeded"
