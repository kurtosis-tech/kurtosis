package output_printers

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
	printer.forPrinting = append(printer.forPrinting, []string{key + ":", value})
}

func (printer *KeyValuePrinter) Print() {
	for _, keyValuePairForPrinting := range printer.forPrinting {
		printer.tabWriter.writeElems(keyValuePairForPrinting...)
	}
	printer.tabWriter.flush()
}
