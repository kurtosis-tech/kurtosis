/*
 * Copyright (c) 2021 - present Kurtosis Technologies LLC.
 * All Rights Reserved.
 */

package service_ip_replacer

import (
	"fmt"
	"github.com/kurtosis-tech/kurtosis/api_container/server/service_network"
	"github.com/kurtosis-tech/kurtosis/api_container/server/service_network/service_network_types"
	"github.com/stretchr/testify/assert"
	"net"
	"testing"
)

const (
	testPrefix = "<<<"
	testSuffix = ">>>"
)

func TestValidStrReplacing(t *testing.T) {
	serviceId := service_network_types.ServiceID("serviceId")
	var serviceIp net.IP = []byte{1, 2, 3, 4}
	serviceIps := map[service_network_types.ServiceID]net.IP{
		serviceId: serviceIp,
	}

	mockNetwork := service_network.NewMockServiceNetwork(serviceIps)
	ipReplacer, err := NewServiceIPReplacer(testPrefix, testSuffix, mockNetwork)
	assert.NoError(t, err, "An unexpected error occurred creating the IP replacer")

	strToReplace := testPrefix + string(serviceId) + testSuffix
	input := genTestString(strToReplace)
	expected := genTestString(serviceIp.String())
	output, err := ipReplacer.ReplaceStr(input)
	assert.NoError(t, err, "An error occurred during string replacement")
	assert.Equal(t, expected, output)
}

func TestErrorOnNonexistentServiceId(t *testing.T) {
	mockNetwork := service_network.NewMockServiceNetwork(map[service_network_types.ServiceID]net.IP{})
	ipReplacer, err := NewServiceIPReplacer(testPrefix, testSuffix, mockNetwork)
	assert.NoError(t, err, "An unexpected error occurred creating the IP replacer")

	input := fmt.Sprintf("Some string %vnonexistent_service_id%v", testPrefix, testSuffix)
	_, err = ipReplacer.ReplaceStr(input)
	assert.Error(t, err, "No error was encountered when one was expected")
}

func TestStrSliceReplacing(t *testing.T) {
	serviceId1 := service_network_types.ServiceID("serviceId1")
	var serviceIp1 net.IP = []byte{1, 2, 3, 4}
	serviceId2 := service_network_types.ServiceID("serviceId2")
	var serviceIp2 net.IP = []byte{5, 6, 7, 8}
	serviceIps := map[service_network_types.ServiceID]net.IP{
		serviceId1: serviceIp1,
		serviceId2: serviceIp2,
	}

	mockNetwork := service_network.NewMockServiceNetwork(serviceIps)
	ipReplacer, err := NewServiceIPReplacer(testPrefix, testSuffix, mockNetwork)
	assert.NoError(t, err, "An unexpected error occurred creating the IP replacer")

	input := []string{
		fmt.Sprintf("%v%v%v", testPrefix, serviceId1, testSuffix),
		fmt.Sprintf("%v%v%v", testPrefix, serviceId2, testSuffix),
	}
	expected := []string{
		serviceIp1.String(),
		serviceIp2.String(),
	}

	output, err := ipReplacer.ReplaceStrSlice(input)
	assert.NoError(t, err, "An error occurred during string slice replacement")
	assert.Equal(t, expected, output)
}

func TestMapValReplacing(t *testing.T) {
	serviceId1 := service_network_types.ServiceID("serviceId1")
	var serviceIp1 net.IP = []byte{1, 2, 3, 4}
	serviceId2 := service_network_types.ServiceID("serviceId2")
	var serviceIp2 net.IP = []byte{5, 6, 7, 8}
	serviceIps := map[service_network_types.ServiceID]net.IP{
		serviceId1: serviceIp1,
		serviceId2: serviceIp2,
	}

	mockNetwork := service_network.NewMockServiceNetwork(serviceIps)
	ipReplacer, err := NewServiceIPReplacer(testPrefix, testSuffix, mockNetwork)
	assert.NoError(t, err, "An unexpected error occurred creating the IP replacer")

	service1ReplacementPattern := fmt.Sprintf("%v%v%v", testPrefix, serviceId1, testSuffix)
	input := map[string]string{
		service1ReplacementPattern: service1ReplacementPattern,
		"key2": fmt.Sprintf("%v%v%v", testPrefix, serviceId2, testSuffix),
		"key3": "no replacement",
	}
	expected := map[string]string{
		service1ReplacementPattern: serviceIp1.String(), // Key should NOT be replaced
		"key2": serviceIp2.String(),
		"key3": "no replacement",
	}

	output, err := ipReplacer.ReplaceMapValues(input)
	assert.NoError(t, err, "An error occurred during string map replacement")
	assert.Equal(t, expected, output)
}

func genTestString(insertStr string) string {
	return fmt.Sprintf("This is one service ID, %v, and this is the same one repeated, %v", insertStr, insertStr)
}