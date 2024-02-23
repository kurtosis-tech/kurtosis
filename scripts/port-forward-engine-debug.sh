#!/usr/bin/env bash
# 2021-07-08 WATERMARK, DO NOT REMOVE - This script was generated from the Kurtosis Bash script template

set -euo pipefail   # Bash "strict mode"

# ==================================================================================================
#                                             Constants
# ==================================================================================================

ENGINE_CONTAINER_K8S_LABEL="kurtosistech.com/resource-type=kurtosis-engine"
# DO NOT CHANGE THIS VALUE
ENGINE_DEBUG_SERVER_PORT=50102

# ==================================================================================================
#                                             Main Logic
# ==================================================================================================

get_engine_namespace_name_cmd="kubectl get ns -l="${ENGINE_CONTAINER_K8S_LABEL}" | cut -d ' ' -f 1  | tail -n +2"
get_engine_pod_name_cmd="kubectl get pods -l="${ENGINE_CONTAINER_K8S_LABEL}" -A | cut -d ' ' -f 1  | tail -n +2"

# execute the port forward to the debug server inside the engine's container
kubectl port-forward -n $(eval "${get_engine_namespace_name_cmd}") $(eval "${get_engine_pod_name_cmd}") ${ENGINE_DEBUG_SERVER_PORT}:${ENGINE_DEBUG_SERVER_PORT}
