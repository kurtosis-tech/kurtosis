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

	Yellow = "\033[33m"
	White  = "\033[37m"
	Reset  = "\033[0m"
)

type kurtosisTabWriter struct {
	underlying *tabwriter.Writer
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
		underlying: tabWriter,
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

type kurtosisTabWriterWithColor struct {
	underlying *tabwriter.Writer
	useColor   bool
}

func newKurtosisTabWriterWithColor() *kurtosisTabWriterWithColor {
	tabWriter := tabwriter.NewWriter(
		logrus.StandardLogger().Out,
		tabWriterMinwidth,
		tabWriterTabwidth,
		tabWriterPadding,
		tabWriterPadchar,
		tabWriterFlags,
	)
	return &kurtosisTabWriterWithColor{
		underlying: tabWriter,
	}
}

func (writer *kurtosisTabWriterWithColor) writeElems(elems ...string) {
	var line string
	if writer.useColor {
		line = fmt.Sprint(Yellow, strings.Join(elems, tabWriterElemJoinChar), Reset)
	} else {
		line = fmt.Sprint(White, strings.Join(elems, tabWriterElemJoinChar), Reset)
	}
	fmt.Fprintln(
		writer.underlying,
		line,
	)
	writer.useColor = !writer.useColor
}

func (writer *kurtosisTabWriterWithColor) flush() {
	writer.underlying.Flush()
}
