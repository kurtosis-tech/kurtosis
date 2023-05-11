//go:build windows

package lsp

import (
	"github.com/sirupsen/logrus"
	"github.com/spf13/cobra"
)

func NewLspCommand() *cobra.Command {
	return &cobra.Command{
		Run: func(cmd *cobra.Command, args []string) {
			logrus.Errorf("Starlark LSP is not supported on Windows")
		},
	}
}
