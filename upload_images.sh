#!/bin/bash

minikube image load kurtosistech/engine:$1
minikube image load kurtosistech/core:$1
minikube image load kurtosistech/files-artifacts-expander:$1