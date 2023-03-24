package output_printers

import (
	"github.com/fatih/color"
	"github.com/kurtosis-tech/stacktrace"
	"strings"
)

type TablePrinter struct {
	tabWriter *kurtosisTabWriter

	columnHeaders []string

	dataRows [][]string
}

var (
	noColorForPadding = color.New(color.Reset).SprintfFunc()
)

// Prints columns of output, each with a header
func NewTablePrinter(columnHeaders ...string) *TablePrinter {
	for index := range columnHeaders {
		columnHeaders[index] = makeInputStrBold(columnHeaders[index])
	}
	return &TablePrinter{
		tabWriter:     newKurtosisTabWriter(),
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

	// This is a hack; tabWriter doesn't support colored strings when it comes to formatting
	// We format uncolored strings with Reset so that they have similar padding as colored strings
	for index := range data {
		if strings.HasPrefix(data[index], "\x1b") {
			continue
		}
		data[index] = noColorForPadding(data[index])
	}

	printer.dataRows = append(printer.dataRows, data)
	return nil
}

func (printer *TablePrinter) Print() {
	printer.tabWriter.writeElems(printer.columnHeaders...)
	for _, dataRow := range printer.dataRows {
		printer.tabWriter.writeElems(dataRow...)
	}
	printer.tabWriter.flush()

}
