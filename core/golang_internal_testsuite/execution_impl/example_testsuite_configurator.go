/*
 * Copyright (c) 2021 - present Kurtosis Technologies Inc.
 * All Rights Reserved.
 */

package execution_impl

import (
	"encoding/json"
	"github.com/kurtosis-tech/kurtosis-testsuite-api-lib/golang/lib/testsuite"
	"github.com/kurtosis-tech/kurtosis/golang_internal_testsuite/testsuite_impl"
	"github.com/palantir/stacktrace"
	"github.com/sirupsen/logrus"
	"strings"
)

type InternalTestsuiteConfigurator struct {}

func NewInternalTestsuiteConfigurator() *InternalTestsuiteConfigurator {
	return &InternalTestsuiteConfigurator{}
}

func (t InternalTestsuiteConfigurator) SetLogLevel(logLevelStr string) error {
	level, err := logrus.ParseLevel(logLevelStr)
	if err != nil {
		return stacktrace.Propagate(err, "An error occurred parsing loglevel string '%v'", logLevelStr)
	}
	logrus.SetLevel(level)
	logrus.SetFormatter(&logrus.TextFormatter{
		ForceColors:   true,
		FullTimestamp: true,
	})
	return nil
}

func (t InternalTestsuiteConfigurator) ParseParamsAndCreateSuite(paramsJsonStr string) (testsuite.TestSuite, error) {
	paramsJsonBytes := []byte(paramsJsonStr)
	var args InternalTestsuiteArgs
	if err := json.Unmarshal(paramsJsonBytes, &args); err != nil {
		return nil, stacktrace.Propagate(err, "An error occurred deserializing the testsuite params JSON")
	}

	if err := validateArgs(args); err != nil {
		return nil, stacktrace.Propagate(err, "An error occurred validating the deserialized testsuite params")
	}

	suite := testsuite_impl.NewInternalTestsuite(args.ApiServiceImage, args.DatastoreServiceImage)
	return suite, nil
}

func validateArgs(args InternalTestsuiteArgs) error {
	if strings.TrimSpace(args.ApiServiceImage) == "" {
		return stacktrace.NewError("API service image is empty")
	}
	if strings.TrimSpace(args.DatastoreServiceImage) == "" {
		return stacktrace.NewError("Datastore service image is empty")
	}
	return nil
}
