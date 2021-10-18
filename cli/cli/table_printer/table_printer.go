package table_printer

import (
	"fmt"
	"github.com/palantir/stacktrace"
	"github.com/sirupsen/logrus"
	"strings"
	"text/tabwriter"
)

const (
	tabWriterMinwidth = 0
	tabWriterTabwidth = 0
	tabWriterPadding  = 3
	tabWriterPadchar  = ' '
	tabWriterFlags    = 0

	tabWriterElemJoinChar = "\t"
)

type TablePrinter struct {
	tabWriter *tabwriter.Writer

	columnHeaders []string

	dataRows [][]string
}

func NewTablePrinter(columnHeaders ...string) *TablePrinter {
	tabWriter := tabwriter.NewWriter(
		logrus.StandardLogger().Out,
		tabWriterMinwidth,
		tabWriterTabwidth,
		tabWriterPadding,
		tabWriterPadchar,
		tabWriterFlags,
	)

	return &TablePrinter{
		tabWriter:     tabWriter,
		columnHeaders: columnHeaders,
		dataRows:      [][]string{},
	}
}

func (printer *TablePrinter) AddRow(data ...string) error {
	numDataElems := len(data)
	numColHeaders := len(printer.columnHeaders)
	if numDataElems != numColHeaders {
		return stacktrace.NewError(
			"Data row '%+v' has %v values but the table (as defined by the header) has %v",
			data,
			numDataElems,
			numColHeaders,
		)
	}

	printer.dataRows = append(printer.dataRows, data)
	return nil
}

func (printer *TablePrinter) Print() {
	fmt.Fprintln(
		printer.tabWriter,
		strings.Join(printer.columnHeaders, tabWriterElemJoinChar),
	)
	for _, dataRow := range printer.dataRows {
		fmt.Fprintln(
			printer.tabWriter,
			strings.Join(dataRow, tabWriterElemJoinChar),
		)
	}
	printer.tabWriter.Flush()

}
