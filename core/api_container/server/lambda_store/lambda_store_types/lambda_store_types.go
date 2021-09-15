/*
 * Copyright (c) 2021 - present Kurtosis Technologies Inc.
 * All Rights Reserved.
 */

package lambda_store_types

import (
	"github.com/docker/go-connections/nat"
	"github.com/kurtosis-tech/kurtosis-lambda-api-lib/golang/kurtosis_lambda_rpc_api_bindings"
	"net"
	"strconv"
	"time"
)

type LambdaID string

type LambdaGUID string

type LambdaInfo interface {
	LambdaGUID() LambdaGUID
	ContainerId() string
	IpAddr() net.IP
	Client() kurtosis_lambda_rpc_api_bindings.LambdaServiceClient
	RegistrationTimestamp() int64
	SetContainerId(containerId string)
	SetIpAddr(ipAddr net.IP)
	SetClient(client kurtosis_lambda_rpc_api_bindings.LambdaServiceClient)
	SetHostPortBinding(hostPortBinding *nat.PortBinding)
}

type lambdaInfoImpl struct {
	lambdaID              LambdaID
	containerId           string
	ipAddr                net.IP
	client                kurtosis_lambda_rpc_api_bindings.LambdaServiceClient
	// NOTE: We don't use module host port bindings for now; we could expose them in the future if it's useful
	hostPortBinding *nat.PortBinding
	registrationTimestamp int64
}

func NewLambdaInfo(
	lambdaID LambdaID,
) LambdaInfo {
	now := time.Now()
	nowSec := now.Unix()
	return &lambdaInfoImpl{lambdaID: lambdaID, registrationTimestamp: nowSec}
}

func (l lambdaInfoImpl) LambdaGUID() LambdaGUID {
	registrationTimestampStr := strconv.FormatInt(l.registrationTimestamp, 10)
	return LambdaGUID(string(l.lambdaID) + "_" + registrationTimestampStr)
}

func (l lambdaInfoImpl) ContainerId() string {
	return l.containerId
}

func (l lambdaInfoImpl) IpAddr() net.IP {
	return l.ipAddr
}

func (l lambdaInfoImpl) Client() kurtosis_lambda_rpc_api_bindings.LambdaServiceClient {
	return l.client
}

func (l lambdaInfoImpl) RegistrationTimestamp() int64 {
	return l.registrationTimestamp
}

func (l *lambdaInfoImpl) SetContainerId(containerId string) {
	l.containerId = containerId
}

func (l *lambdaInfoImpl) SetIpAddr(ipAddr net.IP) {
	l.ipAddr = ipAddr
}

func (l *lambdaInfoImpl) SetClient(client kurtosis_lambda_rpc_api_bindings.LambdaServiceClient) {
	l.client = client
}

func (l *lambdaInfoImpl) SetHostPortBinding(hostPortBinding *nat.PortBinding) {
	l.hostPortBinding = hostPortBinding
}
