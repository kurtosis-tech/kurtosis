/*
 * Copyright (c) 2021 - present Kurtosis Technologies Inc.
 * All Rights Reserved.
 */

package service_network_types

import (
	"github.com/kurtosis-tech/kurtosis/commons/enclave_data_volume"
	"net"
	"strconv"
	"time"
)

// Information that gets created with a service's registration
type ServiceRegistrationInfo interface {
	ServiceGUID() ServiceGUID
	Alias() string
	IpAddr() net.IP
	ServiceDirectory() *enclave_data_volume.ServiceDirectory
	RegistrationTimestamp() int64
	SetServiceDirectory(serviceDirectory *enclave_data_volume.ServiceDirectory)
}

type serviceRegistrationInfoImpl struct {
	serviceID ServiceID
	ipAddr                net.IP
	serviceDirectory      *enclave_data_volume.ServiceDirectory
	registrationTimestamp int64
}

func NewServiceRegistrationInfo(serviceID ServiceID, ipAddr net.IP) ServiceRegistrationInfo {
	now := time.Now()
	nowSec := now.Unix()
	return &serviceRegistrationInfoImpl{serviceID: serviceID, ipAddr: ipAddr, registrationTimestamp: nowSec}
}

func (s serviceRegistrationInfoImpl) ServiceGUID() ServiceGUID {
	registrationTimestampStr := strconv.FormatInt(s.registrationTimestamp, 10)
	return ServiceGUID(string(s.serviceID) + "_" + registrationTimestampStr)
}

func (s serviceRegistrationInfoImpl) Alias() string {
	return string(s.serviceID)
}

func (s serviceRegistrationInfoImpl) IpAddr() net.IP {
	return s.ipAddr
}

func (s serviceRegistrationInfoImpl) ServiceDirectory() *enclave_data_volume.ServiceDirectory {
	return s.serviceDirectory
}

func (s serviceRegistrationInfoImpl) RegistrationTimestamp() int64 {
	return s.registrationTimestamp
}

func (s *serviceRegistrationInfoImpl) SetServiceDirectory(serviceDirectory *enclave_data_volume.ServiceDirectory) {
	s.serviceDirectory = serviceDirectory
}
