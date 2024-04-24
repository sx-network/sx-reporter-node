package config

import (
	"github.com/sx-network/sx-reporter/infra/secrets"
	"github.com/sx-network/sx-reporter/infra/secrets/awsssm"
	"github.com/sx-network/sx-reporter/infra/secrets/local"
)

// It is a map that defines the SecretManager factories for different secret management solutions.
// It maps each SecretsManagerType to its corresponding factory function.
var SecretsManagerBackends = map[secrets.SecretsManagerType]secrets.SecretsManagerFactory{
	secrets.Local:  local.SecretsManagerFactory,
	secrets.AWSSSM: awsssm.SecretsManagerFactory,
}
