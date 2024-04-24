package secrets

import (
	"errors"

	"github.com/hashicorp/go-hclog"
)

// Represents the configuration for a secrets manager.
type SecretsManagerConfig struct {
	Token     string                 `json:"token"`      // Access token to the instance
	ServerURL string                 `json:"server_url"` // The URL of the running server
	Type      SecretsManagerType     `json:"type"`       // The type of SecretsManager
	Name      string                 `json:"name"`       // The name of the current node
	Namespace string                 `json:"namespace"`  // The namespace of the service
	Extra     map[string]interface{} `json:"extra"`      // Any kind of arbitrary data
}

// Represents the type of secrets manager.
type SecretsManagerType string

// Defines the configuration parameters for the secrets manager.
type SecretsManagerParams struct {
	// Local logger object
	Logger hclog.Logger
	// Extra contains additional data needed for the SecretsManager to function
	Extra map[string]interface{}
}

// Constants representing different types of secrets managers.
const (
	// Local pertains to the local FS [Default]
	Local SecretsManagerType = "local"
	// HashicorpVault pertains to the Hashicorp Vault server
	HashicorpVault SecretsManagerType = "hashicorp-vault"
	// AWSSSM pertains to AWS SSM using configured EC2 instance role
	AWSSSM SecretsManagerType = "aws-ssm"
	// GCPSSM pertains to the Google Cloud Computing secret store manager
	GCPSSM SecretsManagerType = "gcp-ssm"
)

// Constants representing keys used in the SecretsManagerParams Extra map for configuration.
const (
	// Path is the path to the base working directory
	Path = "path"
	// Token is the token used for authenticating with a KMS
	Token = "token"
	// Server is the address of the KMS
	Server = "server"
	// Name is the name of the current node
	Name = "name"
)

// It is the factory method for creating secrets managers.
// It takes in the necessary configuration and runtime parameters to instantiate a SecretsManager.
// The function returns a SecretsManager instance and an error if any.
type SecretsManagerFactory func(
	// The `config` parameter contains the necessary configuration saved to/read from JSON,
	// used to configure the SecretsManager with information saved in advance.
	config *SecretsManagerConfig,
	// The `params` parameter contains the runtime configuration parameters, such as the logger used,
	// as well as any additional data the secrets manager might need (SecretsManagerParams.Extra field).
	params *SecretsManagerParams,
) (SecretsManager, error)

// Defines the base public interface that all secret manager implementations should have.
type SecretsManager interface {
	// Setup performs secret manager-specific setup.
	// It initializes the secrets manager and prepares it for use.
	Setup() error
	// GetSecret retrieves the secret by its name.
	// It returns the secret value as a byte slice and an error if any.
	GetSecret(name string) ([]byte, error)
	// SetSecret sets the secret to the provided value.
	// It stores the secret with the given name and value.
	SetSecret(name string, value []byte) error
	// HasSecret checks if the secret with the given name is present.
	// It returns true if the secret exists, false otherwise.
	HasSecret(name string) bool
	// RemoveSecret removes the secret with the given name from storage.
	// It deletes the secret from the secrets manager.
	RemoveSecret(name string) error
}

// Constants representing folder names.
const (
	ConsensusFolderLocal = "consensus"
	NetworkFolderLocal   = "libp2p"
)

// Constants representing names for available secrets.
const (
	// ValidatorKey is the private key secret of the validator node
	ValidatorKey = "validator-key"
	// ValidatorBLSKey is the bls secret key of the validator node
	ValidatorBLSKey = "validator-bls-key"
	// NetworkKey is the libp2p private key secret used for networking
	NetworkKey = "network-key"
)

// Constants representing file names for the local StorageManager.
const (
	// It is the file name for the validator node's private key in the local StorageManager.
	ValidatorKeyLocal = "validator.key"
	// It is the file name for the validator node's BLS secret key in the local StorageManager.
	ValidatorBLSKeyLocal = "validator-bls.key"
	// It is the file name for the libp2p private key used for networking in the local StorageManager.
	NetworkKeyLocal = "libp2p.key"
)

// It is an error indicating that a secret was not found.
var ErrSecretNotFound = errors.New("secret not found")
