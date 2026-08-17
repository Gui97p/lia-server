package main

import (
	"context"
	"fmt"

	"github.com/spf13/cobra"
)

var messagesCmd = &cobra.Command{
	Use:   "messages",
	Short: "Messages Management",
}

var messagesListCmd = &cobra.Command{
	Use:          "list",
	Short:        "Lists an user's messages",
	SilenceUsage: true,
	RunE: func(cmd *cobra.Command, args []string) error {
		username, _ := cmd.Flags().GetString("username")
		limit, _ := cmd.Flags().GetInt("limit")
		ctx := context.Background()

		u, err := usersStore.GetByUsername(ctx, username)
		if err != nil {
			return err
		}

		messages, err := messagesStore.ListByUser(ctx, u.ID, limit)

		fmt.Printf("Listing messages for user id=%s\n\n", u.ID)
		for _, message := range messages {
			fmt.Printf("[%s][%s] -- %s\n", message.CreatedAt.Format("02/01/2006 03:04:05"), message.Role, message.Content)
		}

		return nil
	},
}

func init() {
	messagesListCmd.Flags().String("username", "", "target's username")
	messagesListCmd.MarkFlagRequired("username")

	messagesListCmd.Flags().Int("limit", 5, "target's username")

	messagesCmd.AddCommand(messagesListCmd)
}
