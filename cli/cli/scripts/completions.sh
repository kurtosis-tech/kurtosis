#!/usr/bin/env bash

set -e
if [ -z "${CLI_BINARY_FILENAME}" ]; then
    echo "Environment variable CLI_BINARY_FILENAME must not be empty" >&2
    exit 1
fi

rm -rf completions
mkdir completions

go build -o "${CLI_BINARY_FILENAME}" main.go

for sh in bash zsh fish; do
	"./${CLI_BINARY_FILENAME}" completion "$sh" > "completions/kurtosis.$sh"
done

rm "${CLI_BINARY_FILENAME}"
