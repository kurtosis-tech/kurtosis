#!/usr/bin/env bash
# 2021-07-08 WATERMARK, DO NOT REMOVE - This script was generated from the Kurtosis Bash script template

set -euo pipefail   # Bash "strict mode"
script_dirpath="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"
root_dirpath="$(dirname "${script_dirpath}")"
monorepo_root_dirpath="$(dirname "${root_dirpath}")"

go_mod_file_name="go.mod"
protobuf_generate_script="${monorepo_root_dirpath}/api/scripts/protobuf-bindings-generator.sh"
proto_files_dir="${root_dirpath}/api/protobuf/"
golang_api_dir="${root_dirpath}/api/golang/generated"
golang_language_name="golang"

# ==================================================================================================
#                                             Main Logic
# ==================================================================================================
cd "${root_dirpath}"

# generate protobuf golang files using protobuf-bindings-generator script
GO_MOD_FILEPATH="${root_dirpath}/${go_mod_file_name}" ${protobuf_generate_script} "${proto_files_dir}" "${golang_api_dir}" "${golang_language_name}"
