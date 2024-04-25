package server

import (
	"encoding/json"
	"fmt"

	"github.com/spf13/cobra"
	"github.com/sx-network/sx-reporter/command"
	"github.com/sx-network/sx-reporter/command/flags"
	"github.com/sx-network/sx-reporter/command/helper"
	"github.com/sx-network/sx-reporter/infra/server"
)

var (
	serverParams = &server.YAMLServerConfig{
		YAMLReporterConfig: &server.YAMLReporterConfig{},
	}
)

// This command starts the SX Reporter client.
func GetCommand() *cobra.Command {
	serverCommand := &cobra.Command{
		Use:     "server",
		Short:   "The default command that starts the SX Reporter client, by bootstrapping all modules together",
		PreRunE: runPreRun,
		Run:     runCommand,
	}

	flags.SetConfigPathFlag(serverCommand, &serverParams.ConfigPath)

	return serverCommand
}

// The pre-run hook function for Cobra commands.
// It sets the JSON log format based on the command's flag,
// and initializes configuration parameters from a specified file,
// if a config file is specified in the command.
// It returns an error if there's an issue initializing the config from file.
func runPreRun(cmd *cobra.Command, _ []string) error {
	serverParams.SetJSONLogFormat(flags.IsJSONOutputFlagChanged(cmd))

	if flags.IsConfigPathSpecified(cmd) {
		if err := serverParams.InitServerConfigFromFile(); err != nil {
			return err
		}
	} else {
		return fmt.Errorf("config.yml file not specified. Please provide a path to the CLI config using the '--%s' flag", flags.ConfigPathFlag)
	}

	if err := serverParams.InitRawParams(); err != nil {
		return err
	}

	return nil
}

// The main function executed when a command is run in the CLI.
func runCommand(cmd *cobra.Command, _ []string) {
	outputter := command.InitializeOutputter(cmd)

	if err := runServerLoop(serverParams.GenerateConfig(), outputter); err != nil {
		outputter.SetError(err)
		outputter.WriteOutput()

		return
	}
}

// Simulates the main server loop with a provided configuration and outputter.
// It takes a ServerConfig containing server configuration details and an OutputFormatter for formatting output.
func runServerLoop(
	serverConfig *server.ServerConfig,
	outputter command.OutputFormatter,
) error {
	jsonBytes, _ := json.Marshal(serverConfig)
	fmt.Println(string(jsonBytes))

	serverInstance, err := server.NewServer(serverConfig)
	if err != nil {
		return err
	}

	return helper.HandleSignals(serverInstance.Close, outputter)
}
