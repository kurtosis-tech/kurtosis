package lsp

import (
	"github.com/kurtosis-tech/kurtosis/lsp"
	starlark_lsp_cli "github.com/kurtosis-tech/vscode-kurtosis/starlark-lsp/pkg/cli"
	"github.com/spf13/cobra"
)

func NewLspCommand() *cobra.Command {
	kurtosisPlugins := lsp.GetKurtosisBuiltIn()
	rootCmd := starlark_lsp_cli.NewRootCmd("lsp", kurtosisPlugins)
	rootCmd.Use = "lsp"
	rootCmd.Hidden = true
	return rootCmd.Command
}
