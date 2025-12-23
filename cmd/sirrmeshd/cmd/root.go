package cmd

import (
	"github.com/spf13/cobra"
)

// NewRootCmd creates a new root command for the mail server application.
func NewRootCmd() *cobra.Command {
	rootCmd := &cobra.Command{
		Use:   "sirrmeshd",
		Short: "SirrMesh - Composable all-in-one email server",
		PersistentPreRunE: func(cmd *cobra.Command, _ []string) error {
			cmd.SetOut(cmd.OutOrStdout())
			cmd.SetErr(cmd.ErrOrStderr())
			return nil
		},
	}

	addMailCommands(rootCmd)

	return rootCmd
}
