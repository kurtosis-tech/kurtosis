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

	// these color constants should be exactly the same length, otherwise it will mess with the tabwriter padding
	yellowColorStr = "\033[33m"
	resetColorStr  = "\033[00m"
)

type kurtosisTabWriter struct {
	underlying                *tabwriter.Writer
	shouldAlternateLineColors bool
	shouldColorLine           bool
}

func newKurtosisTabWriter(shouldAlternateLineColors bool) *kurtosisTabWriter {
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
		shouldAlternateLineColors: shouldAlternateLineColors,
	}
}

func (writer *kurtosisTabWriter) writeElems(elems ...string) {
	if writer.shouldAlternateLineColors {
		var leaderChar string
		if writer.shouldColorLine {
			leaderChar = yellowColorStr
		} else {
			leaderChar = resetColorStr
		}
		writer.shouldColorLine = !writer.shouldColorLine

		fmt.Fprintln(
			writer.underlying,
			fmt.Sprint(leaderChar, strings.Join(elems, tabWriterElemJoinChar), resetColorStr),
		)
	} else {
		fmt.Fprintln(
			writer.underlying,
			strings.Join(elems, tabWriterElemJoinChar),
		)
	}
}

func (writer *kurtosisTabWriter) flush() {
	writer.underlying.Flush()
}
