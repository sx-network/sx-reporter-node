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

type RawServerConfig struct {
	RawConfig  *RawConfig
	ConfigPath string
}

type RawConfig struct {
	JSONLogFormat bool         `json:"json_log_format" yaml:"json_log_format"`
	DataDir       string       `json:"data_dir" yaml:"data_dir"`
	RawReporter   *RawReporter `json:"reporter" yaml:"reporter"`
}

type RawReporter struct {
	AMQPURI                string `json:"amqp_uri" yaml:"amqp_uri"`
	AMQPExchangeName       string `json:"amqp_exchange_name" yaml:"amqp_exchange_name"`
	AMQPQueueName          string `json:"amqp_queue_name" yaml:"amqp_queue_name"`
	VerifyOutcomeAPIURL    string `json:"verify_outcome_api_url" yaml:"verify_outcome_api_url"`
	OutcomeReporterAddress string `json:"outcome_reporter_address" yaml:"outcome_reporter_address"`
	SXNodeAddress          string `json:"sx_node_address" yaml:"sx_node_address"`
}

type ServerConfig struct {
	JSONLogFormat        bool
	LogLevel             hclog.Level
	Logger               hclog.Logger
	ReporterConfig       *ReporterConfig
	ReporterService      *reporter.ReporterService
	SecretsManagerConfig *secrets.SecretsManagerConfig
	SecretsManager       secrets.SecretsManager
	DataDir              string
}

type ReporterConfig struct {
	DataFeedAMQPURI          string
	DataFeedAMQPExchangeName string
	DataFeedAMQPQueueName    string
	VerifyOutcomeURI         string
	OutcomeReporterAddress   string
	SXNodeAddress            string
}

// Returns a default RawConfig object with placeholder default values.
func DefaultConfig() *RawConfig {
	return &RawConfig{
		JSONLogFormat: false,
		DataDir:       "",
		RawReporter: &RawReporter{
			AMQPURI:                "",
			AMQPExchangeName:       "",
			AMQPQueueName:          "",
			VerifyOutcomeAPIURL:    "",
			OutcomeReporterAddress: "",
			SXNodeAddress:          "",
		},
	}
}

// Initializes the server configuration from a file path specified in serverParams.ConfigPath.
// It reads and parses the config file into serverParams.RawConfig.
// Returns an error if there's an issue reading or parsing the file.
func (serverParams *RawServerConfig) InitServerConfigFromFile() error {
	var parseErr error

	if serverParams.RawConfig, parseErr = ReadConfigFile(serverParams.ConfigPath); parseErr != nil {
		return parseErr
	}

	return nil
}

// Reads and parses a configuration file specified by 'path'.
// It supports .yaml and .yml file formats.
// Returns a RawConfig object parsed from the file.
// Returns an error if there's an issue reading or parsing the file.
func ReadConfigFile(path string) (*RawConfig, error) {
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

	config := DefaultConfig()

	if err := unmarshalFunc(data, config); err != nil {
		return nil, err
	}

	return config, nil
}

// Sets the JSON log format in the server's configuration.
// It updates the JSONLogFormat field in p.RawConfig based on the provided 'jsonLogFormat' boolean.
func (serverParams *RawServerConfig) SetJSONLogFormat(jsonLogFormat bool) {
	serverParams.RawConfig.JSONLogFormat = jsonLogFormat
}

// Generates a Config object from the serverParams's raw configuration data (RawConfig).
// It populates the Config object with parsed values from the raw configuration.
func (serverParams *RawServerConfig) GenerateConfig() *ServerConfig {
	return &ServerConfig{
		JSONLogFormat: serverParams.RawConfig.JSONLogFormat,
		DataDir:       serverParams.RawConfig.DataDir,
		ReporterConfig: &ReporterConfig{
			DataFeedAMQPURI:          serverParams.RawConfig.RawReporter.AMQPURI,
			DataFeedAMQPExchangeName: serverParams.RawConfig.RawReporter.AMQPExchangeName,
			DataFeedAMQPQueueName:    serverParams.RawConfig.RawReporter.AMQPQueueName,
			VerifyOutcomeURI:         serverParams.RawConfig.RawReporter.VerifyOutcomeAPIURL,
			OutcomeReporterAddress:   serverParams.RawConfig.RawReporter.OutcomeReporterAddress,
			SXNodeAddress:            serverParams.RawConfig.RawReporter.SXNodeAddress,
		},
	}
}
