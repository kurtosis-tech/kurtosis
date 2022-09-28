package output_printers

import (
	"fmt"
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
	underlying                *tabwriter.Writer
}

func newKurtosisTabWriter() *kurtosisTabWriter {
	tabWriter := tabwriter.NewWriter(
		logrus.StandardLogger().Out,
		tabWriterMinwidth,
		tabWriterTabwidth,
		tabWriterPadding,
		tabWriterPadchar,
		tabWriterFlags,
	)
	return &kurtosisTabWriter{
		underlying:                tabWriter,
	}
}

func (writer *kurtosisTabWriter) writeElems(elems ...string) {
	fmt.Fprintln(
		writer.underlying,
		strings.Join(elems, tabWriterElemJoinChar),
	)
}

func (writer *kurtosisTabWriter) flush() {
	writer.underlying.Flush()
}
