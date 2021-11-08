/*
 * Copyright (c) 2021 - present Kurtosis Technologies Inc.
 * All Rights Reserved.
 */

package basic_datastore_test

import (
	"context"
	"github.com/kurtosis-tech/example-datastore-server/api/golang/datastore_rpc_api_bindings"
	"github.com/kurtosis-tech/kurtosis-cli/golang_internal_testsuite/client_helpers"
	"github.com/kurtosis-tech/kurtosis-client/golang/lib/networks"
	"github.com/kurtosis-tech/kurtosis-client/golang/lib/services"
	"github.com/kurtosis-tech/kurtosis-testsuite-api-lib/golang/lib/testsuite"
	"github.com/kurtosis-tech/stacktrace"
	"github.com/sirupsen/logrus"
)

const (
	datastoreServiceId services.ServiceID = "datastore"
	testKey                               = "test-key"
	testValue                             = "test-value"

	waitForStartupDelayMilliseconds = 1000
	waitForStartupMaxPolls          = 15
)

type BasicDatastoreTest struct {
	datastoreImage string
}

func NewBasicDatastoreTest(datastoreImage string) *BasicDatastoreTest {
	return &BasicDatastoreTest{datastoreImage: datastoreImage}
}

func (test BasicDatastoreTest) Configure(builder *testsuite.TestConfigurationBuilder) {
	builder.WithSetupTimeoutSeconds(60).WithRunTimeoutSeconds(60)
}

func (test BasicDatastoreTest) Setup(networkCtx *networks.NetworkContext) (networks.Network, error) {
	ctx := context.Background()

	datastoreContainerConfigSupplier := client_helpers.GetDatastoreContainerConfigSupplier(test.datastoreImage)

	serviceContext, hostPortBindings, err := networkCtx.AddService(datastoreServiceId, datastoreContainerConfigSupplier)
	if err != nil {
		return nil, stacktrace.Propagate(err, "An error occurred adding the datastore service")
	}

	datastoreClient, datastoreClientConnCloseFunc, err := client_helpers.NewDatastoreClient(serviceContext.GetIPAddress())
	if err != nil {
		return nil, stacktrace.Propagate(err, "An error occurred creating a new datastore client for service with ID '%v' and IP address '%v'", datastoreServiceId, serviceContext.GetIPAddress())
	}
	defer func() {
		if err := datastoreClientConnCloseFunc(); err != nil {
			logrus.Warnf("We tried to close the datastore client, but doing so threw an error:\n%v", err)
		}
	}()

	err = client_helpers.WaitForHealthy(ctx, datastoreClient, waitForStartupMaxPolls, waitForStartupDelayMilliseconds)
	if err != nil {
		return nil, stacktrace.Propagate(err, "An error occurred waiting for the datastore service to become available")
	}

	logrus.Infof("Added datastore service with host port bindings: %+v", hostPortBindings)
	return networkCtx, nil
}

func (test BasicDatastoreTest) Run(network networks.Network) error {
	ctx := context.Background()

	// Necessary because Go doesn't have generics
	castedNetwork := network.(*networks.NetworkContext)

	serviceContext, err := castedNetwork.GetServiceContext(datastoreServiceId)
	if err != nil {
		return stacktrace.Propagate(err, "An error occurred getting the datastore service info")
	}

	datastoreClient, datastoreClientConnCloseFunc, err := client_helpers.NewDatastoreClient(serviceContext.GetIPAddress())
	if err != nil {
		return stacktrace.Propagate(err, "An error occurred creating a new datastore client for service with ID '%v' and IP address '%v'", datastoreServiceId, serviceContext.GetIPAddress())
	}
	defer func() {
		if err := datastoreClientConnCloseFunc(); err != nil {
			logrus.Warnf("We tried to close the datastore client, but doing so threw an error:\n%v", err)
		}
	}()

	logrus.Infof("Verifying that key '%v' doesn't already exist...", testKey)
	existsArgs := &datastore_rpc_api_bindings.ExistsArgs{
		Key: testKey,
	}
	existsResponse, err := datastoreClient.Exists(ctx, existsArgs)
	if err != nil {
		return stacktrace.Propagate(err, "An error occurred checking if the test key exists")
	}
	if existsResponse.GetExists() {
		return stacktrace.NewError("Test key should not exist yet")
	}
	logrus.Infof("Confirmed that key '%v' doesn't already exist", testKey)

	logrus.Infof("Inserting value '%v' at key '%v'...", testKey, testValue)
	upsertArgs := &datastore_rpc_api_bindings.UpsertArgs{
		Key:   testKey,
		Value: testValue,
	}
	if _, err = datastoreClient.Upsert(ctx, upsertArgs); err != nil {
		return stacktrace.Propagate(err, "An error occurred upserting the test key")
	}
	logrus.Infof("Inserted value successfully")

	logrus.Infof("Getting the key we just inserted to verify the value...")
	getArgs := &datastore_rpc_api_bindings.GetArgs{
		Key: testKey,
	}
	getResponse, err := datastoreClient.Get(ctx, getArgs)
	if err != nil {
		return stacktrace.Propagate(err, "An error occurred getting the test key after upload")
	}
	if getResponse.GetValue() != testValue {
		return stacktrace.NewError("Returned value '%v' != test value '%v'", getResponse.GetValue(), testValue)
	}
	logrus.Info("Value verified")
	return nil
}
