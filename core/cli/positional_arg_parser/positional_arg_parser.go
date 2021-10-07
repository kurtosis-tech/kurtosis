/*
 * Copyright (c) 2021 - present Kurtosis Technologies Inc.
 * All Rights Reserved.
 */

package positional_arg_parser

import (
	"github.com/palantir/stacktrace"
	"strings"
)

// Parses the args into a map of positional_arg_name -> value
// The result map is guaranteed to have one key for every value in the positionalArgNames string
func ParsePositionalArgs(positionalArgNames []string, args []string) (map[string]string, error) {
	if len(args) != len(positionalArgNames) {
		return nil, stacktrace.NewError(
			"Expected positional arguments '%v' but only got %v args",
			strings.Join(positionalArgNames, " "),
			len(args),
		)
	}

	result := map[string]string{}
	for idx, argValue := range args {
		arg := positionalArgNames[idx]
		result[arg] = argValue
	}
	return result, nil
}
