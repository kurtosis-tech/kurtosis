#!/bin/bash

set -o errexit
set -o nounset
set -o pipefail

if [ $# -eq 0 ]
  then
    COMMIT="$(git --git-dir="${KURTOSIS_PATH}/.git" rev-parse --short HEAD)"
    TAG="kurtosis-$COMMIT"
  else
    TAG=$1
fi

IMAGE_TAG=$1
SCRIPTS_PATH=$(cd $(dirname "${BASH_SOURCE[0]}"); pwd)
KURTOSIS_PATH=$(dirname "${SCRIPTS_PATH}")
DOCKER="${DOCKER:-docker}"

COMMIT="$(git --git-dir="${KURTOSIS_PATH}/.git" rev-parse --short HEAD)"

"${DOCKER}" build -t "$TAG" "${KURTOSIS_PATH}"
