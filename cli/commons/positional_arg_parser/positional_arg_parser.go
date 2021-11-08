/*
 * Copyright (c) 2021 - present Kurtosis Technologies Inc.
 * All Rights Reserved.
 */

package positional_arg_parser

import (
	"github.com/kurtosis-tech/stacktrace"
	"strings"
)

// Parses the args into a map of positional_arg_name -> value
// The result map is guaranteed to:
//  1) have one key for every value in the positionalArgNames string
//  2) not contain any values that are emptystring or whitespace
// This means that users won't need to do a map "found" check, nor a "len(strings.TrimSpace(theArg)) == 0" check
func ParsePositionalArgsAndRejectEmptyStrings(positionalArgNames []string, args []string) (map[string]string, error) {
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
		if len(strings.TrimSpace(argValue)) == 0 {
			return nil, stacktrace.NewError("Positional argument '%v' cannot be empty or whitespace", arg)
		}
		result[arg] = argValue
	}
	return result, nil
}
