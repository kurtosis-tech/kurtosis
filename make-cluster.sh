#!/usr/bin/env bash
set -xou pipefail
CLUSTER_NAME=${CLUSTER_NAME:-kurtosis}
REGISTRY_NAME=${REGISTRY_NAME:-registry} # k3d adds the "k3d-" prefix so yeah don't bother
REGISTRY_DOCKER_NETWORK_PORT=${REGISTRY_DOCKER_NETWORK_PORT:-5000}
REGISTRY_HOST_PORT=${REGISTRY_HOST_PORT:-5151}
#REGISTRY_BIND_ADDR=${REGISTRY_BIND_ADDR:-127.0.0.0:$REGISTRY_HOST_PORT}
REGISTRY_BIND_ADDR=${REGISTRY_BIND_ADDR:-0.0.0.0:$REGISTRY_HOST_PORT}

if ! k3d registry list -ojson | jq -e --arg name "k3d-${REGISTRY_NAME}" 'any(.[]; .name == $name)'; then
	# Yep, you also use port if you specify the bind addr ;)
	if ! k3d registry create ${REGISTRY_NAME} \
		--port ${REGISTRY_BIND_ADDR} \
		--default-network ${CLUSTER_NAME} \
		--default-network bridge \
	; then
		echo "Failed creating registry ${REGISTRY_NAME}. Try creating it manually with 'k3d registry create ${REGISTRY_NAME}' before re-running this script"
		exit 1
	else
		echo "Successfully created registry `${REGISTRY_NAME}`"
	fi
else
	echo "WARNING: Registry already exists, assuming it has the correct configuration "
	echo "you may need to manually delete it with 'k3d registry delete ${REGISTRY_NAME}'"
fi

k3d cluster create ${CLUSTER_NAME} \
	--agents 3 \
	--k3s-arg '--kubelet-arg=eviction-hard=imagefs.available<1%,nodefs.available<1%@agent:*' \
	--k3s-arg '--kubelet-arg=eviction-minimum-reclaim=imagefs.available=1%,nodefs.available=1%@agent:*' \
	--registry-use ${REGISTRY_NAME}:${REGISTRY_DOCKER_NETWORK_PORT} \
	--network bridge \
	--network ${CLUSTER_NAME}
