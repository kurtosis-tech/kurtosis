/*
 * Copyright (c) 2021 - present Kurtosis Technologies LLC.
 * All Rights Reserved.
 */

package bulk_command_execution_engine

import (
	"github.com/kurtosis-tech/kurtosis/api_container/server/service_network"
	"regexp"
	"strings"
)

const (
	serviceIdIpReplacementPrefix = "<<<"
	serviceIdIpReplacementSuffix = ">>>"

	// The *? makes this non-greedy
	serviceIdIpReplacementRegex = serviceIdIpReplacementPrefix + ".*?" + serviceIdIpReplacementSuffix

	considerEntireStringOffset = -1
)

/**
This struct replaces instances of <<<service_id>>> in a string with the IP address of the referenced service ID
 */
type serviceIpReplacer struct {
	serviceNetwork *service_network.ServiceNetwork
	matchPattern *regexp.Regexp
}

func newServiceIpReplacer(serviceNetwork *service_network.ServiceNetwork) *serviceIpReplacer {
	return &serviceIpReplacer{
		serviceNetwork: serviceNetwork,
		matchPattern: regexp.MustCompile(serviceIdIpReplacementRegex),
	}
}

// Returns a copy of the string with all referenced instances of service IDs replaced with their IP address
func (replacer serviceIpReplacer) Replace(str string) (string, error) {
	matches := replacer.matchPattern.FindAllString(str, considerEntireStringOffset)
	if matches == nil {
		return str, nil
	}
	ipReplacementRegexes := map[string]string{} // Maps regex replacement pattern -> service IP
	for _, match := range matches {
		serviceId := strings.TrimSuffix(strings.TrimPrefix(match, serviceIdIpReplacementPrefix), serviceIdIpReplacementSuffix)
		replacer.serviceNetwork.Get
	}
}

