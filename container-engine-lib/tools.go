// +build tools

package main

// This file is a convention to make sure that go.mod continues to have the module we need in our 'go generate' commands
import (
    _ "github.com/dmarkham/enumer"
)
