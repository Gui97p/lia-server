package main

import (
	"context"
	"encoding/json"
	"fmt"

	"github.com/Gui97p/lia-server/internal/auth"
	"github.com/google/uuid"
	"github.com/spf13/cobra"
)

var capabilitiesCmd = &cobra.Command{
	Use:   "capabilities",
	Short: "Client Capabilities Catalog Management",
}

var capabilitiesListCmd = &cobra.Command{
	Use:          "list",
	Short:        "Lists every capability in the catalog",
	SilenceUsage: true,
	RunE: func(cmd *cobra.Command, args []string) error {
		ctx := context.Background()

		cs, err := capabilitiesStore.ListAll(ctx)
		if err != nil {
			return err
		}

		for _, c := range cs {
			source := ""
			if c.Source != nil {
				source = fmt.Sprintf(" (source: %s)", *c.Source)
			}
			fmt.Printf("(%s) [%s]%s %s - %s\n\tparams: %s\n", c.ID, c.TrustLevel, source, c.Name, c.Description, c.Parameters)
		}

		return nil
	},
}

var capabilitiesCreateCmd = &cobra.Command{
	Use:          "add",
	Short:        "Registers a new client capability in the catalog",
	SilenceUsage: true,
	RunE: func(cmd *cobra.Command, args []string) error {
		name, _ := cmd.Flags().GetString("name")
		description, _ := cmd.Flags().GetString("description")
		paramsStr, _ := cmd.Flags().GetString("parameters")
		trustStr, _ := cmd.Flags().GetString("trust")
		source, _ := cmd.Flags().GetString("source")
		ctx := context.Background()

		if !json.Valid([]byte(paramsStr)) {
			return fmt.Errorf("--parameters must be valid JSON")
		}

		trust := auth.TrustLevel(trustStr)
		switch trust {
		case auth.Anonymous, auth.Identified, auth.Authenticated, auth.Trusted:
		default:
			return fmt.Errorf("invalid --trust: %s", trustStr)
		}

		var sourcePtr *string
		if source != "" {
			sourcePtr = &source
		}

		c, err := capabilitiesStore.Create(ctx, name, description, json.RawMessage(paramsStr), trust, sourcePtr)
		if err != nil {
			return err
		}

		fmt.Printf("capability successfully created: id=%s name=%s\n", c.ID, c.Name)
		return nil
	},
}

var capabilitiesUpdateCmd = &cobra.Command{
	Use:          "update",
	Short:        "Updates a capability's description, parameters and/or trust level",
	SilenceUsage: true,
	RunE: func(cmd *cobra.Command, args []string) error {
		idStr, _ := cmd.Flags().GetString("id")
		description, _ := cmd.Flags().GetString("description")
		paramsStr, _ := cmd.Flags().GetString("parameters")
		trustStr, _ := cmd.Flags().GetString("trust")
		ctx := context.Background()

		id, err := uuid.Parse(idStr)
		if err != nil {
			return err
		}

		if description == "" && paramsStr == "" && trustStr == "" {
			return fmt.Errorf("provide at least one of --description, --parameters or --trust")
		}

		if description != "" {
			if err := capabilitiesStore.SetDescription(ctx, id, description); err != nil {
				return err
			}
		}
		if paramsStr != "" {
			if !json.Valid([]byte(paramsStr)) {
				return fmt.Errorf("--parameters must be valid JSON")
			}
			if err := capabilitiesStore.SetParameters(ctx, id, json.RawMessage(paramsStr)); err != nil {
				return err
			}
		}
		if trustStr != "" {
			trust := auth.TrustLevel(trustStr)
			switch trust {
			case auth.Anonymous, auth.Identified, auth.Authenticated, auth.Trusted:
			default:
				return fmt.Errorf("invalid --trust: %s", trustStr)
			}
			if err := capabilitiesStore.SetTrustLevel(ctx, id, trust); err != nil {
				return err
			}
		}

		fmt.Printf("capability successfully updated: id=%s\n", id)
		return nil
	},
}

var capabilitiesDeleteCmd = &cobra.Command{
	Use:          "delete",
	Short:        "Removes a capability from the catalog",
	SilenceUsage: true,
	RunE: func(cmd *cobra.Command, args []string) error {
		idStr, _ := cmd.Flags().GetString("id")
		ctx := context.Background()

		id, err := uuid.Parse(idStr)
		if err != nil {
			return err
		}

		if err := capabilitiesStore.Delete(ctx, id); err != nil {
			return err
		}

		fmt.Printf("capability successfully deleted: id=%s\n", id)
		return nil
	},
}

func init() {
	capabilitiesCreateCmd.Flags().String("name", "", "capability's name (must match what the client announces on handshake)")
	capabilitiesCreateCmd.MarkFlagRequired("name")
	capabilitiesCreateCmd.Flags().String("description", "", "description shown to the LLM")
	capabilitiesCreateCmd.MarkFlagRequired("description")
	capabilitiesCreateCmd.Flags().String("parameters", "", "JSON schema for the capability's parameters")
	capabilitiesCreateCmd.MarkFlagRequired("parameters")
	capabilitiesCreateCmd.Flags().String("trust", string(auth.Authenticated), "required trust level (anonymous/identified/authenticated/trusted)")
	capabilitiesCreateCmd.Flags().String("source", "", "optional: name of another capability this one depends on")

	capabilitiesUpdateCmd.Flags().String("id", "", "target capability's id")
	capabilitiesUpdateCmd.MarkFlagRequired("id")
	capabilitiesUpdateCmd.Flags().String("description", "", "new description")
	capabilitiesUpdateCmd.Flags().String("parameters", "", "new JSON schema for parameters")
	capabilitiesUpdateCmd.Flags().String("trust", "", "new required trust level")

	capabilitiesDeleteCmd.Flags().String("id", "", "target capability's id")
	capabilitiesDeleteCmd.MarkFlagRequired("id")

	capabilitiesCmd.AddCommand(capabilitiesListCmd)
	capabilitiesCmd.AddCommand(capabilitiesCreateCmd)
	capabilitiesCmd.AddCommand(capabilitiesUpdateCmd)
	capabilitiesCmd.AddCommand(capabilitiesDeleteCmd)
}
