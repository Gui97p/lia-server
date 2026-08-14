package main

import "github.com/spf13/cobra"

var rootCmd = &cobra.Command{
	Use:   "lia-admin",
	Short: "Lia's server configuration",
}

func init() {
	rootCmd.AddCommand(usersCmd)
	rootCmd.AddCommand(tokensCmd)
}
