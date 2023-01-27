#!/usr/bin/env bash

# This script regenerates API bindings for the various languages that this repo supports

set -euo pipefail
script_dirpath="$(cd "$(dirname "${BASH_SOURCE[0]}")"; pwd)"
api_dirpath="$(dirname "${script_dirpath}")"

# ================================ CONSTANTS =======================================================
GENERATOR_SCRIPT_FILENAME="${script_dirpath}/protobuf-bindings-generator.sh"
API_DIRNAME="api"
PROTOBUF_DIRNAME="protobuf"
GOLANG_DIRNAME="golang"
TYPESCRIPT_DIRNAME="typescript"

OUTPUT_DIRNAMES=(
    "engine"
    "core"
)

# =============================== MAIN LOGIC =======================================================
for output_dirname in "${OUTPUT_DIRNAMES[@]}"; do
    input_dirpath="${api_dirpath}/${PROTOBUF_DIRNAME}/${output_dirname}"
    rpc_bindings_dirname="kurtosis_${output_dirname}_rpc_api_bindings"

    # Golang
    go_output_dirpath="${api_dirpath}/${GOLANG_DIRNAME}/${output_dirname}/${rpc_bindings_dirname}"
    if ! GO_MOD_FILEPATH="${api_dirpath}/${GOLANG_DIRNAME}/go.mod" "${GENERATOR_SCRIPT_FILENAME}" "${input_dirpath}" "${go_output_dirpath}" golang; then
        echo "Error: An error occurred generating ${output_dirname} Go bindings in directory '${go_output_dirpath}'" >&2
        exit 1
    fi
    echo "Successfully generated ${output_dirname} Go bindings in directory '${go_output_dirpath}'"

    # TypeScript
    typescript_output_dirpath="${api_dirpath}/${TYPESCRIPT_DIRNAME}/src/${output_dirname}/${rpc_bindings_dirname}"
    if ! "${GENERATOR_SCRIPT_FILENAME}" "${input_dirpath}" "${typescript_output_dirpath}" typescript; then
        echo "Error: An error occurred generating ${output_dirname} TypeScript bindings in directory '${typescript_output_dirpath}'" >&2
        exit 1
    fi
    echo "Successfully generated ${output_dirname} TypeScript bindings in directory '${typescript_output_dirpath}'"
done
