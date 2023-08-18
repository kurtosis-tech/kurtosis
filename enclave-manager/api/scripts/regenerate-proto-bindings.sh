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
#api_typescript_proto_generated_rel_dir="typescript/src/generated"

# ==================================================================================================
#                                             Main Logic
# ==================================================================================================

# Dependencies from other modules
api_golang_engine="${repo_root_dirpath}/api/protobuf/engine"

api_proto_abs_dir="${root_dirpath}/${api_proto_rel_dir}"
api_golang_proto_generated_abs_dir="${repo_root_dirpath}"
api_go_mod_abs_file="${root_dirpath}/${api_go_mod_rel_file}"
api_golang_module="github.com/kurtosis-tech/kurtosis"

# TODO: we should find a way to pull the monorepo "protobuf-bindings-generator.sh" to simplify all this
protoc \
  -I="${api_proto_abs_dir}" \
  -I="${api_golang_engine}" \
  --go_out="${api_golang_proto_generated_abs_dir}" \
  --go-grpc_out="${api_golang_proto_generated_abs_dir}" \
  --go_opt=module="${api_golang_module}" \
  --go-grpc_opt=module="${api_golang_module}" \
  --go-grpc_opt=require_unimplemented_servers=false \
  --connect-go_out="${api_golang_proto_generated_abs_dir}" \
  --connect-go_opt=module="${api_golang_module}" \
  "${api_proto_abs_dir}"/*.proto
