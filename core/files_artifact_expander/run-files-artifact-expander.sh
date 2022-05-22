#!/usr/bin/env bash
#
# Copyright (c) 2022 - present Kurtosis Technologies Inc.
# All Rights Reserved.
#

# This script is intended to run onside the Docker Container

set -euo pipefail   # Bash "strict mode"
script_dirpath="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"

echo "starting files artifact expander"
"${script_dirpath}"/files-artifact-expander