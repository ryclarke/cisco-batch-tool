package git

import (
	"github.com/spf13/cobra"
)

// Cmd configures the root git command along with all subcommands and flags
func Cmd() *cobra.Command {
	rootCmd := &cobra.Command{
		Use:   "git [cmd] <repository> ...",
		Short: "Manage git branches and commits",
		Args:  cobra.MinimumNArgs(1),
	}

	defaultCmd := addStatusCmd()

	rootCmd.AddCommand(
		defaultCmd,
		addBranchCmd(),
		addCommitCmd(),
		addDiffCmd(),
		addUpdateCmd(),
	)
	rootCmd.Run = defaultCmd.Run

	return rootCmd
}
