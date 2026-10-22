package cli

import (
	"fmt"

	"github.com/spf13/cobra"

	"github.com/dasosis/hopscope/internal/version"
)

func newVersionCmd() *cobra.Command {
	return &cobra.Command{
		Use:   "version",
		Short: "Print version and build information",
		Args:  cobra.NoArgs,
		RunE: func(cmd *cobra.Command, _ []string) error {
			// Write to cmd.OutOrStdout() rather than os.Stdout so tests can
			// capture output with cmd.SetOut.
			_, err := fmt.Fprintln(cmd.OutOrStdout(), version.Get())
			return err
		},
	}
}
