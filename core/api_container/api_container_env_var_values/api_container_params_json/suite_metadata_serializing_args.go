/*
 * Copyright (c) 2021 - present Kurtosis Technologies LLC.
 * All Rights Reserved.
 */

package api_container_params_json

// Fields are public for JSON de/serialization
type SuiteMetadataSerializationArgs struct {
	// NOTE: There aren't any dynamic arguments for suite metadata serialization right now, but we leave
	//  this here so that we have a slot to add more args in the future if needed
}

// Even though the fields are public due to JSON de/serialization requirements, we still have this constructor so that
//  we get compile errors if there are missing fields
func NewSuiteMetadataSerializationArgs() *SuiteMetadataSerializationArgs {
	return &SuiteMetadataSerializationArgs{}
}
