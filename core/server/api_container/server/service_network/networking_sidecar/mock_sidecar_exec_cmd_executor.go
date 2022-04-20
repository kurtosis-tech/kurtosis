/*
 * Copyright (c) 2021 - present Kurtosis Technologies Inc.
 * All Rights Reserved.
 */

package networking_sidecar

import (
	"context"
)

type mockSidecarExecCmdExecutor struct {
	commands     [][]string
	isBlocked bool
	unblockingChan chan interface{}
}

func newMockSidecarExecCmdExecutor() *mockSidecarExecCmdExecutor {
	return &mockSidecarExecCmdExecutor{
		commands:     [][]string{},
		isBlocked: false,
		unblockingChan: make(chan interface{}),
	}
}

func (m *mockSidecarExecCmdExecutor) setBlocked(isBlockedNew bool) {
	if m.isBlocked && !isBlockedNew {
		m.unblockingChan <- "unblocked"
	}
	m.isBlocked = isBlockedNew
}

func (m *mockSidecarExecCmdExecutor) exec(ctx context.Context, unwrappedCmd []string) error {
	if m.isBlocked {
		<- m.unblockingChan
	}
	m.commands = append(m.commands, unwrappedCmd)
	return nil
}
