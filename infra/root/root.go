package root

import (
	"fmt"
	"os"

	"github.com/spf13/cobra"
	"github.com/sx-network/sx-reporter/command/flags"
	"github.com/sx-network/sx-reporter/command/secrets"
	serverCommand "github.com/sx-network/sx-reporter/command/server"
)

// Represents the root command of the application.
// It contains methods to initialize and execute the main command.
type RootCommand struct {
	baseCmd *cobra.Command
}

// Creates a new instance of RootCommand, initializes its base command with a short description, and registers subcommands.
func NewRootCommand() *RootCommand {
	rootCommand := &RootCommand{
		baseCmd: &cobra.Command{
			Short: "SX Reporter is ...",
		},
	}
	flags.SetJSONOutputFlag(rootCommand.baseCmd)
	rootCommand.registerSubCommands()

	return rootCommand
}

// Adds subcommands to the base command of the RootCommand, such as serverCommand.
func (rc *RootCommand) registerSubCommands() {
	rc.baseCmd.AddCommand(
		serverCommand.GetCommand(),
		secrets.GetCommand(),
	)
}

// Runs the base command of the RootCommand.
// If an error occurs during execution, it prints the error to stderr and exits the program with an exit code of 1.
func (rc *RootCommand) Execute() {
	if err := rc.baseCmd.Execute(); err != nil {
		_, _ = fmt.Fprintln(os.Stderr, err)

		os.Exit(1)
	}
}
