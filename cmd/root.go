package cmd

import (
	"fmt"
	"os"

	"github.com/spf13/cobra"
	"github.com/spf13/viper"

	"github.com/ryclarke/cisco-batch-tool/cmd/git"
	"github.com/ryclarke/cisco-batch-tool/cmd/pr"
	"github.com/ryclarke/cisco-batch-tool/config"
)

// RootCmd configures the top-level root command along with all subcommands and flags
func RootCmd() *cobra.Command {
	rootCmd := &cobra.Command{
		Use:   "batch-tool",
		Short: "Batch tool for working across multiple git repositories",
		Long:  `Batch tool for working across multiple git repositories

This tool provides a collection of utility functions that facilitate work across
multiple git repositories, including branch management and pull request creation.`,
		PersistentPreRun: func(cmd *cobra.Command, _ []string) {
			// Allow the `--no-sort` flag to override sorting configuration
			if noSort, _ := cmd.Flags().GetBool("no-sort"); noSort {
				viper.Set(config.SortRepos, false)
			}
		},
	}

	// Add all subcommands to the root
	rootCmd.AddCommand(
		git.Cmd(),
		pr.Cmd(),
		addMakeCmd(),
		addShellCmd(),
	)

	rootCmd.PersistentFlags().StringVar(&config.CfgFile, "config", "", "config file (default is .config.yaml)")

	rootCmd.PersistentFlags().Bool("sync", false, "execute commands synchronously")
	viper.BindPFlag(config.UseSync, rootCmd.PersistentFlags().Lookup("sync"))

	rootCmd.PersistentFlags().Bool("sort", true, "sort the provided repositories")
	viper.BindPFlag(config.SortRepos, rootCmd.PersistentFlags().Lookup("sort"))

	// --no-sort is excluded from usage and help output, and is an alternative to --sort=false
	rootCmd.PersistentFlags().Bool("no-sort", false, "")
	rootCmd.PersistentFlags().MarkHidden("no-sort")

	return rootCmd
}

// Execute adds all child commands to the root command and sets flags appropriately.
// This is called by main.main(). It only needs to happen once to the RootCmd.
func Execute() {
	cobra.OnInitialize(config.Init)

	if err := RootCmd().Execute(); err != nil {
		fmt.Println(err)
		os.Exit(1)
	}
}
