# This script is intended to be sources by the other scripts in this directory
DOCKER_ORG="kurtosistech"
REPO_BASE="kurtosis-core"

API_IMAGE="${DOCKER_ORG}/${REPO_BASE}_api"

KURTOSIS_DIRPATH="$HOME/.kurtosis"

BUILD_DIRNAME="build"
GORELEASER_OUTPUT_DIRNAME="dist"

GET_DOCKER_IMAGES_TAG_SCRIPT_FILENAME="get-docker-images-tag.sh"

# ------------------------ Testing  -------------------------------------------------------
INTERNAL_TESTSUITE_IMAGE_SUFFIX="internal-testsuite"

# ---------------------- Interactive  -----------------------------------------------------
GORELEASER_CLI_BUILD_ID="cli"
CLI_BINARY_FILENAME="kurtosis"  # Name of the CLI binary
JAVASCRIPT_REPL_IMAGE="${DOCKER_ORG}/javascript-interactive-repl"
