/*
 * Copyright (c) 2021 - present Kurtosis Technologies LLC.
 * All Rights Reserved.
 */

package execution_impl

type ExampleTestsuiteArgs struct {
	ApiServiceImage	string 			`json:"apiServiceImage"`
	DatastoreServiceImage string	`json:"datastoreServiceImage"`

	// Indicates that this testsuite is being run as part of CI testing in Kurtosis Core
	IsKurtosisCoreDevMode bool		`json:"isKurtosisCoreDevMode"`
}
