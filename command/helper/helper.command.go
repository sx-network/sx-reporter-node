package helper

import (
	"errors"
	"fmt"
	"time"

	"github.com/hashicorp/go-hclog"
	"github.com/ryanuber/columnize"
	"github.com/sx-network/sx-reporter/command"
	"github.com/sx-network/sx-reporter/infra/common"
	"github.com/sx-network/sx-reporter/infra/secrets"
	"github.com/sx-network/sx-reporter/infra/secrets/awsssm"
)

type ClientCloseResult struct {
	Message string `json:"message"`
}

func (r *ClientCloseResult) GetOutput() string {
	return r.Message
}

// HandleSignals is a helper method for handling signals sent to the console
// Like stop, error, etc.
func HandleSignals(
	closeFn func(),
	outputter command.OutputFormatter,
) error {
	signalCh := common.GetTerminationSignalCh()
	sig := <-signalCh

	closeMessage := fmt.Sprintf("\n[SIGNAL] Caught signal: %v\n", sig)
	closeMessage += "Gracefully shutting down client...\n"

	outputter.SetCommandResult(
		&ClientCloseResult{
			Message: closeMessage,
		},
	)
	outputter.WriteOutput()

	// Call the Minimal server close callback
	gracefulCh := make(chan struct{})

	go func() {
		if closeFn != nil {
			closeFn()
		}

		close(gracefulCh)
	}()

	select {
	case <-signalCh:
		return errors.New("shutdown by signal channel")
	case <-time.After(5 * time.Second):
		return errors.New("shutdown by timeout")
	case <-gracefulCh:
		return nil
	}
}

// GetCloudSecretsManager returns the cloud secrets manager from the provided config
func InitCloudSecretsManager(secretsConfig *secrets.SecretsManagerConfig) (secrets.SecretsManager, error) {
	var secretsManager secrets.SecretsManager

	switch secretsConfig.Type {
	case secrets.AWSSSM:
		AWSSSM, err := setupAWSSSM(secretsConfig)
		if err != nil {
			return secretsManager, err
		}

		secretsManager = AWSSSM
	default:
		return secretsManager, errors.New("unsupported secrets manager")
	}

	return secretsManager, nil
}

// setupAWSSSM is a helper method for boilerplate aws ssm secrets manager setup
func setupAWSSSM(
	secretsConfig *secrets.SecretsManagerConfig,
) (secrets.SecretsManager, error) {
	return awsssm.SecretsManagerFactory(
		secretsConfig,
		&secrets.SecretsManagerParams{
			Logger: hclog.NewNullLogger(),
		},
	)
}

// FormatKV formats key value pairs:
//
// Key = Value
//
// Key = <none>
func FormatKV(in []string) string {
	columnConf := columnize.DefaultConfig()
	columnConf.Empty = "<none>"
	columnConf.Glue = " = "

	return columnize.Format(in, columnConf)
}
