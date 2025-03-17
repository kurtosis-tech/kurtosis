#!/usr/bin/env bash
# 2021-07-08 WATERMARK, DO NOT REMOVE - This script was generated from the Kurtosis Bash script template

set -euo pipefail   # Bash "strict mode"
script_dirpath="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"
openapi_root_dirpath="$(dirname "${script_dirpath}")"
api_root_dirpath="$(dirname "${openapi_root_dirpath}")"

echo "Generating data models for REST API "
echo $PATH
echo $GOBIN
which oapi-codegen
oapi-codegen --config="$openapi_root_dirpath/generators/api_types.cfg.yaml" "$openapi_root_dirpath/specs/kurtosis_api.yaml"

echo "Generating server code for REST API "
oapi-codegen --config="$openapi_root_dirpath/generators/engine_server.cfg.yaml" "$openapi_root_dirpath/specs/kurtosis_api.yaml"
oapi-codegen --config="$openapi_root_dirpath/generators/core_server.cfg.yaml" "$openapi_root_dirpath/specs/kurtosis_api.yaml"
oapi-codegen --config="$openapi_root_dirpath/generators/websocket_server.cfg.yaml" "$openapi_root_dirpath/specs/kurtosis_api.yaml"

echo "Generating Go client code for REST API "
oapi-codegen --config="$openapi_root_dirpath/generators/go_client.cfg.yaml" "$openapi_root_dirpath/specs/kurtosis_api.yaml"

echo "Generating Typescript client code for REST API "
openapi-typescript "$openapi_root_dirpath/specs/kurtosis_api.yaml" -o "$api_root_dirpath/typescript/src/engine/rest_api_bindings/types.d.ts"
