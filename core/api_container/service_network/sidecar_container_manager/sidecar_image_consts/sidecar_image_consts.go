/*
 * Copyright (c) 2021 - present Kurtosis Technologies LLC.
 * All Rights Reserved.
 */

package sidecar_image_consts

import "strings"

const (
	ImageName = "kurtosistech/iproute2"
)

var RunForeverCmd = []string{
	// We sleep forever because all the commands this container will run will be executed
	//  via Docker exec
	"sleep","infinity",
}

// Embeds the given command in a call to whichever shell is native to the image, so that a command with things
//  like '&&' will get executed as expected
var ShWrappingCmd = func(unwrappedCmd []string) []string {
	return []string{
		"sh",
		"-c",
		strings.Join(unwrappedCmd, " "),
	}
}
