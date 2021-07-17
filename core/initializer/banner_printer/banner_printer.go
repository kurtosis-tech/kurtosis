/*
 * Copyright (c) 2020 - present Kurtosis Technologies LLC.
 * All Rights Reserved.
 */

package banner_printer

import (
	"fmt"
	"github.com/sirupsen/logrus"
	"strings"
)

const (
	bannerWidth = 100
	bannerChar = "="
	sectionChar = "-"
)

func PrintBanner(log *logrus.Logger, contents string, isError bool) {
	bannerString := strings.Repeat(bannerChar, bannerWidth)
	numPaddingSpaces := (bannerWidth - len(contents)) / 2
	contentString := strings.Repeat(" ", numPaddingSpaces) + contents
	if !isError {
		log.Info("")
		log.Info(bannerString)
		log.Info(contentString)
		log.Info(bannerString)
	} else {
		log.Error("")
		log.Error(bannerString)
		log.Error(contentString)
		log.Error(bannerString)
	}
}

func PrintSection(log *logrus.Logger, contents string, isError bool) {
	contentsPlusBuffer := fmt.Sprintf(" %v ", contents)
	numLeadingDashes := (bannerWidth - len(contentsPlusBuffer)) / 2
	numTrailingDashes := bannerWidth - (numLeadingDashes + len(contentsPlusBuffer))
	logStr := strings.Repeat(sectionChar, numLeadingDashes) + contentsPlusBuffer + strings.Repeat(sectionChar, numTrailingDashes)
	if isError {
		log.Error(logStr)
	} else {
		log.Info(logStr)
	}
}
