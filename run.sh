#!/bin/bash

git add .
git commit -m"debug"

kurtosis clean -a
kurtosis engine stop

./scripts/build.sh
kurtosis run --cli-log-level=debug main.star