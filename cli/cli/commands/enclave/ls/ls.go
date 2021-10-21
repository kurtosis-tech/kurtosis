/*
 * Copyright (c) 2021 - present Kurtosis Technologies Inc.
 * All Rights Reserved.
 */

package ls

import (
	"context"
	"fmt"
	"github.com/kurtosis-tech/kurtosis-cli/cli/enclave_statuses"
	"github.com/kurtosis-tech/kurtosis-cli/cli/logrus_log_levels"
	"github.com/kurtosis-tech/kurtosis-cli/cli/output_printers"
	"github.com/kurtosis-tech/kurtosis-engine-api-lib/golang/lib/kurtosis_context"
	"github.com/palantir/stacktrace"
	"github.com/sirupsen/logrus"
	"github.com/spf13/cobra"
	"sort"
	"strings"
)

const (
	kurtosisLogLevelArg = "kurtosis-log-level"

	enclaveIdColumnHeader = "EnclaveID"
	enclaveStatusColumnHeader = "Status"

	shouldExamineStoppedContainersForListingEnclaves = true
)

var kurtosisLogLevelStr string

var defaultKurtosisLogLevel = logrus.InfoLevel.String()

var LsCmd = &cobra.Command{
	Use:   "ls",
	Short: "List Kurtosis enclaves",
	RunE:  run,
}

func init() {
	LsCmd.Flags().StringVarP(
		&kurtosisLogLevelStr,
		kurtosisLogLevelArg,
		"l",
		defaultKurtosisLogLevel,
		fmt.Sprintf(
			"The log level that Kurtosis itself should log at (%v)",
			strings.Join(logrus_log_levels.GetAcceptableLogLevelStrs(), "|"),
		),
	)
}

func run(cmd *cobra.Command, args []string) error {
	ctx := context.Background()

	kurtosisLogLevel, err := logrus.ParseLevel(kurtosisLogLevelStr)
	if err != nil {
		return stacktrace.Propagate(err, "An error occurred parsing Kurtosis loglevel string '%v' to a log level object", kurtosisLogLevelStr)
	}
	logrus.SetLevel(kurtosisLogLevel)

	kurtosisContext, err := kurtosis_context.NewKurtosisContextFromLocalEngine()
	if err != nil {
		return stacktrace.Propagate(err, "An error occurred creating a new Kurtosis Context")
	}

	enclaves, err := kurtosisContext.GetEnclaves(ctx)
	if err != nil {
		return stacktrace.Propagate(err, "An error occurred creating an enclave, make sure that you already started Kurtosis Engine Sever with `kurtosis engine start` command")
	}

	orderedEnclaveIds := []string{}
	enclaveStatuses := map[string]enclave_statuses.EnclaveStatus{}
	for enclaveId, enclaveContext := range enclaves {
		orderedEnclaveIds = append(orderedEnclaveIds, enclaveId)
		enclaveStatuses[enclaveId] = enclaveContext.GetStatus()
	}
	sort.Strings(orderedEnclaveIds)

	tablePrinter := output_printers.NewTablePrinter(enclaveIdColumnHeader, enclaveStatusColumnHeader)
	for _, enclaveId := range orderedEnclaveIds {
		enclaveStatus, found := enclaveStatuses[enclaveId]
		if !found {
			return stacktrace.NewError("We're about to print enclave '%v', but it doesn't have a status; this is a bug in Kurtosis!", enclaveId)
		}
		if err := tablePrinter.AddRow(enclaveId, string(enclaveStatus)); err != nil {
			return stacktrace.NewError("An error occurred adding row for enclave '%v' to the table printer", enclaveId)
		}
	}
	tablePrinter.Print()

	return nil
}
