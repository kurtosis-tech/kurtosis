package output_printers

import "github.com/fatih/color"

type KeyValuePrinter struct {
	tabWriter   *kurtosisTabWriter
	forPrinting [][]string
}

func NewKeyValuePrinter() *KeyValuePrinter {
	return &KeyValuePrinter{
		tabWriter:   newKurtosisTabWriter(),
		forPrinting: [][]string{},
	}
}

func (printer *KeyValuePrinter) AddPair(key string, value string) {
	printer.forPrinting = append(printer.forPrinting, []string{makeInputStrBold(key) + ":", value})
}

func (printer *KeyValuePrinter) Print() {
	for _, keyValuePairForPrinting := range printer.forPrinting {
		printer.tabWriter.writeElems(keyValuePairForPrinting...)
	}
	printer.tabWriter.flush()
}

func makeInputStrBold(inputStr string) string {
	return color.New(color.Bold).SprintFunc()(inputStr)
}
