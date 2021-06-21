#!/usr/bin/env bash

# This script regenerates Go bindings corresponding to all .proto files inside this project
# It requires the Golang Protobuf extension to the 'protoc' compiler, as well as the Golang gRPC extension

set -euo pipefail
script_dirpath="$(cd "$(dirname "${BASH_SOURCE[0]}")"; pwd)"
root_dirpath="$(dirname "${script_dirpath}")"

# ================================ CONSTANTS =======================================================
GENERATOR_SCRIPT_FILENAME="generate-protobuf-bindings.sh"  # Must be on the PATH
# The name of the directory to contain generated Protobuf Go bindings
BINDINGS_OUTPUT_DIRNAME="bindings"
# Dirpaths where protobuf files live
# The bindings for the Protobuf files will be generated inside a 'bindings' directory in this directory
PROTOBUF_DIRPATHS=(
    "${root_dirpath}/test_suite/test_suite_rpc_api"
)

# =============================== MAIN LOGIC =======================================================
for input_dirpath in "${PROTOBUF_DIRPATHS[@]}"; do
    output_dirpath="${input_dirpath}/${BINDINGS_OUTPUT_DIRNAME}"
    if ! GO_MOD_FILEPATH="${root_dirpath}/go.mod" "${GENERATOR_SCRIPT_FILENAME}" "${input_dirpath}" "${output_dirpath}" golang; then
        echo "Error: An error occurred generating bindings to directory '${output_dirpath}' from Protobuf files in directory '${input_dirpath}'" >&2
        exit 1
    fi
    echo "Successfully generated Protobuf bindings for files in '${input_dirpath}' to output directory '${output_dirpath}'"
done
