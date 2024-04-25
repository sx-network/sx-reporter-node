package server

import (
	"fmt"
	"os"
	"strings"

	"github.com/hashicorp/go-hclog"
	"github.com/sx-network/sx-reporter/infra/secrets"
	"github.com/sx-network/sx-reporter/reporter"
	"gopkg.in/yaml.v3"
)

// YAMLServerConfig represents the configuration of a server, typically loaded from a YAML file.
type YAMLServerConfig struct {
	// Specifies the path to the configuration file.
	ConfigPath string
	// Indicates whether to use JSON log format.
	JSONLogFormat bool `json:"json_log_format" yaml:"json_log_format"`
	// Specifies the directory for storing data.
	DataDir string `json:"data_dir" yaml:"data_dir"`
	// Specifies the path to the secrets configuration file.
	SecretsConfigPath string `json:"secrets_config" yaml:"secrets_config"`
	// Contains the configuration for the reporter.
	YAMLReporterConfig *YAMLReporterConfig `json:"reporter" yaml:"reporter"`
	// Contains the configuration for secrets management.
	SecretsConfig *secrets.SecretsManagerConfig
}

// YAMLReporterConfig represents the configuration of a reporter, typically part of the server configuration.
type YAMLReporterConfig struct {
	AMQPURI                string `json:"amqp_uri" yaml:"amqp_uri"`
	AMQPExchangeName       string `json:"amqp_exchange_name" yaml:"amqp_exchange_name"`
	AMQPQueueName          string `json:"amqp_queue_name" yaml:"amqp_queue_name"`
	VerifyOutcomeAPIURL    string `json:"verify_outcome_api_url" yaml:"verify_outcome_api_url"`
	OutcomeReporterAddress string `json:"outcome_reporter_address" yaml:"outcome_reporter_address"`
	SXNodeAddress          string `json:"sx_node_address" yaml:"sx_node_address"`
}

// Represents the configuration of the server.
type ServerConfig struct {
	JSONLogFormat        bool                          // Indicates whether to use JSON log format
	LogLevel             hclog.Level                   // Log level for the server logger
	Logger               hclog.Logger                  // Logger instance for the server
	ReporterConfig       *ReporterConfig               // Configuration for the reporter
	ReporterService      *reporter.ReporterService     // Reporter service instance
	SecretsManagerConfig *secrets.SecretsManagerConfig // Configuration for the secrets manager
	SecretsManager       secrets.SecretsManager        // Secrets manager instance
	DataDir              string                        // Directory for storing data
}

// Represents the configuration for the reporter service.
type ReporterConfig struct {
	DataFeedAMQPURI          string // URI for the AMQP connection
	DataFeedAMQPExchangeName string // Name of the AMQP exchange
	DataFeedAMQPQueueName    string // Name of the AMQP queue
	VerifyOutcomeURI         string // URI for verifying outcome
	OutcomeReporterAddress   string // Address of the outcome reporter
	SXNodeAddress            string // Address of the SX node
}

// Initializes the server configuration from a file path specified in YAMLServerConfig.ConfigPath.
// It reads and parses the config file into yamlServerConfig.
// Returns an error if there's an issue reading or parsing the file.
func (yamlServerConfig *YAMLServerConfig) InitServerConfigFromFile() error {
	config, err := ReadConfigFile(yamlServerConfig.ConfigPath)
	if err != nil {
		return err
	}

	*yamlServerConfig = *config

	return nil
}

// Reads and parses a configuration file specified by 'path'.
// It supports .yaml and .yml file formats.
// Returns a YAMLServerConfig object parsed from the file.
// Returns an error if there's an issue reading or parsing the file.
func ReadConfigFile(path string) (*YAMLServerConfig, error) {
	data, err := os.ReadFile(path)

	if err != nil {
		return nil, err
	}

	var unmarshalFunc func([]byte, interface{}) error

	switch {
	case strings.HasSuffix(path, ".yaml"), strings.HasSuffix(path, ".yml"):
		unmarshalFunc = yaml.Unmarshal
	default:
		return nil, fmt.Errorf("suffix of %s is neither hcl, json, yaml nor yml", path)
	}

	yamlServerConfig := &YAMLServerConfig{}

	if err := unmarshalFunc(data, &yamlServerConfig); err != nil {
		return nil, err
	}

	return yamlServerConfig, nil
}

// Generates a Config object from the yamlServerConfig's configuration data.
// It populates the Config object with parsed values from the raw configuration.
func (yamlServerConfig *YAMLServerConfig) GenerateConfig() *ServerConfig {
	return &ServerConfig{
		LogLevel:             hclog.LevelFromString("DEBUG"),
		JSONLogFormat:        yamlServerConfig.JSONLogFormat,
		DataDir:              yamlServerConfig.DataDir,
		SecretsManagerConfig: yamlServerConfig.SecretsConfig,
		ReporterConfig: &ReporterConfig{
			DataFeedAMQPURI:          yamlServerConfig.YAMLReporterConfig.AMQPURI,
			DataFeedAMQPExchangeName: yamlServerConfig.YAMLReporterConfig.AMQPExchangeName,
			DataFeedAMQPQueueName:    yamlServerConfig.YAMLReporterConfig.AMQPQueueName,
			VerifyOutcomeURI:         yamlServerConfig.YAMLReporterConfig.VerifyOutcomeAPIURL,
			OutcomeReporterAddress:   yamlServerConfig.YAMLReporterConfig.OutcomeReporterAddress,
			SXNodeAddress:            yamlServerConfig.YAMLReporterConfig.SXNodeAddress,
		},
	}
}

// Sets the JSON log format in the server's configuration.
// It updates the JSONLogFormat field in yamlServerConfig based on the provided 'jsonLogFormat' boolean.
func (yamlServerConfig *YAMLServerConfig) SetJSONLogFormat(jsonLogFormat bool) {
	yamlServerConfig.JSONLogFormat = jsonLogFormat
}

// Initializes the raw parameters of the YAMLServerConfig.
// It calls initSecretsConfig to initialize secrets configuration if SecretsConfigPath is set.
func (yamlServerConfig *YAMLServerConfig) InitRawParams() error {
	if err := yamlServerConfig.initSecretsConfig(); err != nil {
		return err
	}

	return nil
}

// Initializes the secrets configuration of the YAMLServerConfig.
// It reads the secrets configuration file specified by SecretsConfigPath and assigns it to SecretsConfig.
func (yamlServerConfig *YAMLServerConfig) initSecretsConfig() error {
	if !yamlServerConfig.isSecretsConfigPathSet() {
		return nil
	}

	var parseErr error

	if yamlServerConfig.SecretsConfig, parseErr = secrets.ReadConfig(
		yamlServerConfig.SecretsConfigPath,
	); parseErr != nil {
		return fmt.Errorf("unable to read secrets config file, %w", parseErr)
	}

	return nil
}

// Checks if the SecretsConfigPath is set.
func (yamlServerConfig *YAMLServerConfig) isSecretsConfigPathSet() bool {
	return yamlServerConfig.SecretsConfigPath != ""
}
