// Package cli defines hopscope's command-line interface.
//
// Commands are built by constructor functions rather than package-level
// variables so every call returns a fresh, independent command tree. Tests can
// then execute commands in isolation without flag state leaking between them.
package cli

import (
	"github.com/spf13/cobra"
)

// NewRootCmd builds the top-level hopscope command and its subcommands.
func NewRootCmd() *cobra.Command {
	root := &cobra.Command{
		Use:   "hopscope",
		Short: "Explain why your internet feels slow",
		Long: `hopscope walks the network path hop by hop and layer by layer
(device -> router -> ISP -> DNS -> destination), finds the faulty layer,
and explains it in plain English with evidence and a suggested fix.`,
		// Errors are printed once by main; usage text on every runtime error
		// (e.g. a timeout) would bury the actual message.
		SilenceUsage:  true,
		SilenceErrors: true,
	}

	root.AddCommand(newVersionCmd())
	return root
}
