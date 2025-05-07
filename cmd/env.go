package cmd

import (
	"fmt"

	"github.com/spf13/cobra"
)

type EnvCommand struct {
	cmd     *cobra.Command
	options *CommandOptions
	root    *RootCommand
}

func NewEnvCommand(options *CommandOptions, root *RootCommand) *EnvCommand {
	cmd := &cobra.Command{
		Use:   "env",
		Short: "Generate a template environment file",
		Long:  `Creates a template .env file with default configuration values that can be customized.`,
		RunE: func(cmd *cobra.Command, args []string) error {
			cfg, log, _, err := root.LoadLoggerAndConfig()
			if err != nil {
				return fmt.Errorf("failed to initialize: %w", err)
			}

			if err := cfg.GenerateEnvFile(); err != nil {
				log.LogError("Failed to generate environment file: %v", err)
				return err
			}

			fmt.Printf("Environment file generated at: %s\n", cfg.EnvFile)
			return nil
		},
	}

	return &EnvCommand{
		cmd:     cmd,
		options: options,
		root:    root,
	}
}
