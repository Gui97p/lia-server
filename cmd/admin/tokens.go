package main

import (
	"context"
	"fmt"

	"github.com/Gui97p/lia-server/internal/auth"
	"github.com/spf13/cobra"
)

var tokensCmd = &cobra.Command{
	Use:   "tokens",
	Short: "Tokens Management",
}

var tokensGenerateCmd = &cobra.Command{
	Use:          "generate",
	Short:        "Generates a new JWT token based on username",
	SilenceUsage: true,
	RunE: func(cmd *cobra.Command, args []string) error {
		username, _ := cmd.Flags().GetString("username")
		ctx := context.Background()

		u, err := userStore.GetByUsername(ctx, username)
		if err != nil {
			return err
		}

		token, err := auth.GenerateToken(jwtSecret, u)
		if err != nil {
			return err
		}

		fmt.Printf("token successfully generated for id=%s\n\n%s\n", u.ID, token)
		return nil
	},
}

var tokensBumpCmd = &cobra.Command{
	Use:          "bump",
	Short:        "bumps token version for a user, invalidating every token generated",
	SilenceUsage: true,
	RunE: func(cmd *cobra.Command, args []string) error {
		username, _ := cmd.Flags().GetString("username")
		ctx := context.Background()

		u, err := userStore.GetByUsername(ctx, username)
		if err != nil {
			return err
		}

		err = userStore.BumpTokenVersion(ctx, u.ID)
		if err != nil {
			return err
		}

		fmt.Printf("Token Version successfuly updated for id=%s\n", u.ID)
		return nil
	},
}

func init() {
	tokensGenerateCmd.Flags().String("username", "", "target's username")
	tokensGenerateCmd.MarkFlagRequired("username")

	tokensBumpCmd.Flags().String("username", "", "target's username")
	tokensBumpCmd.MarkFlagRequired("username")

	tokensCmd.AddCommand(tokensGenerateCmd)
	tokensCmd.AddCommand(tokensBumpCmd)
}
