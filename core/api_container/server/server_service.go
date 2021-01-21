/*
 * Copyright (c) 2021 - present Kurtosis Technologies LLC.
 * All Rights Reserved.
 */

package server

type ApiContainerServerService interface {

	// Hook to tell the service that a testsuite successfully registered with the API container server
	// Services shouldn't do much before receiving this event
	ReceiveSuiteRegistrationEvent()
}
