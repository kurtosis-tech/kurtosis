package event

import (
	"crypto/sha256"
	"fmt"
	"strings"
)

const (
	// We are following these naming conventions for event's data
	// https://segment.com/docs/getting-started/04-full-install/#event-naming-best-practices
	enclaveIDPropertyKey           = "enclave_id"
	serviceIDPropertyKey           = "service_id"
	didUserAcceptSendingMetricsKey = "did_user_accept_sending_metrics"
	packageIdKey                   = "package_id"
	isRemotePackageKey             = "is_remote_package"
	isDryRunKey                    = "is_dry_run"
	isScriptKey                    = "is_script"
	numServicesKey                 = "num_services"
	isSuccessKey                   = "is_success"
	isSubnetworkingEnabledKey      = "is_subnetworking_enabled"
	userEmailAddressKey            = "user_email"
	analyticsStatusKey             = "analytics_status"

	// Categories
	installCategory = "install"
	enclaveCategory = "enclave"
	// the Kurtosis category is for commands at the root level of the cli
	// we went this way cause this is in pattern with other categories above
	// any further root level commands should use this category
	kurtosisCategory = "kurtosis"

	// Actions
	consentAction         = "consent"
	shareEmailAction      = "share-email"
	createAction          = "create"
	stopAction            = "stop"
	destroyAction         = "destroy"
	runAction             = "run"
	updateAction          = "update"
	serviceStartAction    = "service-start"
	serviceStopAction     = "service-stop"
	runFinishedAction     = "run-finished"
	analyticsToggleAction = "analytics-toggle"
)

// WARNING WARNING WARNING WARNING WARNING WARNING WARNING WARNING WARNING WARNING WARNING WARNING WARNING WARNING
// NO EVENTS SHOULD RETURN AN ERROR! Instead, each event should *always* return an event (even if the value is garbage)
// This is becasue if we return an error, the error will propagate which means that we don't even send the event
//  at all, which means that we'll silently drop data, which means we won't even realize that something is wrong!
// If we send the event with garbage data, it means that we at least get the chance to notice it in our product analytics
//  dashboards.
// WARNING WARNING WARNING WARNING WARNING WARNING WARNING WARNING WARNING WARNING WARNING WARNING WARNING WARNING

func NewShouldSendMetricsUserElectionEvent(didUserAcceptSendingMetrics bool) *Event {
	didUserAcceptSendingMetricsStr := fmt.Sprintf("%v", didUserAcceptSendingMetrics)
	properties := map[string]string{
		didUserAcceptSendingMetricsKey: didUserAcceptSendingMetricsStr,
	}
	event := newEvent(installCategory, consentAction, properties)
	return event
}

func NewUserSharesEmailAddress(userSharedEmailAddress string) *Event {
	properties := map[string]string{
		userEmailAddressKey: userSharedEmailAddress,
	}
	event := newEvent(installCategory, shareEmailAction, properties)
	return event
}

func NewCreateEnclaveEvent(enclaveId string, isSubnetworkingEnabled bool) *Event {
	hashedEnclaveId := hashString(strings.TrimSpace(enclaveId))
	isSubnetworkingEnabledStr := fmt.Sprintf("%v", isSubnetworkingEnabled)
	properties := map[string]string{
		enclaveIDPropertyKey:      hashedEnclaveId,
		isSubnetworkingEnabledKey: isSubnetworkingEnabledStr,
	}
	event := newEvent(enclaveCategory, createAction, properties)
	return event
}

func NewStopEnclaveEvent(enclaveId string) *Event {
	hashedEnclaveId := hashString(strings.TrimSpace(enclaveId))
	properties := map[string]string{
		enclaveIDPropertyKey: hashedEnclaveId,
	}
	event := newEvent(enclaveCategory, stopAction, properties)
	return event
}

func NewDestroyEnclaveEvent(enclaveId string) *Event {
	hashedEnclaveId := hashString(strings.TrimSpace(enclaveId))
	properties := map[string]string{
		enclaveIDPropertyKey: hashedEnclaveId,
	}
	event := newEvent(enclaveCategory, destroyAction, properties)
	return event
}

func NewKurtosisRunEvent(packageId string, isRemote bool, isDryRun bool, isScript bool) *Event {
	isRemotePackageStr := fmt.Sprintf("%v", isRemote)
	isDryRunStr := fmt.Sprintf("%v", isDryRun)
	isScriptStr := fmt.Sprintf("%v", isScript)

	properties := map[string]string{
		packageIdKey:       packageId,
		isRemotePackageKey: isRemotePackageStr,
		isDryRunKey:        isDryRunStr,
		isScriptKey:        isScriptStr,
	}

	event := newEvent(kurtosisCategory, runAction, properties)
	return event
}

func NewStartServiceEvent(enclaveId string, serviceId string) *Event {
	properties := map[string]string{
		enclaveIDPropertyKey: enclaveId,
		serviceIDPropertyKey: serviceId,
	}

	event := newEvent(kurtosisCategory, serviceStartAction, properties)
	return event
}

func NewStopServiceEvent(enclaveId string, serviceId string) *Event {
	properties := map[string]string{
		enclaveIDPropertyKey: enclaveId,
		serviceIDPropertyKey: serviceId,
	}

	event := newEvent(kurtosisCategory, serviceStopAction, properties)
	return event
}

func NewUpdateServiceEvent(enclaveId string, serviceId string) *Event {
	properties := map[string]string{
		enclaveIDPropertyKey: enclaveId,
		serviceIDPropertyKey: serviceId,
	}

	event := newEvent(kurtosisCategory, updateAction, properties)
	return event
}

func NewKurtosisRunFinishedEvent(packageId string, numServices int, isSuccess bool) *Event {
	numServicesStr := fmt.Sprintf("%v", numServices)
	isSuccessStr := fmt.Sprintf("%v", isSuccess)
	properties := map[string]string{
		packageIdKey:   packageId,
		numServicesKey: numServicesStr,
		isSuccessKey:   isSuccessStr,
	}

	event := newEvent(kurtosisCategory, runFinishedAction, properties)
	return event
}

func NewKurtosisAnalyticsToggleEvent(analayticsStatus bool) *Event {
	properties := map[string]string{
		analyticsStatusKey: fmt.Sprintf("%v", analayticsStatus),
	}
	event := newEvent(kurtosisCategory, analyticsToggleAction, properties)
	return event
}

// ================================================================================================
//
//	Private Helper Functions
//
// ================================================================================================
func hashString(value string) string {
	hash := sha256.New()

	hash.Write([]byte(value))

	hashedByteSlice := hash.Sum(nil)

	hexValue := fmt.Sprintf("%x", hashedByteSlice)

	return hexValue
}
