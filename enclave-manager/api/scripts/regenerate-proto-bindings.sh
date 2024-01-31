#!/usr/bin/env bash
# 2021-07-08 WATERMARK, DO NOT REMOVE - This script was generated from the Kurtosis Bash script template

set -euo pipefail # Bash "strict mode"
script_dirpath="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"
root_dirpath="$(dirname "${script_dirpath}")"
module_root_dirpath="$(dirname "${root_dirpath}")"
repo_root_dirpath="$(dirname "${module_root_dirpath}")"

# ==================================================================================================
#                                             Constants
# ==================================================================================================
GO_MOD_FILE_MODULE_KEYWORD="module"

# protobuf
api_proto_rel_dir="protobuf"

# Golang
api_go_mod_rel_file="golang/go.mod"

# Typescript
api_typescript_rel_dir="typescript"

NODE_ES_TOOLS_PROTOC_BIN_FILENAME="protoc-gen-es"

NODE_CONNECT_ES_TOOLS_PROTOC_BIN_FILENAME="protoc-gen-connect-es"

# ==================================================================================================
#                                             Main Logic
# ==================================================================================================

if ! node_es_protoc_bin_filepath="$(which "${NODE_ES_TOOLS_PROTOC_BIN_FILENAME}")"; then
    echo "Error: Couldn't find Node gRPC tools protoc plugin binary '${NODE_ES_TOOLS_PROTOC_BIN_FILENAME}' on the PATH" >&2
    return 1
fi
if [ -z "${node_es_protoc_bin_filepath}" ]; then
    echo "Error: Got an empty filepath when looking for the Node gRPC tools protoc plugin binary '${NODE_ES_TOOLS_PROTOC_BIN_FILENAME}'" >&2
    return 1
fi

if ! node_connect_es_protoc_bin_filepath="$(which "${NODE_CONNECT_ES_TOOLS_PROTOC_BIN_FILENAME}")"; then
    echo "Error: Couldn't find Node gRPC tools protoc plugin binary '${NODE_CONNECT_ES_TOOLS_PROTOC_BIN_FILENAME}' on the PATH" >&2
    return 1
fi
if [ -z "${node_connect_es_protoc_bin_filepath}" ]; then
    echo "Error: Got an empty filepath when looking for the Node gRPC tools protoc plugin binary '${NODE_CONNECT_ES_TOOLS_PROTOC_BIN_FILENAME}'" >&2
    return 1
fi
# Dependencies from other modules
api_golang_engine="${repo_root_dirpath}/api/protobuf/engine"
api_golang_core="${repo_root_dirpath}/api/protobuf/core"

api_typescript_abs_dir="${root_dirpath}/${api_typescript_rel_dir}"
api_proto_abs_dir="${root_dirpath}/${api_proto_rel_dir}"
api_golang_proto_generated_abs_dir="${repo_root_dirpath}"
api_go_mod_abs_file="${root_dirpath}/${api_go_mod_rel_file}"
api_golang_module="github.com/kurtosis-tech/kurtosis"

# TODO: we should find a way to pull the monorepo "protobuf-bindings-generator.sh" to simplify all this
protoc \
  -I="${api_proto_abs_dir}" \
  -I="${api_golang_engine}" \
  -I="${api_golang_core}" \
  --go_out="${api_golang_proto_generated_abs_dir}" \
  --go-grpc_out="${api_golang_proto_generated_abs_dir}" \
  --go_opt=module="${api_golang_module}" \
  --go-grpc_opt=module="${api_golang_module}" \
  --go-grpc_opt=require_unimplemented_servers=false \
  --connect-go_out="${api_golang_proto_generated_abs_dir}" \
  --connect-go_opt=module="${api_golang_module}" \
  --plugin=protoc-gen-es="${node_es_protoc_bin_filepath}" \
  --es_out="${api_typescript_abs_dir}/src/" \
  --es_opt=target=ts \
  --plugin=protoc-gen-connect-es="${node_connect_es_protoc_bin_filepath}" \
  --connect-es_out="${api_typescript_abs_dir}/src/" \
  --connect-es_opt=target=ts \
  "${api_proto_abs_dir}"/*.proto