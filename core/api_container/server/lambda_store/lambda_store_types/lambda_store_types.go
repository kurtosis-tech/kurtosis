/*
 * Copyright (c) 2021 - present Kurtosis Technologies Inc.
 * All Rights Reserved.
 */

package lambda_store_types

type LambdaID string

// This is the Lambda Global Unique Identifier necessary to identify the lambda's container and
//the lambda's folder in the enclave data volume when two lambdas with the same ID are loaded
//in the same execution period. For instance if a lambda with ID "MyLambda" is loaded with
//Kurt Interactive and stopped and then a new lambda with the same ID is loaded the names
//of the containers would collide if they have the LambdaID as the name, but using
//the LambdaGUID avoid this collision
type LambdaGUID string
