package main

import (
	"context"
	"fmt"

	"github.com/google/uuid"
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
		if err != nil {
			return err
		}

		fmt.Printf("Listing messages for user id=%s\n\n", u.ID)
		for _, message := range messages {
			fmt.Printf("(%s) [%s][%s] -- %s\n", message.ID, message.CreatedAt.Format("02/01/2006 03:04:05"), message.Role, message.Content)
		}

		return nil
	},
}

var messagesListByTaskCmd = &cobra.Command{
	Use:          "list-by-task",
	Short:        "Lists messages belonging to a specific task",
	SilenceUsage: true,
	RunE: func(cmd *cobra.Command, args []string) error {
		username, _ := cmd.Flags().GetString("username")
		taskIDStr, _ := cmd.Flags().GetString("task-id")
		limit, _ := cmd.Flags().GetInt("limit")
		ctx := context.Background()

		u, err := usersStore.GetByUsername(ctx, username)
		if err != nil {
			return err
		}

		taskID, err := uuid.Parse(taskIDStr)
		if err != nil {
			return err
		}

		messages, err := messagesStore.ListByTask(ctx, u.ID, taskID, limit)
		if err != nil {
			return err
		}

		fmt.Printf("Listing messages for task id=%s\n\n", taskID)
		for _, message := range messages {
			fmt.Printf("(%s) [%s][%s] -- %s\n", message.ID, message.CreatedAt.Format("02/01/2006 03:04:05"), message.Role, message.Content)
		}

		return nil
	},
}

var messagesDeleteCmd = &cobra.Command{
	Use:          "delete",
	Short:        "Deletes a message",
	SilenceUsage: true,
	RunE: func(cmd *cobra.Command, args []string) error {
		idStr, _ := cmd.Flags().GetString("id")
		ctx := context.Background()

		id, err := uuid.Parse(idStr)
		if err != nil {
			return err
		}

		if err := messagesStore.Delete(ctx, id); err != nil {
			return err
		}

		fmt.Printf("message successfully deleted: id=%s\n", id)
		return nil
	},
}

func init() {
	messagesListCmd.Flags().String("username", "", "target's username")
	messagesListCmd.MarkFlagRequired("username")

	messagesListCmd.Flags().Int("limit", 5, "max number of messages to show")

	messagesListByTaskCmd.Flags().String("username", "", "target's username")
	messagesListByTaskCmd.MarkFlagRequired("username")
	messagesListByTaskCmd.Flags().String("task-id", "", "target task's id")
	messagesListByTaskCmd.MarkFlagRequired("task-id")
	messagesListByTaskCmd.Flags().Int("limit", 20, "max number of messages to show")

	messagesDeleteCmd.Flags().String("id", "", "target message's id")
	messagesDeleteCmd.MarkFlagRequired("id")

	messagesCmd.AddCommand(messagesListCmd)
	messagesCmd.AddCommand(messagesListByTaskCmd)
	messagesCmd.AddCommand(messagesDeleteCmd)
}
