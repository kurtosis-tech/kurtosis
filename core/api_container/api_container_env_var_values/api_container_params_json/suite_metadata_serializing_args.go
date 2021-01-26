/*
 * Copyright (c) 2021 - present Kurtosis Technologies LLC.
 * All Rights Reserved.
 */

package api_container_params_json

import (
	"github.com/palantir/stacktrace"
	"reflect"
	"strings"
)

// Fields are public for JSON de/serialization
type SuiteMetadataSerializationArgs struct {
	// NOTE: There aren't any dynamic arguments for suite metadata serialization right now, but we leave
	//  this here so that we have a slot to add more args in the future if needed
}

// Even though the fields are public due to JSON de/serialization requirements, we still have this constructor so that
//  we get compile errors if there are missing fields
func NewSuiteMetadataSerializationArgs() (*SuiteMetadataSerializationArgs, error) {
	result := SuiteMetadataSerializationArgs{}
	if err := result.validate(); err != nil {
		return nil, stacktrace.Propagate(err, "An error occurred validating the suite metadata serialization args")
	}
	return &result, nil
}

func (args SuiteMetadataSerializationArgs) validate() error {
	reflectVal := reflect.ValueOf(args)
	reflectValType := reflectVal.Type()
	reflectValElem := reflectVal.Elem()
	for i := 0; i < reflectValType.NumField(); i++ {
		field := reflectValType.Field(i);
		jsonFieldName := field.Tag.Get(jsonFieldTag)
		strVal := reflectValElem.Field(i).String()
		if strings.TrimSpace(strVal) == "" {
			return stacktrace.NewError("JSON field '%s' is whitespace or empty string", jsonFieldName)
		}
	}
	return nil
}
