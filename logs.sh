#!/bin/zsh

cd /Users/tewodrosmitiku/Desktop/kurtosis
kurtosis clean -a
kurtosis engine stop
./cli/cli/scripts/launch-cli.sh engine start --version=$1
./cli/cli/scripts/launch-cli.sh run github.com/kurtosis-tech/mongodb-package

