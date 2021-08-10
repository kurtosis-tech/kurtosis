/*
 * Copyright (c) 2021 - present Kurtosis Technologies Inc.
 * All Rights Reserved.
 */

package enclave_manager

import "github.com/sirupsen/logrus"

type EnclaveManagerTask struct {
	enclaveId string
	outputLog *logrus.Logger
	core      EnclaveManagerTaskCore
}

// TODO constructor

// TODO getters
func (task *EnclaveManagerTask) GetEnclaveId() string {
	return task.enclaveId
}

func (task *EnclaveManagerTask) GetOutputLog() *logrus.Logger {
	return task.outputLog
}

func (task *EnclaveManagerTask) GetCore() EnclaveManagerTaskCore {
	return task.core
}

