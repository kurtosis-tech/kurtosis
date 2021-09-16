/*
 * Copyright (c) 2021 - present Kurtosis Technologies Inc.
 * All Rights Reserved.
 */

package service_network_types

type ServiceID string

// This is the Service Global Unique Identifier necessary to identify the service's container and
//the service's folder in the enclave data volume when two services with the same ID are loaded
//in the same execution period. For instance if a service with ID "MyService" is loaded with
//Kurt Interactive and stopped, and then a new service with the same ID is loaded the names of
//the containers would collide if they have the ServiceID as the name, but using the ServiceGUID
//avoid this collision
type ServiceGUID string

type PartitionID string

