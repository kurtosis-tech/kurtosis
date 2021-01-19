/*
 * Copyright (c) 2021 - present Kurtosis Technologies LLC.
 * All Rights Reserved.
 */

package execution_path

// Represents a lambda encapsulating a runtime path that the API server can be running
type ExecutionPath interface {
	Execute() error
}
