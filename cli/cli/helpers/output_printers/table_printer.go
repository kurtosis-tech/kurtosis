package output_printers

import (
	"github.com/kurtosis-tech/stacktrace"
)

type TablePrinter struct {
	tabWriter *kurtosisTabWriter

	columnHeaders []string

	dataRows [][]string
}

// Prints columns of output, each with a header
func NewTablePrinter(columnHeaders ...string) *TablePrinter {
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
