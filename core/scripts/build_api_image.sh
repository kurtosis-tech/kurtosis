#!/bin/bash

set -euo pipefail
script_dirpath="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"

DOCKER_ORG="kurtosistech"
REPO_BASE="kurtosis-core"
API_REPO="${REPO_BASE}_api"
LATEST_TAG="latest"

root_dirpath="$(dirname "${script_dirpath}")"

tag="${DOCKER_ORG}/${API_REPO}:${LATEST_TAG}"
docker build -t "${tag}" "${root_dirpath}" -f "${root_dirpath}/api_container/Dockerfile"

# TODO REPLACE THIS PART
container_ip="172.17.0.0.2"
docker run --mount type=bind,source=/var/run/docker.sock,target=/var/run/docker.sock -p 8080:8080 --env NETWORK_ID=b453ce4bac01 --env SUBNET_MASK=172.17.0.0/16 --env CONTAINER_IP=${container_ip} --env GATEWAY_IP=172.17.0.1 --env LOG_FILEPATH=/tmp/logfile.txt --env LOG_LEVEL=trace --ip=${container_ip} "${tag}"
