/*
 * Copyright (c) 2021 - present Kurtosis Technologies Inc.
 * All Rights Reserved.
 */

package test_mocks

import (
	session_cache2 "github.com/kurtosis-tech/kurtosis/cli/commands/test/testing_machinery/auth/session_cache"
	"github.com/palantir/stacktrace"
)

type MockSessionCache struct {
	throwErrorOnSave bool
	throwErrorOnLoad bool
	sessionToReturn *session_cache2.Session
	sessionsSavedInOrder []session_cache2.Session
}

func NewMockSessionCache(throwErrorOnSave bool, throwErrorOnLoad bool, sessionToReturn *session_cache2.Session) *MockSessionCache {
	return &MockSessionCache{
		throwErrorOnSave: throwErrorOnSave,
		throwErrorOnLoad: throwErrorOnLoad,
		sessionToReturn: sessionToReturn,
	}
}

func (m *MockSessionCache) SaveSession(session session_cache2.Session) error {
	if m.throwErrorOnSave {
		return stacktrace.NewError("Test error thrown on session save, as requested")
	}
	m.sessionsSavedInOrder = append(m.sessionsSavedInOrder, session)
	return nil
}

func (m *MockSessionCache) LoadSession() (tokenResponse *session_cache2.Session, err error) {
	if m.throwErrorOnLoad {
		return nil, stacktrace.NewError("Test error thrown on session load, as requested")
	}
	return m.sessionToReturn, nil
}

func (m *MockSessionCache) GetSavedSessions() []session_cache2.Session {
	return m.sessionsSavedInOrder
}





