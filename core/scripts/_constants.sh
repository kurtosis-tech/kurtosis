# This script is intended to be sources by the other scripts in this directory
DOCKER_ORG="kurtosistech"
REPO_BASE="kurtosis-core"

API_IMAGE="${DOCKER_ORG}/${REPO_BASE}_api"

KURTOSIS_DIRPATH="$HOME/.kurtosis"

BUILD_DIRNAME="build"

GET_DOCKER_IMAGES_TAG_SCRIPT_FILENAME="get-docker-images-tag.sh"

# ------------------------ Testing  -------------------------------------------------------
WRAPPER_GENERATOR_DIRNAME="wrapper_generator"
WRAPPER_GENERATOR_BINARY_OUTPUT_REL_FILEPATH="${BUILD_DIRNAME}/wrapper-generator"   # Relative to repo root
WRAPPER_TEMPLATE_REL_FILEPATH="${WRAPPER_GENERATOR_DIRNAME}/kurtosis.template.sh"
WRAPPER_OUTPUT_REL_FILEPATH="${BUILD_DIRNAME}/kurtosis.sh"

INITIALIZER_DIRNAME="initializer"

INITIALIZER_IMAGE="${DOCKER_ORG}/${REPO_BASE}_initializer"

INTERNAL_TESTSUITE_IMAGE="${DOCKER_ORG}/${REPO_BASE}_internal-testsuite"

# ---------------------- Interactive  -----------------------------------------------------
CLI_DIRPATH="cli"
CLI_BINARY_OUTPUT_REL_FILEPATH="${BUILD_DIRNAME}/cli"   # Relative to repo root
JAVASCRIPT_REPL_DIRNAME="javascript_cli_image"
JAVASCRIPT_REPL_IMAGE="${DOCKER_ORG}/javascript-interactive-repl"
