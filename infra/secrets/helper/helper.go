package helper

import (
	"errors"
	"fmt"

	"github.com/hashicorp/go-hclog"
	"github.com/sx-network/sx-reporter/command/crypto"
	"github.com/sx-network/sx-reporter/helper/types"
	"github.com/sx-network/sx-reporter/infra/secrets"
	"github.com/sx-network/sx-reporter/infra/secrets/awsssm"
	"github.com/sx-network/sx-reporter/infra/secrets/local"
)

// SetupLocalSecretsManager is a helper method for boilerplate local secrets manager setup
func SetupLocalSecretsManager(dataDir string) (secrets.SecretsManager, error) {
	return local.SecretsManagerFactory(
		nil, // Local secrets manager doesn't require a config
		&secrets.SecretsManagerParams{
			Logger: hclog.NewNullLogger(),
			Extra: map[string]interface{}{
				secrets.Path: dataDir,
			},
		},
	)
}

// InitECDSAValidatorKey creates new ECDSA key and set as a validator key
func InitECDSAValidatorKey(secretsManager secrets.SecretsManager) (types.Address, error) {
	if secretsManager.HasSecret(secrets.ReporterKey) {
		return types.ZeroAddress, fmt.Errorf(`secrets "%s" has been already initialized`, secrets.ReporterKey)
	}

	validatorKey, validatorKeyEncoded, err := crypto.GenerateAndEncodeECDSAPrivateKey()
	if err != nil {
		return types.ZeroAddress, err
	}

	address := crypto.PubKeyToAddress(&validatorKey.PublicKey)

	// Write the validator private key to the secrets manager storage
	if setErr := secretsManager.SetSecret(
		secrets.ReporterKey,
		validatorKeyEncoded,
	); setErr != nil {
		return types.ZeroAddress, setErr
	}

	return address, nil
}

// LoadValidatorAddress loads ECDSA key by SecretsManager and returns reporter address
func LoadValidatorAddress(secretsManager secrets.SecretsManager) (types.Address, error) {
	if !secretsManager.HasSecret(secrets.ReporterKey) {
		return types.ZeroAddress, nil
	}

	encodedKey, err := secretsManager.GetSecret(secrets.ReporterKey)
	if err != nil {
		return types.ZeroAddress, err
	}

	privateKey, err := crypto.BytesToECDSAPrivateKey(encodedKey)
	if err != nil {
		return types.ZeroAddress, err
	}

	return crypto.PubKeyToAddress(&privateKey.PublicKey), nil
}

// LoadNodeID loads Libp2p key by SecretsManager and returns Node ID
func LoadNodeID(secretsManager secrets.SecretsManager) (string, error) {
	// encodedKey, err := secretsManager.GetSecret(secrets.ReporterKey)
	// if err != nil {
	// 	return "", err
	// }

	// parsedKey, err := keystore.ParseLibp2pKey(encodedKey)
	// if err != nil {
	// 	return "", err
	// }

	// nodeID, err := peer.IDFromPrivateKey(parsedKey)
	// if err != nil {
	// 	return "", err
	// }

	// return nodeID.String(), nil

	return "", nil
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
