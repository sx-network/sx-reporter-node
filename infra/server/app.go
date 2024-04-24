package server

import (
	"fmt"

	"github.com/hashicorp/go-hclog"
	"github.com/sx-network/sx-reporter/infra/secrets"
	"github.com/sx-network/sx-reporter/infra/secrets/config"
	"github.com/sx-network/sx-reporter/reporter"
)

func NewServer(serverConfig *ServerConfig) (*ServerConfig, error) {
	logger, err := newLogger(serverConfig)
	if err != nil {
		return nil, fmt.Errorf("could not setup new logger instance, %w", err)
	}
	serverConfig.Logger = logger

	serverConfig.Logger.Info("Data dir", "path", serverConfig.DataDir)
	serverConfig.Logger.Info("working")

	if err := serverConfig.setupSecretsManager(); err != nil {
		return nil, fmt.Errorf("failed to set up the secrets manager: %w", err)
	}

	// setup and start reporter consumer
	if err := serverConfig.setupReporterService(); err != nil {
		return nil, err
	}

	return serverConfig, nil
}

// Creates a new logger which logs to a specified file.
// If log file is not set it outputs to standard output ( console ).
// If log file is specified, and it can't be created the server command will error out
func newLogger(serverConfig *ServerConfig) (hclog.Logger, error) {
	return hclog.New(&hclog.LoggerOptions{
		Name:       "sx-network-log",
		Level:      serverConfig.LogLevel,
		JSONFormat: serverConfig.JSONLogFormat,
	}), nil
}

// Sets up the secrets manager based on the provided configuration.
// If no configuration is provided, it defaults to using the local secrets manager.
// The function then retrieves the appropriate factory method based on the secrets manager type.
// It instantiates the secrets manager using the factory method and assigns it to serverConfig.SecretsManager.
func (serverConfig *ServerConfig) setupSecretsManager() error {
	secretsManagerConfig := serverConfig.SecretsManagerConfig

	if secretsManagerConfig == nil {
		secretsManagerConfig = &secrets.SecretsManagerConfig{
			Type: secrets.Local,
		}
	}

	secretsManagerType := secretsManagerConfig.Type
	secretsManagerParams := &secrets.SecretsManagerParams{
		Logger: serverConfig.Logger,
	}

	if secretsManagerType == secrets.Local {
		// Only the base directory is required for the local secrets manager
		secretsManagerParams.Extra = map[string]interface{}{
			secrets.Path: serverConfig.DataDir,
		}
	}

	secretsManagerFactory, ok := config.SecretsManagerBackends[secretsManagerType]
	if !ok {
		return fmt.Errorf("secrets manager type '%s' not found", secretsManagerType)
	}

	secretsManager, factoryErr := secretsManagerFactory(
		secretsManagerConfig,
		secretsManagerParams,
	)

	if factoryErr != nil {
		return fmt.Errorf("unable to instantiate secrets manager, %w", factoryErr)
	}

	serverConfig.SecretsManager = secretsManager

	return nil
}

// setupDataFeedService set up and start datafeed service
func (serverConfig *ServerConfig) setupReporterService() error {
	conf := &reporter.ReporterConfig{
		MQConfig: &reporter.MQConfig{
			AMQPURI:      serverConfig.ReporterConfig.DataFeedAMQPURI,
			ExchangeName: serverConfig.ReporterConfig.DataFeedAMQPExchangeName,
			QueueConfig: &reporter.QueueConfig{
				QueueName: serverConfig.ReporterConfig.DataFeedAMQPQueueName,
			},
		},
		VerifyOutcomeURI:       serverConfig.ReporterConfig.VerifyOutcomeURI,
		OutcomeReporterAddress: serverConfig.ReporterConfig.OutcomeReporterAddress,
		SXNodeAddress:          serverConfig.ReporterConfig.SXNodeAddress,
	}

	reporterService, err := reporter.NewReporterService(
		serverConfig.Logger,
		conf,
		// s.grpcServer,
		&serverConfig.SecretsManager,
	)
	if err != nil {
		return err
	}

	serverConfig.ReporterService = reporterService

	return nil
}

// Closes the server (reporter...)
func (serverConfig *ServerConfig) Close() {
	// close the txpool's main loop
	// serverConfig.txpool.Close()
}
