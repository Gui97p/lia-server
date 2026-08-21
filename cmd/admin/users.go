package main

import (
	"context"
	"fmt"

	"github.com/Gui97p/lia-server/internal/crypto"
	"github.com/Gui97p/lia-server/internal/providers"
	"github.com/Gui97p/lia-server/internal/users"
	"github.com/google/uuid"
	"github.com/spf13/cobra"
)

var usersCmd = &cobra.Command{
	Use:   "users",
	Short: "User Management",
}

var usersFindCmd = &cobra.Command{
	Use:          "find",
	Short:        "Finds a user based on ID or username",
	SilenceUsage: true,
	RunE: func(cmd *cobra.Command, args []string) error {
		username, _ := cmd.Flags().GetString("username")
		id, _ := cmd.Flags().GetString("id")
		ctx := context.Background()

		var u *users.User
		var err error

		if id != "" {
			var userId uuid.UUID
			userId, err = uuid.Parse(id)
			if err != nil {
				return err
			}
			u, err = usersStore.GetByID(ctx, userId)
		} else if username != "" {
			u, err = usersStore.GetByUsername(ctx, username)
		}

		if err != nil {
			return err
		}
		if u == nil {
			return users.ErrNotFound
		}

		fmt.Printf("id = %s\nusername = %s\ntoken_version = %d\ncreated_at = %s\nupdated_at = %s\n", u.ID, u.Username, u.TokenVersion, u.CreatedAt, u.UpdatedAt)
		return nil
	},
}

var usersCreateCmd = &cobra.Command{
	Use:          "create",
	Short:        "Creates a new user",
	SilenceUsage: true,
	RunE: func(cmd *cobra.Command, args []string) error {
		username, _ := cmd.Flags().GetString("username")

		u, err := usersStore.Create(context.Background(), username)
		if err != nil {
			return err
		}

		fmt.Printf("user successfully created: id=%s username=%s\n", u.ID, u.Username)
		return nil
	},
}

var usersDeleteCmd = &cobra.Command{
	Use:          "delete",
	Short:        "Deletes an user",
	SilenceUsage: true,
	RunE: func(cmd *cobra.Command, args []string) error {
		username, _ := cmd.Flags().GetString("username")
		id, _ := cmd.Flags().GetString("id")
		ctx := context.Background()

		var u *users.User
		var err error

		if id != "" {
			var userId uuid.UUID
			userId, err = uuid.Parse(id)
			if err != nil {
				return err
			}
			u, err = usersStore.GetByID(ctx, userId)
		} else if username != "" {
			u, err = usersStore.GetByUsername(ctx, username)
		}

		if err != nil {
			return err
		}
		if u == nil {
			return users.ErrNotFound
		}

		err = usersStore.Delete(ctx, u.ID)
		if err != nil {
			return err
		}

		fmt.Printf("user successfully deleted: id=%s\n", u.ID)
		return nil
	},
}

var usersSetKeyCmd = &cobra.Command{
	Use:          "set-key",
	Short:        "Encrypt and update user's api key",
	SilenceUsage: true,
	RunE: func(cmd *cobra.Command, args []string) error {
		username, _ := cmd.Flags().GetString("username")
		provider, _ := cmd.Flags().GetString("provider")
		key, _ := cmd.Flags().GetString("key")
		reset, _ := cmd.Flags().GetBool("reset")
		ctx := context.Background()

		u, err := usersStore.GetByUsername(ctx, username)
		if err != nil {
			return err
		}

		if reset {
			err = providersStore.ResetKey(ctx, u.ID, providers.ProviderName(provider))
			if err != nil {
				return err
			}
		} else {
			encryptedKey, err := crypto.Encrypt(key, encryptionKey)
			if err != nil {
				return err
			}

			err = providersStore.SetKey(ctx, u.ID, providers.ProviderName(provider), encryptedKey)
			if err != nil {
				return err
			}
		}

		fmt.Printf("key successfully updated: id=%s\n", u.ID)
		return nil
	},
}

func init() {
	usersFindCmd.Flags().String("username", "", "target's username")
	usersFindCmd.Flags().String("id", "", "target's id")
	usersFindCmd.MarkFlagsOneRequired("username", "id")

	usersCreateCmd.Flags().String("username", "", "new user's username")
	usersCreateCmd.MarkFlagRequired("username")

	usersDeleteCmd.Flags().String("username", "", "target's username")
	usersDeleteCmd.Flags().String("id", "", "target's id")
	usersDeleteCmd.MarkFlagsOneRequired("username", "id")

	usersSetKeyCmd.Flags().String("username", "", "target's username")
	usersSetKeyCmd.MarkFlagRequired("username")
	usersSetKeyCmd.Flags().VarP(StringChoice(providers.ProviderList), "provider", "", "provider to set key")
	usersSetKeyCmd.MarkFlagRequired("provider")
	usersSetKeyCmd.Flags().String("key", "", "API key")
	usersSetKeyCmd.Flags().Bool("reset", false, "use this flag to reset key to null")
	usersSetKeyCmd.MarkFlagsMutuallyExclusive("key", "reset")

	usersCmd.AddCommand(usersFindCmd)
	usersCmd.AddCommand(usersCreateCmd)
	usersCmd.AddCommand(usersDeleteCmd)
	usersCmd.AddCommand(usersSetKeyCmd)
}
