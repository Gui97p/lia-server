package main

import (
	"context"
	"fmt"

	"github.com/Gui97p/lia-server/internal/users"
	"github.com/google/uuid"
	"github.com/spf13/cobra"
)

var rulesCmd = &cobra.Command{
	Use:   "rules",
	Short: "Behavior Rules Management",
}

var rulesListCmd = &cobra.Command{
	Use:          "list",
	Short:        "Lists global and user rules based on ID or username",
	SilenceUsage: true,
	RunE: func(cmd *cobra.Command, args []string) error {
		username, _ := cmd.Flags().GetString("username")
		id, _ := cmd.Flags().GetString("id")
		ctx := context.Background()

		var u *users.User
		var err error

		if id != "" {
			var userID uuid.UUID
			userID, err = uuid.Parse(id)
			if err != nil {
				return err
			}
			u, err = usersStore.GetByID(ctx, userID)
		} else if username != "" {
			u, err = usersStore.GetByUsername(ctx, username)
		}

		if err != nil {
			return err
		}

		var userID uuid.UUID
		if u != nil {
			userID = u.ID
		}

		rules, err := behaviorRulesStore.ListActive(ctx, userID)
		if err != nil {
			return err
		}

		text := ""
		for _, rule := range rules {
			if rule.UserID != nil {
				text += fmt.Sprintf("[%s] ", u.Username)
			} else {
				text += "[global] "
			}
			text += fmt.Sprintf("(%s) - %s\n", rule.ID, rule.Rule)
		}

		fmt.Printf("Behavior Rules:\n%s", text)

		return nil
	},
}

var rulesCreateCmd = &cobra.Command{
	Use:          "add",
	Short:        "Creates a new rule",
	SilenceUsage: true,
	RunE: func(cmd *cobra.Command, args []string) error {
		username, _ := cmd.Flags().GetString("username")
		rule, _ := cmd.Flags().GetString("rule")
		ctx := context.Background()

		var userID *uuid.UUID
		if username != "" {
			u, err := usersStore.GetByUsername(ctx, username)
			if err != nil {
				return err
			}
			userID = &u.ID
		}

		br, err := behaviorRulesStore.Create(ctx, userID, rule)
		if err != nil {
			return err
		}

		fmt.Printf("rule successfully created: id=%s\n", br.ID)
		return nil
	},
}

var rulesUpdateCmd = &cobra.Command{
	Use:          "update",
	Short:        "Updates a rule",
	SilenceUsage: true,
	RunE: func(cmd *cobra.Command, args []string) error {
		idStr, _ := cmd.Flags().GetString("id")
		rule, _ := cmd.Flags().GetString("rule")
		ctx := context.Background()

		id, err := uuid.Parse(idStr)
		if err != nil {
			return err
		}

		err = behaviorRulesStore.Update(ctx, id, rule)
		if err != nil {
			return err
		}

		fmt.Printf("rule successfully updated: id=%s\n", id)
		return nil
	},
}

var rulesDeleteCmd = &cobra.Command{
	Use:          "delete",
	Short:        "Deletes a rule",
	SilenceUsage: true,
	RunE: func(cmd *cobra.Command, args []string) error {
		idStr, _ := cmd.Flags().GetString("id")
		ctx := context.Background()

		id, err := uuid.Parse(idStr)
		if err != nil {
			return err
		}

		err = behaviorRulesStore.Delete(ctx, id)
		if err != nil {
			return err
		}

		fmt.Printf("rule successfully deleted: id=%s\n", id)
		return nil
	},
}

func init() {
	rulesListCmd.Flags().String("username", "", "target rules user")
	rulesListCmd.Flags().String("id", "", "target's id")

	rulesCreateCmd.Flags().String("username", "", "new rule's user to link (omit for a global rule)")
	rulesCreateCmd.Flags().String("rule", "", "new rule's content")
	rulesCreateCmd.MarkFlagRequired("rule")

	rulesUpdateCmd.Flags().String("id", "", "target rule's id")
	rulesUpdateCmd.Flags().String("rule", "", "new rule content")
	rulesUpdateCmd.MarkFlagRequired("id")
	rulesUpdateCmd.MarkFlagRequired("rule")

	rulesDeleteCmd.Flags().String("id", "", "target's id")
	rulesDeleteCmd.MarkFlagRequired("id")

	rulesCmd.AddCommand(rulesListCmd)
	rulesCmd.AddCommand(rulesCreateCmd)
	rulesCmd.AddCommand(rulesUpdateCmd)
	rulesCmd.AddCommand(rulesDeleteCmd)
}
