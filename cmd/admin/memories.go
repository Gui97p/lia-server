package main

import (
	"context"
	"fmt"

	"github.com/Gui97p/lia-server/internal/memories"
	"github.com/google/uuid"
	"github.com/spf13/cobra"
)

var memoriesCmd = &cobra.Command{
	Use:   "memories",
	Short: "Memories Management",
}

func printMemories(header string, ms []memories.Memory) {
	fmt.Printf("%s\n\n", header)
	for _, m := range ms {
		category := ""
		if m.Category != nil {
			category = fmt.Sprintf("[%s]", *m.Category)
		}
		fmt.Printf("(%s) [%s]%s - %s\n", m.ID, m.Scope, category, m.Fact)
	}
}

var memoriesListCmd = &cobra.Command{
	Use:          "list",
	Short:        "Lists memories for a user, or by scope (global/private)",
	SilenceUsage: true,
	RunE: func(cmd *cobra.Command, args []string) error {
		username, _ := cmd.Flags().GetString("username")
		scope, _ := cmd.Flags().GetString("scope")
		limit, _ := cmd.Flags().GetInt("limit")
		ctx := context.Background()

		if username != "" {
			u, err := usersStore.GetByUsername(ctx, username)
			if err != nil {
				return err
			}

			ms, err := memoriesStore.ListByUser(ctx, u.ID, limit)
			if err != nil {
				return err
			}
			printMemories(fmt.Sprintf("Memories for user id=%s", u.ID), ms)
			return nil
		}

		ms, err := memoriesStore.ListByScope(ctx, memories.MemoryScope(scope), limit)
		if err != nil {
			return err
		}
		printMemories(fmt.Sprintf("Memories for scope=%s", scope), ms)
		return nil
	},
}

var memoriesCreateCmd = &cobra.Command{
	Use:          "add",
	Short:        "Creates a new memory",
	SilenceUsage: true,
	RunE: func(cmd *cobra.Command, args []string) error {
		username, _ := cmd.Flags().GetString("username")
		scopeStr, _ := cmd.Flags().GetString("scope")
		fact, _ := cmd.Flags().GetString("fact")
		category, _ := cmd.Flags().GetString("category")
		ctx := context.Background()

		scope := memories.MemoryScope(scopeStr)

		var userID *uuid.UUID
		if scope == memories.User {
			if username == "" {
				return fmt.Errorf("--username is required for scope=user")
			}
			u, err := usersStore.GetByUsername(ctx, username)
			if err != nil {
				return err
			}
			userID = &u.ID
		}

		m, err := memoriesStore.Create(ctx, scope, fact, userID)
		if err != nil {
			return err
		}

		if category != "" {
			if err := memoriesStore.SetCategory(ctx, m.ID, category); err != nil {
				return err
			}
		}

		fmt.Printf("memory successfully created: id=%s\n", m.ID)
		return nil
	},
}

var memoriesUpdateCmd = &cobra.Command{
	Use:          "update",
	Short:        "Updates a memory's fact and/or category",
	SilenceUsage: true,
	RunE: func(cmd *cobra.Command, args []string) error {
		idStr, _ := cmd.Flags().GetString("id")
		fact, _ := cmd.Flags().GetString("fact")
		category, _ := cmd.Flags().GetString("category")
		ctx := context.Background()

		id, err := uuid.Parse(idStr)
		if err != nil {
			return err
		}

		if fact == "" && category == "" {
			return fmt.Errorf("provide at least one of --fact or --category")
		}

		if fact != "" {
			if err := memoriesStore.SetFact(ctx, id, fact); err != nil {
				return err
			}
		}
		if category != "" {
			if err := memoriesStore.SetCategory(ctx, id, category); err != nil {
				return err
			}
		}

		fmt.Printf("memory successfully updated: id=%s\n", id)
		return nil
	},
}

var memoriesDeleteCmd = &cobra.Command{
	Use:          "delete",
	Short:        "Deletes a memory",
	SilenceUsage: true,
	RunE: func(cmd *cobra.Command, args []string) error {
		idStr, _ := cmd.Flags().GetString("id")
		ctx := context.Background()

		id, err := uuid.Parse(idStr)
		if err != nil {
			return err
		}

		if err := memoriesStore.Delete(ctx, id); err != nil {
			return err
		}

		fmt.Printf("memory successfully deleted: id=%s\n", id)
		return nil
	},
}

func init() {
	memoriesListCmd.Flags().String("username", "", "target's username (lists USER-scope memories)")
	memoriesListCmd.Flags().String("scope", "", "scope to list (global/private) — ignored if --username is set")
	memoriesListCmd.Flags().Int("limit", 20, "max number of memories to show")
	memoriesListCmd.MarkFlagsOneRequired("username", "scope")

	memoriesCreateCmd.Flags().String("username", "", "target user (required if --scope=user)")
	memoriesCreateCmd.Flags().String("scope", "", "user/global/private")
	memoriesCreateCmd.MarkFlagRequired("scope")
	memoriesCreateCmd.Flags().String("fact", "", "the memory's content")
	memoriesCreateCmd.MarkFlagRequired("fact")
	memoriesCreateCmd.Flags().String("category", "", "optional category")

	memoriesUpdateCmd.Flags().String("id", "", "target memory's id")
	memoriesUpdateCmd.MarkFlagRequired("id")
	memoriesUpdateCmd.Flags().String("fact", "", "new fact content")
	memoriesUpdateCmd.Flags().String("category", "", "new category")

	memoriesDeleteCmd.Flags().String("id", "", "target memory's id")
	memoriesDeleteCmd.MarkFlagRequired("id")

	memoriesCmd.AddCommand(memoriesListCmd)
	memoriesCmd.AddCommand(memoriesCreateCmd)
	memoriesCmd.AddCommand(memoriesUpdateCmd)
	memoriesCmd.AddCommand(memoriesDeleteCmd)
}
