#!/usr/bin/env bash

# This script regenerates API bindings for the various languages that this repo supports

set -euo pipefail
script_dirpath="$(cd "$(dirname "${BASH_SOURCE[0]}")"; pwd)"
root_dirpath="$(dirname "${script_dirpath}")"

# ================================ CONSTANTS =======================================================
GENERATOR_SCRIPT_FILENAME="generate-protobuf-bindings.sh"  # Must be on the PATH
API_DIRNAME="api"
PROTOBUF_DIRNAME="protobuf"
GOLANG_DIRNAME="golang"
TYPESCRIPT_DIRNAME="typescript"
RPC_BINDINGS_DIRNAME="kurtosis_engine_rpc_api_bindings"

# =============================== MAIN LOGIC =======================================================
api_dirpath="${root_dirpath}/${API_DIRNAME}"
input_dirpath="${api_dirpath}/${PROTOBUF_DIRNAME}"

# Golang
go_output_dirpath="${api_dirpath}/${GOLANG_DIRNAME}/${RPC_BINDINGS_DIRNAME}"
if ! GO_MOD_FILEPATH="${api_dirpath}/${GOLANG_DIRNAME}/go.mod" "${GENERATOR_SCRIPT_FILENAME}" "${input_dirpath}" "${go_output_dirpath}" golang; then
    echo "Error: An error occurred generating Go bindings in directory '${go_output_dirpath}'" >&2
    exit 1
fi
echo "Successfully generated Go bindings in directory '${go_output_dirpath}'"

# TypeScript
typescript_output_dirpath="${api_dirpath}/${TYPESCRIPT_DIRNAME}/src/${RPC_BINDINGS_DIRNAME}"
if ! "${GENERATOR_SCRIPT_FILENAME}" "${input_dirpath}" "${typescript_output_dirpath}" typescript; then
    echo "Error: An error occurred generating TypeScript bindings in directory '${typescript_output_dirpath}'" >&2
    exit 1
fi
echo "Successfully generated TypeScript bindings in directory '${typescript_output_dirpath}'"
