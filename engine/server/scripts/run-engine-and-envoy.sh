#!/usr/bin/env bash
# 2021-07-08 WATERMARK, DO NOT REMOVE - This script was generated from the Kurtosis Bash script template

set -euo pipefail   # Bash "strict mode"
script_dirpath="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"

# ==================================================================================================
#                                             Constants
# ==================================================================================================

COMMANDS_TO_RUN=(
    "run-envoy-proxy"
    "run-kurtosis-engine"
)

# ==================================================================================================
#                                             Main Logic
# ==================================================================================================

PIDS_TO_WAIT_FOR=()

function run-kurtosis-engine() {
    echo "starting kurtosis engine..."
    "${script_dirpath}"/kurtosis-engine
}

function run-envoy-proxy() {
    echo "starting envoy proxy..."
    envoy -c /etc/envoy/envoy.yaml
}

for command_to_run in "${COMMANDS_TO_RUN[@]}"; do
    "${command_to_run}" &
    PIDS_TO_WAIT_FOR+=("${!}")
done

function cleanup() {
  echo "cleaning up before exiting..."
  for pid in "${PIDS_TO_WAIT_FOR[@]}"; do
      kill "${pid}"
      wait $!
  done
}

trap 'echo signal received!; cleanup' SIGINT SIGTERM

for pid in "${PIDS_TO_WAIT_FOR[@]}"; do
    if wait "${pid}"; then
        echo "PID '${pid}' exited successfully"
    else
        echo "PID '${pid}' errored"
    fi
done

echo "Finished running Kurtosis engine & Envoy proxy"