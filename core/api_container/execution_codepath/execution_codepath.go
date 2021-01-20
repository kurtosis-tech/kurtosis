/*
 * Copyright (c) 2021 - present Kurtosis Technologies LLC.
 * All Rights Reserved.
 */

package execution_codepath

// Represents a lambda encapsulating a runtime path that the API server can be running
type ExecutionCodepath interface {
	// Returns: exit code that the API container should exit with, and an error if one exists
	Execute() (int, error)
}
