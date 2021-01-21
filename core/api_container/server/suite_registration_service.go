/*
 * Copyright (c) 2021 - present Kurtosis Technologies LLC.
 * All Rights Reserved.
 */

package server

import (
	"context"
	"github.com/kurtosis-tech/kurtosis/api_container/api/bindings"
	"github.com/palantir/stacktrace"
	"google.golang.org/protobuf/types/known/emptypb"
	"sync"
)

// Service that allows for a testsuite container to register itself exactly once
type suiteRegistrationService struct {
	mutex *sync.Mutex

	suiteRegistered bool

	// The action the newly-registered suite should take
	suiteAction bindings.SuiteAction

	suiteRegistrationChan chan interface{}
}

func newSuiteRegistrationService(suiteAction bindings.SuiteAction, suiteRegistrationChan chan interface{}) *suiteRegistrationService {
	return &suiteRegistrationService{
		mutex: &sync.Mutex{},
		suiteRegistered: false,
		suiteAction: suiteAction,
		suiteRegistrationChan: suiteRegistrationChan,
	}
}

func (service *suiteRegistrationService) RegisterSuite(_ context.Context, _ *emptypb.Empty) (*bindings.SuiteRegistrationResponse, error) {
	service.mutex.Lock()
	defer service.mutex.Unlock()

	if service.suiteRegistered {
		return nil, stacktrace.NewError("Suite has already been registered")
	}
	service.suiteRegistered = true
	service.suiteRegistrationChan <- "Suite registered"
	response := &bindings.SuiteRegistrationResponse{SuiteAction: service.suiteAction}
	return response, nil
}