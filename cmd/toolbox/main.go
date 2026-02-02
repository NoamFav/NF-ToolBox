package main

import (
	"fmt"
	"os"

	"github.com/spf13/cobra"

	"github.com/nf-software/nf-toolbox/cmd/toolbox/commands"
)

var version = "0.1.0"

func main() {
	rootCmd := &cobra.Command{
		Use:   "nf-toolbox",
		Short: "NF-ToolBox - Manage your NF Software tools",
		Long: `NF-ToolBox is the manager for NF Software CLI tools.

Use it to login, install tools, manage devices, and keep everything updated.

Get started:
  nf-toolbox login
  nf-toolbox list
  nf-toolbox install <tool>`,
		Version: version,
	}

	// Add commands
	rootCmd.AddCommand(commands.LoginCmd)
	rootCmd.AddCommand(commands.LogoutCmd)
	rootCmd.AddCommand(commands.ListCmd)
	rootCmd.AddCommand(commands.InstallCmd)
	rootCmd.AddCommand(commands.UninstallCmd)
	rootCmd.AddCommand(commands.UpdateCmd)
	rootCmd.AddCommand(commands.DevicesCmd)
	rootCmd.AddCommand(commands.SyncCmd)
	rootCmd.AddCommand(commands.StatusCmd)

	if err := rootCmd.Execute(); err != nil {
		fmt.Fprintln(os.Stderr, err)
		os.Exit(1)
	}
}
