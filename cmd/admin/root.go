package main

import (
	"fmt"

	"github.com/spf13/cobra"
)

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

type choiceValue struct {
	value    string
	validate func(string) error
}

func (f *choiceValue) Set(s string) error {
	err := f.validate(s)
	if err != nil {
		return err
	}

	f.value = s
	return nil
}

func (f *choiceValue) Type() string   { return "string" }
func (f *choiceValue) String() string { return f.value }

func StringChoice(choices []string) *choiceValue {
	return &choiceValue{
		validate: func(s string) error {
			for _, choice := range choices {
				if s == choice {
					return nil
				}
			}
			return fmt.Errorf("must be one of %v", choices)
		},
	}
}
