package flags

import (
	"github.com/spf13/cobra"
)

// Constants for flag names used in the CLI application.
const (
	ConfigPathFlag = "config"
	JSONOutputFlag = "json"
)

// Helper function to add a --json flag to a command.
// Registers the --json output setting for all child commands of the provided Cobra command.
// It adds a persistent boolean flag with the given name, default value and description to the command.
func SetJSONOutputFlag(cmd *cobra.Command) {
	cmd.PersistentFlags().Bool(
		JSONOutputFlag,
		false,
		"get all outputs in json format (default false)",
	)
}

// Extracts the status of the JSON Output flag from the provided Cobra command.
// It returns true if the JSON Output flag is set (changed) and false otherwise.
func IsJSONOutputFlagChanged(cmd *cobra.Command) bool {
	return cmd.Flag(JSONOutputFlag).Changed
}

// Helper function to add a --config flag to a command.
// This flag allows users to specify the path to the CLI config.
// It sets up the flag with a default empty string value and provides a usage message.
func SetConfigPathFlag(cmd *cobra.Command, configPath *string) {
	cmd.Flags().StringVar(
		configPath,
		ConfigPathFlag,
		"",
		"the path to the CLI config. Supports .json",
	)
}

// IsConfigPathSpecified checks if the --config flag has been specified by the user.
// Returns true if the --config flag has been specified, false otherwise.
func IsConfigPathSpecified(cmd *cobra.Command) bool {
	return cmd.Flags().Changed(ConfigPathFlag)
}
