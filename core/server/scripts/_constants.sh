# This script is intended to be sources by the other scripts in this directory
DOCKER_ORG="kurtosistech"
REPO_BASE="kurtosis-core"

API_IMAGE="${DOCKER_ORG}/${REPO_BASE}_api"

BUILD_DIRNAME="build"
GORELEASER_OUTPUT_DIRNAME="dist"

GET_FIXED_DOCKER_IMAGES_TAG_SCRIPT_FILENAME="get-fixed-docker-images-tag.sh"
