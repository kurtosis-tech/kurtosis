#!/usr/bin/env bash
#
# Copyright (c) 2022 - present Kurtosis Technologies Inc.
# All Rights Reserved.
#

# This script is intended to run onside the Docker Container

# 2021-07-08 WATERMARK, DO NOT REMOVE - This script was generated from the Kurtosis Bash script template

set -euo pipefail   # Bash "strict mode"
script_dirpath="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"


# ==================================================================================================
#                                             Constants
# ==================================================================================================

COMMANDS_TO_RUN=(
    "run-envoy"
    "run-api-container"
)

# ==================================================================================================
#                                             Main Logic
# ==================================================================================================

PIDS_TO_WAIT_FOR=()

function run-api-container() {
    echo "starting api container"
    "${script_dirpath}"/api-container
}

function run-envoy() {
    echo "running run-envoy func"
    envoy -c /etc/envoy/envoy.yaml
}

for command_to_run in "${COMMANDS_TO_RUN[@]}"; do
    "${command_to_run}" &
    command_pid="${!}"
    PIDS_TO_WAIT_FOR+=("${command_pid}")
    echo "Launched command '${command_to_run}' with PID '${command_pid}'"
done

did_errors_occur="false"

for pid in "${PIDS_TO_WAIT_FOR[@]}"; do
    if wait "${pid}"; then
        echo "PID '${pid}' exited successfully"
    else
        did_errors_occur="true"
        echo "PID '${pid}' errored"
    fi
done

if "${did_errors_occur}"; then
    echo "Error: One or more errors occurred running the API container & Envoy proxy" >&2
    exit 1
fi

echo "The API container & Envoy proxy finished successfully"