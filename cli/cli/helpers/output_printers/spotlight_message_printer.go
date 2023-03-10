package output_printers

import (
	"fmt"
	"github.com/sirupsen/logrus"
	"strings"
	"sync"
	"unicode/utf8"
)

const (
	frameChar   = "="
	borderChars = "||"

	borderWidth     = 2
	marginWidth     = 10
	amountOfBorders = 2
	amountOfMargins = 2

	spaceUnicodeChar = "\u0020"
)

var (
	// NOTE: This will be initialized exactly once (singleton pattern)
	currentSpotlightMessagePrinter *spotlightMessagePrinter
	once                           sync.Once
)

type spotlightMessagePrinter struct{}

// Prints a centered spotlight message
func GetSpotlightMessagePrinter() *spotlightMessagePrinter {
	// NOTE: We use a 'once' to initialize the spotlightMessagePrinter because we don't
	// want multiple spotlightMessagePrinter instances in existence
	once.Do(func() {
		currentSpotlightMessagePrinter = &spotlightMessagePrinter{}
	})
	return currentSpotlightMessagePrinter
}

func (printer *spotlightMessagePrinter) Print(message string) {
	frame, formattedMsgLine := printer.getFrameAndFormattedMsgLine(message)
	fmt.Println(frame)
	fmt.Println(formattedMsgLine)
	fmt.Println(frame)
}

func (printer *spotlightMessagePrinter) PrintWithLogger(message string) {
	frame, formattedMsgLine := printer.getFrameAndFormattedMsgLine(message)
	logrus.Infof(frame)
	logrus.Infof(formattedMsgLine)
	logrus.Infof(frame)
}

// ====================================================================================================
//
//	Private Helper Functions
//
// ====================================================================================================
func (printer *spotlightMessagePrinter) getFrameAndFormattedMsgLine(message string) (string, string) {
	columnWidth := printer.calculateColumnWidth(message)

	marginStr := strings.Repeat(spaceUnicodeChar, marginWidth)
	frameStr := strings.Repeat(frameChar, columnWidth)
	messageLineStr := fmt.Sprintf("%s%s%s%s%s", borderChars, marginStr, message, marginStr, borderChars)

	return frameStr, messageLineStr
}

func (printer *spotlightMessagePrinter) calculateColumnWidth(message string) int {

	bordersAndMarginWidth := borderWidth*amountOfBorders + marginWidth*amountOfMargins

	messageWidth := utf8.RuneCountInString(message)

	columnWidth := bordersAndMarginWidth + messageWidth

	return columnWidth
}
