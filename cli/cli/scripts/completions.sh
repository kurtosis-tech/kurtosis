#!/usr/bin/env bash

set -e
rm -rf completions
mkdir completions
for sh in bash zsh fish; do
	go run main.go completion "$sh" > "completions/kurtosis.$sh"
done