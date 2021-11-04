/*
 * Copyright (c) 2021 - present Kurtosis Technologies Inc.
 * All Rights Reserved.
 */

package module_store_types

type ModuleID string

// This is the module globally-unique identifier necessary to identify a module container and
//  the module's folder in the enclave data dir when two modules with the same ID are loaded
//  in the same execution period. For example, if we were to use ModuleID as the container name
//  and a module with ID "MyModule" is loaded, removed, and then a new module with the same ID is
//  loaded then the names of the containers would collide
// Using the ModuleGUID avoid this collision
type ModuleGUID string
