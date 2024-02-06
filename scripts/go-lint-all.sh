#!/usr/bin/env bash

set -euo pipefail   # Bash "strict mode"
script_dirpath="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"
root_dirpath="$(dirname "${script_dirpath}")"

find $root_dirpath -type f -name 'go.mod' -exec sh -c 'dir=$(dirname "{}") && cd "$dir" && echo "$dir" && golangci-lint run' \;