#!/usr/bin/env bash

set -euo pipefail   # Bash "strict mode"
script_dirpath="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"
root_dirpath="$(dirname "${script_dirpath}")"

GOLINT_VERSION="v2.11.3"

find $root_dirpath -type f -name 'go.mod' -exec sh -c 'dir=$(dirname "{}") && cd "$dir" && echo "$dir" && GOTOOLCHAIN=local go run github.com/golangci/golangci-lint/v2/cmd/golangci-lint@'"$GOLINT_VERSION"' run' \;