package secrets

import (
	"github.com/spf13/cobra"
	initCmd "github.com/sx-network/sx-reporter/command/secrets/init"
)

func GetCommand() *cobra.Command {
	secretsCmd := &cobra.Command{
		Use:   "secrets",
		Short: "Top level SecretsManager command for interacting with secrets functionality. Only accepts subcommands.",
	}

	registerSubcommands(secretsCmd)

	return secretsCmd
}

func registerSubcommands(baseCmd *cobra.Command) {
	baseCmd.AddCommand(
		initCmd.GetCommand(),
	)
}
