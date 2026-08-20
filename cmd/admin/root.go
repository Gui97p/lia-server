package main

import "github.com/spf13/cobra"

var rootCmd = &cobra.Command{
	Use:   "lia-admin",
	Short: "Lia's server configuration",
}

func init() {
	rootCmd.AddCommand(usersCmd)
	rootCmd.AddCommand(tokensCmd)
	rootCmd.AddCommand(messagesCmd)
	rootCmd.AddCommand(rulesCmd)
	rootCmd.AddCommand(memoriesCmd)
	rootCmd.AddCommand(capabilitiesCmd)
}
