package output_printers

import (
	"fmt"
	"github.com/dzobbe/PoTE-kurtosis/cli/cli/out"
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

type kurtosisTabWriter struct {
	underlying *tabwriter.Writer
}

func newKurtosisTabWriter() *kurtosisTabWriter {
	tabWriter := tabwriter.NewWriter(
		out.GetOut(),
		tabWriterMinwidth,
		tabWriterTabwidth,
		tabWriterPadding,
		tabWriterPadchar,
		tabWriterFlags,
	)
	return &kurtosisTabWriter{
		underlying: tabWriter,
	}
}

func (writer *kurtosisTabWriter) writeElems(elems ...string) {
	if _, err := fmt.Fprintln(writer.underlying, strings.Join(elems, tabWriterElemJoinChar)); err != nil {
		logrus.Errorf("Error printing table element to StdfOut. Error was: \n%v", err.Error())
	}
}

func (writer *kurtosisTabWriter) flush() {
	if err := writer.underlying.Flush(); err != nil {
		logrus.Errorf("Error flushing table to StdfOut. Error was: \n%v", err.Error())
	}
}
