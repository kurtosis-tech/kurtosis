//go:build !windows

package lsp

import (
	starlark_lsp_cli "github.com/kurtosis-tech/vscode-kurtosis/starlark-lsp/pkg/cli"
	"github.com/spf13/cobra"
)

func NewLspCommand() *cobra.Command {
	kurtosisPlugins := getKurtosisBuiltIn()
	rootCmd := starlark_lsp_cli.NewRootCmd("lsp", kurtosisPlugins)
	rootCmd.Use = "lsp"
	rootCmd.Hidden = true
	return rootCmd.Command
}
