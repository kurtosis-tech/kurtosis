/*
 * Copyright (c) 2021 - present Kurtosis Technologies Inc.
 * All Rights Reserved.
 */

package user_service

import (
	"github.com/spf13/cobra"
)

var UserServiceCmd = &cobra.Command{
	Use:   "user-service",
	Short: "Manage user services",
	RunE:  nil,
}

func init() {
	//UserServiceCmd.AddCommand(xxx)
}

