/*
 * Copyright (c) 2021 - present Kurtosis Technologies Inc.
 * All Rights Reserved.
 */

package lambda_store

import (
	"context"
	"github.com/docker/go-connections/nat"
	"github.com/kurtosis-tech/container-engine-lib/lib/docker_manager"
	"github.com/kurtosis-tech/kurtosis-lambda-api-lib/golang/kurtosis_lambda_rpc_api_bindings"
	"github.com/kurtosis-tech/kurtosis/api_container/server/lambda_store/lambda_launcher"
	"github.com/kurtosis-tech/kurtosis/api_container/server/lambda_store/lambda_store_types"
	"github.com/palantir/stacktrace"
	"net"
	"strings"
	"sync"
)

type lambdaInfo struct {
	containerId           string
	ipAddr                net.IP
	client                kurtosis_lambda_rpc_api_bindings.LambdaServiceClient
	hostPortBinding *nat.PortBinding
}

type LambdaStore struct {
	isDestroyed bool

	mutex *sync.Mutex

	dockerManager *docker_manager.DockerManager

	// lambda_id -> IP addr, container ID, etc.
	lambdas map[lambda_store_types.LambdaID]lambdaInfo

	lambdaLauncher *lambda_launcher.LambdaLauncher
}

func NewLambdaStore(dockerManager *docker_manager.DockerManager, lambdaLauncher *lambda_launcher.LambdaLauncher) *LambdaStore {
	return &LambdaStore{
		isDestroyed:    false,
		mutex:          &sync.Mutex{},
		dockerManager:  dockerManager,
		lambdas:        map[lambda_store_types.LambdaID]lambdaInfo{},
		lambdaLauncher: lambdaLauncher,
	}
}

func (store *LambdaStore) LoadLambda(ctx context.Context, lambdaId lambda_store_types.LambdaID, containerImage string, serializedParams string) error {
	store.mutex.Lock()
	defer store.mutex.Unlock()
	if store.isDestroyed {
		return stacktrace.NewError("Cannot load Lambda; the Lambda store is destroyed")
	}

	if _, found := store.lambdas[lambdaId]; found {
		return stacktrace.NewError("Lambda ID '%v' already exists in the lambda map", lambdaId)
	}

	// NOTE: We don't use module host port bindings for now; we could expose them in the future if it's useful
	containerId, containerIpAddr, client, hostPortBinding, err := store.lambdaLauncher.Launch(ctx, lambdaId, containerImage, serializedParams)
	if err != nil {
		return stacktrace.Propagate(
			err,
			"An error occurred launching Lambda from container image '%v' and serialized params '%v'",
			containerImage,
			serializedParams,
		)
	}

	lambdaInfo :=  lambdaInfo{
		containerId: containerId,
		ipAddr: containerIpAddr,
		client: client,
		hostPortBinding: hostPortBinding,
	}

	store.lambdas[lambdaId] = lambdaInfo
	return nil
}

func (store *LambdaStore) UnloadLambda(ctx context.Context, lambdaId lambda_store_types.LambdaID) error {
	store.mutex.Lock()
	defer store.mutex.Unlock()
	if store.isDestroyed {
		return stacktrace.NewError("Cannot unload Lambda; the Lambda store is destroyed")
	}

	 lambdaInfo, found := store.lambdas[lambdaId]
	 if !found {
		return stacktrace.NewError("Lambda ID '%v' do not exists in the lambda map", lambdaId)
	}

	containerId := lambdaInfo.containerId
	if err := store.dockerManager.KillContainer(ctx, containerId); err != nil {
		return  stacktrace.Propagate(err, "An error occurred killing Lambda container '%v' while unloading the Lambda from the store", lambdaId)
	}

	delete(store.lambdas, lambdaId)

	return nil
}

func (store *LambdaStore) ExecuteLambda(ctx context.Context, lambdaId lambda_store_types.LambdaID, serializedParams string) (serializedResult string, resultErr error) {
	// NOTE: technically we don't need this mutex for this function since we're not modifying internal state, but we do need it to check isDestroyed
	// TODO PERF: Don't block the entire store on executing a lambda
	store.mutex.Lock()
	defer store.mutex.Unlock()
	if store.isDestroyed {
		return "", stacktrace.NewError("Cannot execute Lambda '%v'; the Lambda store is destroyed", lambdaId)
	}

	info, found := store.lambdas[lambdaId]
	if !found {
		return "", stacktrace.NewError("No Lambda '%v' exists in the Lambda store", lambdaId)
	}
	client := info.client
	args := &kurtosis_lambda_rpc_api_bindings.ExecuteArgs{ParamsJson: serializedParams}
	resp, err := client.Execute(ctx, args)
	if err != nil {
		return "", stacktrace.Propagate(err, "An error occurred calling Lambda '%v' with serialized params '%v'", lambdaId, serializedParams)
	}
	return resp.ResponseJson, nil
}

func (store *LambdaStore) GetLambdaIPAddrByID(lambdaId lambda_store_types.LambdaID) (net.IP, error) {
	store.mutex.Lock()
	defer store.mutex.Unlock()
	if store.isDestroyed {
		return nil, stacktrace.NewError("Cannot get IP address for Lambda '%v'; the Lambda store is destroyed", lambdaId)
	}

	info, found := store.lambdas[lambdaId]
	if !found {
		return nil, stacktrace.NewError("No Lambda with ID '%v' has been loaded", lambdaId)
	}
	return info.ipAddr, nil
}

func (store *LambdaStore) Destroy(ctx context.Context) error {
	store.mutex.Lock()
	defer store.mutex.Unlock()
	if store.isDestroyed {
		return stacktrace.NewError("Cannot destroy the Lambda store because it's already destroyed")
	}

	lambdaKillErrorTexts := []string{}
	for lambdaId, lambdaInfo := range store.lambdas {
		containerId := lambdaInfo.containerId
		if err := store.dockerManager.KillContainer(ctx, containerId); err != nil {
			killError := stacktrace.Propagate(err, "An error occurred killing Lambda container '%v' while destroying the Lambda store", lambdaId)
			lambdaKillErrorTexts = append(lambdaKillErrorTexts, killError.Error())
		}
	}
	store.isDestroyed = true

	if len(lambdaKillErrorTexts) > 0 {
		resultErrText := strings.Join(lambdaKillErrorTexts, "\n\n")
		return stacktrace.NewError("One or more error(s) occurred killing the Lambda containers during Lambda store destruction:\n%v", resultErrText)
	}

	return nil
}

func (store *LambdaStore) GetLambdas() map[lambda_store_types.LambdaID]bool {

	lambdaIDs := make(map[lambda_store_types.LambdaID]bool, len(store.lambdas))

	for key, _ := range store.lambdas {
		if _, ok := lambdaIDs[key]; !ok{
			lambdaIDs[key] = true
		}
	}
	return lambdaIDs
}
