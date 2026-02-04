package main

import (
	"fmt"
	"log/slog"
	"os"

	"github.com/jongio/azd-copilot/src/cmd/copilot/commands"
	"github.com/jongio/azd-copilot/src/internal/logging"

	"github.com/spf13/cobra"
)

var (
	outputFormat   string
	debugMode      bool
	structuredLogs bool
	cwdFlag        string
)

func main() {
	rootCmd := &cobra.Command{
		Use:   "copilot",
		Short: "Copilot - GitHub Copilot integration for Azure Developer CLI",
		Long:  `Copilot is an Azure Developer CLI extension that provides GitHub Copilot integration.`,
		PersistentPreRunE: func(cmd *cobra.Command, args []string) error {
			// Change working directory if --cwd is specified
			if cwdFlag != "" {
				if err := os.Chdir(cwdFlag); err != nil {
					return fmt.Errorf("failed to change to directory '%s': %w", cwdFlag, err)
				}
			}

			// Set global output format and debug mode
			if debugMode {
				os.Setenv("AZD_DEBUG", "true")
				slog.SetLogLoggerLevel(slog.LevelDebug)
			}

			// Configure logging
			logging.SetupLogger(debugMode, structuredLogs)

			// Log startup in debug mode
			if debugMode {
				logging.Debug("Starting azd copilot extension",
					"version", commands.Version,
					"command", cmd.Name(),
					"args", args,
					"cwd", cwdFlag,
				)
			}

			return nil
		},
	}

	// Add global flags
	rootCmd.PersistentFlags().StringVarP(&outputFormat, "output", "o", "default", "Output format (default, json)")
	rootCmd.PersistentFlags().BoolVar(&debugMode, "debug", false, "Enable debug logging")
	rootCmd.PersistentFlags().BoolVar(&structuredLogs, "structured-logs", false, "Enable structured JSON logging to stderr")
	rootCmd.PersistentFlags().StringVarP(&cwdFlag, "cwd", "C", "", "Sets the current working directory")

	// Register all commands
	rootCmd.AddCommand(
		commands.NewVersionCommand(),
		commands.NewListenCommand(),
	)

	if err := rootCmd.Execute(); err != nil {
		fmt.Fprintln(os.Stderr, err)
		os.Exit(1)
	}
}
