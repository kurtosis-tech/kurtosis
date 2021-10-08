/*
 * Copyright (c) 2021 - present Kurtosis Technologies Inc.
 * All Rights Reserved.
 */

package advanced_network_test

import (
	"github.com/kurtosis-tech/kurtosis-client/golang/lib/networks"
	"github.com/kurtosis-tech/kurtosis-cli/golang_internal_testsuite/networks_impl"
	"github.com/kurtosis-tech/kurtosis-testsuite-api-lib/golang/lib/testsuite"
	"github.com/palantir/stacktrace"
	"github.com/sirupsen/logrus"
)

const (
	testPersonId = 46
)

type AdvancedNetworkTest struct {
	datastoreServiceImage string
	apiServiceImage string
}

func NewAdvancedNetworkTest(datastoreServiceImage string, apiServiceImage string) *AdvancedNetworkTest {
	return &AdvancedNetworkTest{datastoreServiceImage: datastoreServiceImage, apiServiceImage: apiServiceImage}
}

func (test *AdvancedNetworkTest) Configure(builder *testsuite.TestConfigurationBuilder) {
	builder.WithSetupTimeoutSeconds(60).WithRunTimeoutSeconds(60)
}

func (test *AdvancedNetworkTest) Setup(networkCtx *networks.NetworkContext) (networks.Network, error) {
	network := networks_impl.NewTestNetwork(networkCtx, test.datastoreServiceImage, test.apiServiceImage)
	// Note how setup logic has been pushed into a custom Network implementation, to make test-writing easy
	if err := network.SetupDatastoreAndTwoApis(); err != nil {
		return nil, stacktrace.Propagate(err, "An error occurred setting up the network")
	}
	return network, nil
}

func (test *AdvancedNetworkTest) Run(network networks.Network) error {
	castedNetwork := network.(*networks_impl.TestNetwork)
	personModifierClient, err := castedNetwork.GetPersonModifyingApiClient()
	if err != nil {
		return stacktrace.Propagate(err, "An error occurred getting the person-modifying API client")
	}
	personRetrieverClient, err := castedNetwork.GetPersonRetrievingApiClient()
	if err != nil {
		return stacktrace.Propagate(err, "An error occurred getting the person-retrieving API client")
	}

	logrus.Infof("Adding test person via person-modifying API client...")
	if err := personModifierClient.AddPerson(testPersonId); err != nil {
		return stacktrace.Propagate(err, "An error occurred adding test person")
	}
	logrus.Info("Test person added")

	logrus.Infof("Incrementing test person's number of books read through person-modifying API client...")
	if err := personModifierClient.IncrementBooksRead(testPersonId); err != nil {
		return stacktrace.Propagate(err, "An error occurred incrementing the number of books read")
	}
	logrus.Info("Incremented number of books read")

	logrus.Info("Retrieving test person to verify number of books read person-retrieving API client...")
	person, err := personRetrieverClient.GetPerson(testPersonId)
	if err != nil {
		return stacktrace.Propagate(err, "An error occurred getting the test person")
	}
	logrus.Info("Retrieved test person")

	if person.BooksRead != 1 {
		return stacktrace.NewError(
			"Expected number of books read to be incremented, but was '%v'",
			person.BooksRead,
		)
	}
	return nil
}
