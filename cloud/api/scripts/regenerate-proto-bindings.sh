#!/usr/bin/env bash
# 2021-07-08 WATERMARK, DO NOT REMOVE - This script was generated from the Kurtosis Bash script template

set -euo pipefail   # Bash "strict mode"
script_dirpath="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"
root_dirpath="$(dirname "${script_dirpath}")"
repo_root_dirpath="$(dirname "${root_dirpath}")"

# ==================================================================================================
#                                             Constants
# ==================================================================================================
GO_MOD_FILE_MODULE_KEYWORD="module"

# protobuf
api_proto_rel_dir="protobuf"

# Golang
api_golang_proto_generated_rel_dir="golang"
api_go_mod_rel_file="golang/go.mod"

# Typescript
#api_typescript_proto_generated_rel_dir="typescript/src/generated"

# ==================================================================================================
#                                             Main Logic
# ==================================================================================================

api_proto_abs_dir="${root_dirpath}/${api_proto_rel_dir}"
api_google_dependency_abs_dir="${api_proto_abs_dir}/google"

api_golang_proto_generated_abs_dir="${root_dirpath}/${api_golang_proto_generated_rel_dir}"
api_go_mod_abs_file="${root_dirpath}/${api_go_mod_rel_file}"
api_golang_module="$(grep "^${GO_MOD_FILE_MODULE_KEYWORD}" "${api_go_mod_abs_file}" | awk '{print $2}')"

#api_typescript_proto_generated_abs_dir="${root_dirpath}/${api_typescript_proto_generated_rel_dir}"


cd "${root_dirpath}"

# TODO: we should find a way to pull the monorepo "protobuf-bindings-generator.sh" to simplify all this
protoc \
  -I="${api_proto_abs_dir}" \
  --go_out="${api_golang_proto_generated_abs_dir}" \
  --go-grpc_out="${api_golang_proto_generated_abs_dir}" \
  --go_opt=module="${api_golang_module}" \
  --go-grpc_opt=module="${api_golang_module}" \
  --go-grpc_opt=require_unimplemented_servers=false \
  "${api_proto_abs_dir}"/*.proto

#grpc_tools_node_protoc \
#  -I="${api_proto_abs_dir}" \
#  -I="${api_google_dependency_abs_dir}" \
#  "--js_out=import_style=commonjs,binary:${api_typescript_proto_generated_abs_dir}" \
#  `# NOTE: we pass the grpc_js option to generate code using '@grpc/grpc-js', as the old 'grpc' package is deprecated` \
#  "--grpc_out=grpc_js:${api_typescript_proto_generated_abs_dir}" \
#  "--plugin=protoc-gen-grpc=$(which grpc_tools_node_protoc_plugin)" \
#  `# NOTE: we pass mode=grpc-js to get Typescript definition files that use '@grpc/grpc-js' rather than 'grpc' `\
#  "--ts_out=service=grpc-node,mode=grpc-js:${api_typescript_proto_generated_abs_dir}" \
#  `# NOTE: those 2 google files are need for the google.api.http annotations inside client and server protos` \
#  "${api_proto_abs_dir}"/google/api/http.proto \
#  "${api_proto_abs_dir}"/google/api/annotations.proto \
#  "${api_proto_abs_dir}"/*.proto

#grpc_tools_node_protoc \
#  -I="${api_proto_abs_dir}" \
#  -I="${api_google_dependency_abs_dir}" \
#  "--js_out=import_style=commonjs:${api_typescript_proto_generated_abs_dir}" \
#  "--grpc-web_out=import_style=commonjs+dts,mode=grpcwebtext:${api_typescript_proto_generated_abs_dir}" \
#  `# NOTE: those 2 google files are need for the google.api.http annotations inside client and server protos` \
#  "${api_google_dependency_abs_dir}"/api/http.proto \
#  "${api_google_dependency_abs_dir}"/api/annotations.proto \
#  "${api_proto_abs_dir}"/*.proto
