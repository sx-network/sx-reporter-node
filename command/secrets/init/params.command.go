package init

import (
	"errors"

	"github.com/sx-network/sx-reporter/command"
	"github.com/sx-network/sx-reporter/command/helper"
	"github.com/sx-network/sx-reporter/infra/secrets"
	helperSecret "github.com/sx-network/sx-reporter/infra/secrets/helper"
)

// Defining flag names for data directory, config path, ECDSA key generation,
// and number of secrets to create.
const (
	dataDirFlag = "data-dir"
	configFlag  = "config"
	ecdsaFlag   = "ecdsa"
	numFlag     = "num"
)

// Errors related to invalid secrets configuration, missing parameters,
// and unsupported secrets manager type.
var (
	errInvalidConfig   = errors.New("invalid secrets configuration")
	errInvalidParams   = errors.New("no config file or data directory passed in")
	errUnsupportedType = errors.New("unsupported secrets manager")
)

// Holds initialization parameters including data directory,
// config path, ECDSA key generation flag, secrets manager, and secrets manager config.
type initParams struct {
	dataDir        string
	configPath     string
	generatesECDSA bool
	secretsManager secrets.SecretsManager
	secretsConfig  *secrets.SecretsManagerConfig
}

// Checks if the initialization parameters are valid.
// It ensures that either a data directory or a config path is provided.
func (ip *initParams) validateFlags() error {
	if ip.dataDir == "" && ip.configPath == "" {
		return errInvalidParams
	}

	return nil
}

// Initializes secrets by setting up the secrets manager and validator key.
func (ip *initParams) initSecrets() error {
	if err := ip.initSecretsManager(); err != nil {
		return err
	}

	if err := ip.initValidatorKey(); err != nil {
		return err
	}

	return nil
}

// Initializes the secrets manager based on the config path
// or initializes a local secrets manager if no config path is provided.
func (ip *initParams) initSecretsManager() error {
	var err error
	if ip.hasConfigPath() {
		if err = ip.parseConfig(); err != nil {
			return err
		}

		ip.secretsManager, err = helper.InitCloudSecretsManager(ip.secretsConfig)

		return err
	}

	return ip.initLocalSecretsManager()
}

// Initializes a local secrets manager with the provided data directory.
func (ip *initParams) initLocalSecretsManager() error {
	local, err := helperSecret.SetupLocalSecretsManager(ip.dataDir)
	if err != nil {
		return err
	}

	ip.secretsManager = local

	return nil
}

// Initializes the validator key if ECDSA key generation is enabled.
func (ip *initParams) initValidatorKey() error {
	var err error

	if ip.generatesECDSA {
		if _, err = helperSecret.InitECDSAValidatorKey(ip.secretsManager); err != nil {
			return err
		}
	}

	return nil
}

// Retrieves the results of the initialization process, including the validator address and node ID.
func (ip *initParams) getResult() (command.CommandResult, error) {
	var (
		res = &SecretsInitResult{}
		err error
	)

	if res.Address, err = helperSecret.LoadValidatorAddress(ip.secretsManager); err != nil {
		return nil, err
	}

	if res.NodeID, err = helperSecret.LoadNodeID(ip.secretsManager); err != nil {
		return nil, err
	}

	return res, nil
}

// Checks if a config path is provided in the initialization parameters.
func (ip *initParams) hasConfigPath() bool {
	return ip.configPath != ""
}

// Reads and parses the secrets manager configuration from the provided config path.
// It validates the configuration and sets it in the initParams if successful.
func (ip *initParams) parseConfig() error {
	secretsConfig, readErr := secrets.ReadConfig(ip.configPath)
	if readErr != nil {
		return errInvalidConfig
	}

	if !secrets.SupportedServiceManager(secretsConfig.Type) {
		return errUnsupportedType
	}

	ip.secretsConfig = secretsConfig

	return nil
}
