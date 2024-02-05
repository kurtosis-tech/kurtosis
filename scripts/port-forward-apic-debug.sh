#!/usr/bin/env bash
# 2021-07-08 WATERMARK, DO NOT REMOVE - This script was generated from the Kurtosis Bash script template

set -euo pipefail   # Bash "strict mode"

# ==================================================================================================
#                                             Constants
# ==================================================================================================
APIC_CONTAINER_K8S_LABEL="kurtosistech.com/resource-type=api-container"
# DO NOT CHANGE THIS VALUE
APIC_DEBUG_SERVER_PORT=50103

KURTOSIS_KUBERNETES_NAMESPACE_PREFIX="kt"

# ==================================================================================================
#                                       Arg Parsing & Validation
# ==================================================================================================
show_helptext_and_exit() {
    echo "Usage: $(basename "${0}") enclave_name"
    echo ""
    echo "  enclave_name     The APIC's enclave name where you want to connect the debugger"
    echo ""
    exit 1
}

# Raw parsing of the arguments
enclave_name="${1:-}"

if [ -z "${enclave_name}" ]; then
  echo "Error: No enclave name provided, please pass a valid enclave name as the first argument." >&2
  show_helptext_and_exit
fi


# ==================================================================================================
#                                             Main Logic
# ==================================================================================================
apic_namespace_name="${KURTOSIS_KUBERNETES_NAMESPACE_PREFIX}-${enclave_name}"

get_apic_pod_name_cmd="kubectl get pods -l="${APIC_CONTAINER_K8S_LABEL}" -n ${apic_namespace_name} | cut -d ' ' -f 1  | tail -n +2"

# execute the port forward to the debug server inside the APIC's container
kubectl port-forward -n "${apic_namespace_name}" $(eval "${get_apic_pod_name_cmd}") ${APIC_DEBUG_SERVER_PORT}:${APIC_DEBUG_SERVER_PORT}
