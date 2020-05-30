#!/bin/bash

set -o errexit
set -o nounset
set -o pipefail

SCRIPTS_PATH=$(cd $(dirname "${BASH_SOURCE[0]}"); pwd)
KURTOSIS_PATH=$(dirname "${SCRIPTS_PATH}")
DOCKER="${DOCKER:-docker}"

COMMIT="$(git --git-dir="${KURTOSIS_PATH}/.git" rev-parse --short HEAD)"

"${DOCKER}" build -t "kurtosis-$COMMIT" "${KURTOSIS_PATH}"
