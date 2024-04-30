package init

import (
	"fmt"

	"github.com/spf13/cobra"
	"github.com/sx-network/sx-reporter/command"
)

const (
	// maxInitNum is the maximum value for "num" flag
	maxInitNum = 30
)

var (
	errInvalidNum = fmt.Errorf("num flag value should be between 1 and %d", maxInitNum)
	basicParams   initParams
	initNumber    int
)

// Returns a Cobra command for initializing private keys in the SX Reporter node.
func GetCommand() *cobra.Command {
	secretsInitCmd := &cobra.Command{
		Use:     "init",
		Short:   "Initializes private keys for the SX Reporter node to the specified Secrets Manager",
		PreRunE: runPreRun,
		Run:     runCommand,
	}

	setFlags(secretsInitCmd)

	return secretsInitCmd
}

// Configures flags for the given Cobra command. It sets up flags for
// data directory, config path, number of secrets to create, and ECDSA key generation.
func setFlags(cmd *cobra.Command) {
	cmd.Flags().StringVar(
		&basicParams.dataDir,
		dataDirFlag,
		"",
		"the directory for the SX Reporter Node data if the local FS is used",
	)

	cmd.Flags().StringVar(
		&basicParams.configPath,
		configFlag,
		"",
		"the path to the SecretsManager config file, "+
			"if omitted, the local FS secrets manager is used",
	)

	cmd.Flags().IntVar(
		&initNumber,
		numFlag,
		1,
		"the flag indicating how many secrets should be created, only for the local FS",
	)

	// Don't accept data-dir and config flags because they are related to different secrets managers.
	// data-dir is about the local FS as secrets storage, config is about remote secrets manager.
	cmd.MarkFlagsMutuallyExclusive(dataDirFlag, configFlag)

	// num flag should be used with data-dir flag only so it should not be used with config flag.
	cmd.MarkFlagsMutuallyExclusive(numFlag, configFlag)

	cmd.Flags().BoolVar(
		&basicParams.generatesECDSA,
		ecdsaFlag,
		true,
		"the flag indicating whether new ECDSA key is created",
	)
}

// It validates the number of secrets to create and checks if the flags are valid.
func runPreRun(_ *cobra.Command, _ []string) error {
	if initNumber < 1 || initNumber > maxInitNum {
		return errInvalidNum
	}

	return basicParams.validateFlags()
}

// Creates a list of parameters based on basicParams and initNumber,
// and iterates through the list to initialize secrets and collect results.
func runCommand(cmd *cobra.Command, _ []string) {
	outputter := command.InitializeOutputter(cmd)
	defer outputter.WriteOutput()

	paramsList := newParamsList(basicParams, initNumber)
	results := make(Results, len(paramsList))

	for i, params := range paramsList {
		if err := params.initSecrets(); err != nil {
			outputter.SetError(err)

			return
		}

		res, err := params.getResult()
		if err != nil {
			outputter.SetError(err)

			return
		}

		results[i] = res
	}

	outputter.SetCommandResult(results)
}

// Creates a list of initParams based on the given parameters and number.
// If the number is 1, it returns a list containing the input parameters.
// Otherwise, it generates a list of parameters with incremented data directory names
// and the same ECDSA key generation flag.
func newParamsList(params initParams, num int) []initParams {
	if num == 1 {
		return []initParams{params}
	}

	paramsList := make([]initParams, num)
	for i := 1; i <= num; i++ {
		paramsList[i-1] = initParams{
			dataDir:        fmt.Sprintf("%s%d", params.dataDir, i),
			generatesECDSA: params.generatesECDSA,
		}
	}

	return paramsList
}
