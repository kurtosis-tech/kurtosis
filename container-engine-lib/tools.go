// +build tools

package main

// This file is to make sure that go.mod continues to record the tools we need in our 'go generate' commands
// See: https://marcofranssen.nl/manage-go-tools-via-go-modules
import (
    _ "github.com/dmarkham/enumer"
)
