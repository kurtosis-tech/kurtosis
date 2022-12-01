#!/usr/bin/env bash

set -e
rm -rf completions
mkdir completions

go build -o "${CLI_BINARY_FILENAME}" main.go

for sh in bash zsh fish; do
	"./${CLI_BINARY_FILENAME}" completion "$sh" > "completions/kurtosis.$sh"
done