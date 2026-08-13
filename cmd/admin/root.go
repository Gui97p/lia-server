package main

import "github.com/spf13/cobra"

var rootCmd = &cobra.Command{
	Use:   "lia-admin",
	Short: "Administração local do servidor Lia",
}

func init() {
	rootCmd.AddCommand(usersCmd)
}
