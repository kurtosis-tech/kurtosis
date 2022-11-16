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
	currentFeaturedMessagePrinter *featuredMessagePrinter
	once sync.Once
)

type featuredMessagePrinter struct {}

// Prints a centered featured message
func GetFeaturedMessagePrinter() *featuredMessagePrinter {
	// NOTE: We use a 'once' to initialize the featuredMessagePrinter because we don't
	// want multiple featuredMessagePrinter instances in existence
	once.Do(func() {
		currentFeaturedMessagePrinter = &featuredMessagePrinter{}
	})
	return currentFeaturedMessagePrinter
}

func (printer *featuredMessagePrinter) Print(message string)  {
	columnWith := printer.calculateColumnWith(message)

	marginStr := strings.Repeat(spaceUnicodeChar, marginWith)
	frameStr := strings.Repeat(frameChar, columnWith)
	messageLineStr := fmt.Sprintf("%s%s%s%s%s", borderChars, marginStr, message, marginStr, borderChars)

	logrus.Infof(frameStr)
	logrus.Infof(messageLineStr)
	logrus.Infof(frameStr)

	return
}

func (printer *featuredMessagePrinter) calculateColumnWith(message string) int {

	bordersAndMarginWith := borderWith * amountOfBorders + marginWith * amountOfMargins

	messageWith := utf8.RuneCountInString(message)

	columnWith := bordersAndMarginWith + messageWith

	return columnWith
}