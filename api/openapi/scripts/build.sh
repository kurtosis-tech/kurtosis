#!/usr/bin/env bash
# 2021-07-08 WATERMARK, DO NOT REMOVE - This script was generated from the Kurtosis Bash script template

set -euo pipefail   # Bash "strict mode"
script_dirpath="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"
api_root_dirpath="$(dirname "${script_dirpath}")"

echo "Generating data models for REST API "
oapi-codegen --config="$api_root_dirpath/generators/api_types.cfg.yaml" "$api_root_dirpath/specs/kurtosis_api.yaml"

echo "Generating server code for REST API "
oapi-codegen --config="$api_root_dirpath/generators/engine_server.cfg.yaml" "$api_root_dirpath/specs/kurtosis_api.yaml"
oapi-codegen --config="$api_root_dirpath/generators/core_server.cfg.yaml" "$api_root_dirpath/specs/kurtosis_api.yaml"
oapi-codegen --config="$api_root_dirpath/generators/websocket_server.cfg.yaml" "$api_root_dirpath/specs/kurtosis_api.yaml"