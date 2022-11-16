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

	borderWith = 2
	marginWith = 10
	amountOfBorders = 2
	amountOfMargins = 2

	spaceUnicodeChar = "\u0020"
)

var (
	// NOTE: This will be initialized exactly once (singleton pattern)
	currentSpotlightMessagePrinter *spotlightMessagePrinter
	once                           sync.Once
)

type spotlightMessagePrinter struct {}

// Prints a centered spotlight message
func GetSpotlightMessagePrinter() *spotlightMessagePrinter {
	// NOTE: We use a 'once' to initialize the spotlightMessagePrinter because we don't
	// want multiple spotlightMessagePrinter instances in existence
	once.Do(func() {
		currentSpotlightMessagePrinter = &spotlightMessagePrinter{}
	})
	return currentSpotlightMessagePrinter
}

func (printer *spotlightMessagePrinter) Print(message string)  {
	columnWith := printer.calculateColumnWith(message)

	marginStr := strings.Repeat(spaceUnicodeChar, marginWith)
	frameStr := strings.Repeat(frameChar, columnWith)
	messageLineStr := fmt.Sprintf("%s%s%s%s%s", borderChars, marginStr, message, marginStr, borderChars)

	logrus.Infof(frameStr)
	logrus.Infof(messageLineStr)
	logrus.Infof(frameStr)
}

func (printer *spotlightMessagePrinter) calculateColumnWith(message string) int {

	bordersAndMarginWith := borderWith * amountOfBorders + marginWith * amountOfMargins

	messageWith := utf8.RuneCountInString(message)

	columnWith := bordersAndMarginWith + messageWith

	return columnWith
}
