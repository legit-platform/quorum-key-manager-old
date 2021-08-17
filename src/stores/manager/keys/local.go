package keys

import (
	"github.com/consensys/quorum-key-manager/src/infra/log"
	"github.com/consensys/quorum-key-manager/src/stores"

	"github.com/consensys/quorum-key-manager/pkg/errors"
	localkeys "github.com/consensys/quorum-key-manager/src/stores/store/keys/local"

	manifest "github.com/consensys/quorum-key-manager/src/manifests/types"
	msecrets "github.com/consensys/quorum-key-manager/src/stores/manager/secrets"
)

type LocalKeySpecs struct {
	SecretStore manifest.Kind
	Specs       interface{}
}

func NewLocalKeyStore(specs *LocalKeySpecs, logger log.Logger) (*localkeys.Store, error) {
	var secretStore stores.SecretStore
	var err error

	switch specs.SecretStore {
	case stores.HashicorpSecrets:
		spec := &msecrets.HashicorpSecretSpecs{}
		if err = manifest.UnmarshalSpecs(specs.Specs, spec); err != nil {
			errMessage := "failed to unmarshal Hashicorp secret store specs"
			logger.WithError(err).Error(errMessage)
			return nil, errors.InvalidFormatError(errMessage)
		}
		secretStore, err = msecrets.NewHashicorpSecretStore(spec, logger)
	case stores.AKVSecrets:
		spec := &msecrets.AkvSecretSpecs{}
		if err = manifest.UnmarshalSpecs(specs.Specs, spec); err != nil {
			errMessage := "failed to unmarshal AKV secret store specs"
			logger.WithError(err).Error(errMessage)
			return nil, errors.InvalidFormatError(errMessage)
		}
		secretStore, err = msecrets.NewAkvSecretStore(spec, logger)
	case stores.AWSSecrets:
		spec := &msecrets.AwsSecretSpecs{}
		if err = manifest.UnmarshalSpecs(specs.Specs, spec); err != nil {
			errMessage := "failed to unmarshal AWS secret store specs"
			logger.WithError(err).Error(errMessage)
			return nil, errors.InvalidFormatError(errMessage)
		}
		secretStore, err = msecrets.NewAwsSecretStore(spec, logger)
	default:
		errMessage := "invalid secret store kind"
		logger.Error(errMessage, "kind", specs.SecretStore)
		return nil, errors.InvalidFormatError(errMessage)
	}
	if err != nil {
		return nil, err
	}

	return localkeys.New(secretStore, logger), nil
}