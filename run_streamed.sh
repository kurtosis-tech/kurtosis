#!/bin/bash

git add .
git commit -m"debug"

kurtosis clean -a
kurtosis engine stop

./scripts/build.sh
./cli/cli/scripts/launch-cli.sh run --cli-log-level=debug main.star