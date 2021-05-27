package aws

import (
	"context"
	"fmt"
	"testing"

	"github.com/ConsenSysQuorum/quorum-key-manager/pkg/errors"
	"github.com/ConsenSysQuorum/quorum-key-manager/pkg/log"
	"github.com/ConsenSysQuorum/quorum-key-manager/src/infra/aws/mocks"
	"github.com/ConsenSysQuorum/quorum-key-manager/src/store/entities"
	"github.com/ConsenSysQuorum/quorum-key-manager/src/store/entities/testutils"
	"github.com/ConsenSysQuorum/quorum-key-manager/src/store/secrets"
	"github.com/aws/aws-sdk-go/aws/awserr"
	"github.com/aws/aws-sdk-go/service/secretsmanager"
	"github.com/golang/mock/gomock"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/suite"
)

type awsSecretStoreTestSuite struct {
	suite.Suite
	mockVault   *mocks.MockSecretsManagerClient
	secretStore secrets.Store
}

func TestAwsSecretStore(t *testing.T) {
	s := new(awsSecretStoreTestSuite)
	suite.Run(t, s)
}

func (s *awsSecretStoreTestSuite) SetupTest() {
	ctrl := gomock.NewController(s.T())
	defer ctrl.Finish()

	s.mockVault = mocks.NewMockSecretsManagerClient(ctrl)

	s.secretStore = New(s.mockVault, log.DefaultLogger())
}

func (s *awsSecretStoreTestSuite) TestSet() {
	ctx := context.Background()
	id := "my-secret1"
	version := "2"
	value := "my-value1"
	attributes := testutils.FakeAttributes()

	createOutput := &secretsmanager.CreateSecretOutput{
		Name:      &id,
		VersionId: &version,
	}

	currentMark := CurrentVersionMark
	versionID2stages := map[string][]*string{
		version: {&currentMark},
	}

	descSecretOutput := &secretsmanager.DescribeSecretOutput{
		Name:               &id,
		VersionIdsToStages: versionID2stages,
		Tags:               ToSecretmanagerTags(attributes.Tags),
	}

	s.T().Run("should set a new secret successfully", func(t *testing.T) {
		s.mockVault.EXPECT().CreateSecret(gomock.Any(), id, value).Return(createOutput, nil)
		s.mockVault.EXPECT().TagSecretResource(gomock.Any(), id, attributes.Tags).Return(&secretsmanager.TagResourceOutput{}, nil)
		s.mockVault.EXPECT().DescribeSecret(gomock.Any(), id).Return(descSecretOutput, nil)

		secret, err := s.secretStore.Set(ctx, id, value, attributes)

		assert.NoError(t, err)
		assert.Equal(t, value, secret.Value)

		assert.ObjectsAreEqual(attributes.Tags, secret.Tags)
		assert.Equal(t, version, secret.Metadata.Version)
		assert.False(t, secret.Metadata.Disabled)
		assert.True(t, secret.Metadata.ExpireAt.IsZero())
		assert.True(t, secret.Metadata.DeletedAt.IsZero())
	})

	s.T().Run("should fail when too many tags", func(t *testing.T) {
		tooManyTags := map[string]string{}

		for i := 0; i <= maxTagsAllowed; i++ {
			tooManyTags[fmt.Sprintf("tag%d", i)] = fmt.Sprintf("value%d", i)
		}
		attributes.Tags = tooManyTags
		s.mockVault.EXPECT().CreateSecret(gomock.Any(), id, value).Return(createOutput, nil)
		secret, err := s.secretStore.Set(ctx, id, value, attributes)

		// tags back to normal
		attributes.Tags = testutils.FakeTags()

		assert.NotNil(t, err)
		assert.True(t, errors.IsInvalidParameterError(err))
		assert.Nil(t, secret)
	})

	s.T().Run("should fail with describe error", func(t *testing.T) {
		expectedErr := fmt.Errorf("any error")
		s.mockVault.EXPECT().CreateSecret(gomock.Any(), id, value).Return(createOutput, nil)
		s.mockVault.EXPECT().TagSecretResource(gomock.Any(), id, attributes.Tags).Return(&secretsmanager.TagResourceOutput{}, nil)
		s.mockVault.EXPECT().DescribeSecret(gomock.Any(), id).Return(descSecretOutput, expectedErr)

		secret, err := s.secretStore.Set(ctx, id, value, attributes)

		assert.Equal(t, err, expectedErr)
		assert.Nil(t, secret)

	})

	s.T().Run("should fail with tag error", func(t *testing.T) {
		expectedErr := fmt.Errorf("any error")
		s.mockVault.EXPECT().CreateSecret(gomock.Any(), id, value).Return(createOutput, nil)
		s.mockVault.EXPECT().TagSecretResource(gomock.Any(), id, attributes.Tags).Return(nil, expectedErr)

		secret, err := s.secretStore.Set(ctx, id, value, attributes)

		assert.Equal(t, err, expectedErr)
		assert.Nil(t, secret)

	})

	s.T().Run("should fail with same error if write fails", func(t *testing.T) {
		expectedErr := fmt.Errorf("error")
		s.mockVault.EXPECT().CreateSecret(gomock.Any(), id, value).Return(&secretsmanager.CreateSecretOutput{}, expectedErr)
		s.mockVault.EXPECT().TagSecretResource(gomock.Any(), id, attributes.Tags).Return(&secretsmanager.TagResourceOutput{}, nil)
		s.mockVault.EXPECT().DescribeSecret(gomock.Any(), id).Return(descSecretOutput, nil)

		secret, err := s.secretStore.Set(ctx, id, value, attributes)

		assert.Nil(t, secret)
		assert.Equal(t, expectedErr, err)
	})

	s.T().Run("should update secret if already exists", func(t *testing.T) {

		s.mockVault.EXPECT().CreateSecret(gomock.Any(), id, value).Return(&secretsmanager.CreateSecretOutput{}, awserr.New(secretsmanager.ErrCodeResourceExistsException, "", nil))
		s.mockVault.EXPECT().PutSecretValue(gomock.Any(), id, value).Return(&secretsmanager.PutSecretValueOutput{}, nil)
		s.mockVault.EXPECT().TagSecretResource(gomock.Any(), id, attributes.Tags).Return(&secretsmanager.TagResourceOutput{}, nil)
		s.mockVault.EXPECT().DescribeSecret(gomock.Any(), id).Return(descSecretOutput, nil)

		secret, err := s.secretStore.Set(ctx, id, value, attributes)

		assert.NoError(t, err)
		assert.Equal(t, value, secret.Value)

		assert.ObjectsAreEqual(attributes.Tags, secret.Tags)
		assert.Equal(t, version, secret.Metadata.Version)
	})

}

func (s *awsSecretStoreTestSuite) TestGet() {
	ctx := context.Background()
	id := "my-secret-get"
	version := "some-version"
	secretValue := "secret-value"

	expectedSecret := &entities.Secret{
		ID:    id,
		Value: secretValue,
	}

	getSecretOutput := &secretsmanager.GetSecretValueOutput{
		Name:         &id,
		SecretString: &secretValue,
		VersionId:    &version,
	}

	currentMark := CurrentVersionMark
	versionID2stages := map[string][]*string{
		version: {&currentMark},
	}

	descSecretOutput := &secretsmanager.DescribeSecretOutput{
		Name:               &id,
		VersionIdsToStages: versionID2stages,
	}

	s.T().Run("should get a secret successfully", func(t *testing.T) {
		s.mockVault.EXPECT().GetSecret(gomock.Any(), id, "").Return(getSecretOutput, nil)
		s.mockVault.EXPECT().DescribeSecret(gomock.Any(), id).Return(descSecretOutput, nil)
		retValue, err := s.secretStore.Get(ctx, id, "")
		assert.NoError(t, err)
		assert.Equal(t, retValue.Value, expectedSecret.Value)
		assert.Equal(t, retValue.ID, expectedSecret.ID)
	})

	s.T().Run("should fail with get error", func(t *testing.T) {
		expectedErr := errors.NotFoundError("secret not found")
		s.mockVault.EXPECT().GetSecret(gomock.Any(), id, version).Return(getSecretOutput, expectedErr)

		retValue, err := s.secretStore.Get(ctx, id, version)
		assert.Nil(t, retValue)
		assert.Equal(t, err, expectedErr)
	})

	s.T().Run("should fail with describe error", func(t *testing.T) {
		expectedErr := errors.NotFoundError("secret not found")
		s.mockVault.EXPECT().GetSecret(gomock.Any(), id, version).Return(getSecretOutput, nil)
		s.mockVault.EXPECT().DescribeSecret(gomock.Any(), id).Return(descSecretOutput, expectedErr)
		retValue, err := s.secretStore.Get(ctx, id, version)
		assert.Nil(t, retValue)
		assert.Equal(t, err, expectedErr)
	})
}

func (s *awsSecretStoreTestSuite) TestGetDeleted() {
	s.T().Run("should fail with not implemented error", func(t *testing.T) {
		ctx := context.Background()
		id := "some-id"
		expectedError := errors.ErrNotImplemented
		_, err := s.secretStore.GetDeleted(ctx, id)

		assert.Equal(t, err, expectedError)
	})
}

func (s *awsSecretStoreTestSuite) TestDeleted() {
	ctx := context.Background()
	id := "my-secret-deleted"
	destroy := false

	deleteOutput := &secretsmanager.DeleteSecretOutput{
		Name: &id,
	}

	s.T().Run("should delete secret successfully", func(t *testing.T) {
		s.mockVault.EXPECT().DeleteSecret(gomock.Any(), id, destroy).Return(deleteOutput, nil)

		_, err := s.secretStore.Delete(ctx, id)

		assert.NoError(t, err)
	})
}

func (s *awsSecretStoreTestSuite) TestDestroy() {
	ctx := context.Background()
	id := "my-secret"
	destroy := true

	deleteOutput := &secretsmanager.DeleteSecretOutput{
		Name: &id,
	}

	s.T().Run("should destroy secret successfully", func(t *testing.T) {
		s.mockVault.EXPECT().DeleteSecret(gomock.Any(), id, destroy).Return(deleteOutput, nil)

		err := s.secretStore.Destroy(ctx, id)

		assert.NoError(t, err)
	})

	s.T().Run("should fail to destroy secret with internal error", func(t *testing.T) {

		awsError := awserr.New(secretsmanager.ErrCodeInternalServiceError, "", nil)
		expectedError := errors.InternalError("internal error")
		s.mockVault.EXPECT().DeleteSecret(gomock.Any(), id, destroy).Return(nil, awsError)

		err := s.secretStore.Destroy(ctx, id)

		assert.Error(t, err)
		assert.Equal(t, err, expectedError)
	})

	s.T().Run("should fail to destroy secret with internal error", func(t *testing.T) {

		awsError := awserr.New(secretsmanager.ErrCodeInternalServiceError, "", nil)
		expectedError := errors.InternalError("internal error")
		s.mockVault.EXPECT().DeleteSecret(gomock.Any(), id, destroy).Return(nil, awsError)

		err := s.secretStore.Destroy(ctx, id)

		assert.Error(t, err)
		assert.Equal(t, err, expectedError)
	})

	s.T().Run("should fail to destroy secret with not found error", func(t *testing.T) {

		awsError := awserr.New(secretsmanager.ErrCodeResourceNotFoundException, "", nil)
		expectedError := errors.NotFoundError("resource was not found")
		s.mockVault.EXPECT().DeleteSecret(gomock.Any(), id, destroy).Return(nil, awsError)

		err := s.secretStore.Destroy(ctx, id)

		assert.Error(t, err)
		assert.Equal(t, err, expectedError)
	})

	s.T().Run("should fail to destroy secret with not found error", func(t *testing.T) {

		awsError := awserr.New(secretsmanager.ErrCodeLimitExceededException, "", nil)
		expectedError := errors.InternalError("internal error, limit exceeded")
		s.mockVault.EXPECT().DeleteSecret(gomock.Any(), id, destroy).Return(nil, awsError)

		err := s.secretStore.Destroy(ctx, id)

		assert.Error(t, err)
		assert.Equal(t, err, expectedError)
	})

	s.T().Run("should fail to destroy secret with invalid parameter error", func(t *testing.T) {

		awsError := awserr.New(secretsmanager.ErrCodeInvalidParameterException, "", nil)
		expectedError := errors.InvalidParameterError("invalid parameter")
		s.mockVault.EXPECT().DeleteSecret(gomock.Any(), id, destroy).Return(nil, awsError)

		err := s.secretStore.Destroy(ctx, id)

		assert.Error(t, err)
		assert.Equal(t, err, expectedError)
	})

	s.T().Run("should fail to destroy secret with invalid request error", func(t *testing.T) {

		awsError := awserr.New(secretsmanager.ErrCodeInvalidRequestException, "", nil)
		expectedError := errors.InvalidRequestError("invalid request")
		s.mockVault.EXPECT().DeleteSecret(gomock.Any(), id, destroy).Return(nil, awsError)

		err := s.secretStore.Destroy(ctx, id)

		assert.Error(t, err)
		assert.Equal(t, err, expectedError)
	})
}

func (s *awsSecretStoreTestSuite) TestList() {
	ctx := context.Background()
	sec3, sec4 := "my-secret3", "my-secret4"
	expected := []string{sec3, sec4}
	secretsList := []*secretsmanager.SecretListEntry{{Name: &sec3}, {Name: &sec4}}

	s.T().Run("should list all secret ids successfully", func(t *testing.T) {

		listOutput := &secretsmanager.ListSecretsOutput{
			SecretList: secretsList,
		}

		s.mockVault.EXPECT().ListSecrets(gomock.Any(), int64(0), "").Return(listOutput, nil)
		ids, err := s.secretStore.List(ctx)

		assert.NoError(t, err)
		assert.Equal(t, expected, ids)
	})

	s.T().Run("should list all secret ids successfully with a nextToken", func(t *testing.T) {

		nextToken := "next"
		listOutput := &secretsmanager.ListSecretsOutput{
			SecretList: secretsList,
			NextToken:  &nextToken,
		}

		s.mockVault.EXPECT().ListSecrets(gomock.Any(), int64(0), "").Return(listOutput, nil)
		listOutput.NextToken = nil
		s.mockVault.EXPECT().ListSecrets(gomock.Any(), int64(0), nextToken).Return(listOutput, nil)
		ids, err := s.secretStore.List(ctx)

		assert.NoError(t, err)
		assert.Equal(t, expected, ids)
	})

	s.T().Run("should return empty list if result is nil", func(t *testing.T) {
		s.mockVault.EXPECT().ListSecrets(gomock.Any(), int64(0), "").Return(&secretsmanager.ListSecretsOutput{}, nil)
		ids, err := s.secretStore.List(ctx)

		assert.NoError(t, err)
		assert.Empty(t, ids)
	})

	s.T().Run("should fail if list fails", func(t *testing.T) {
		expectedErr := fmt.Errorf("error")

		s.mockVault.EXPECT().ListSecrets(gomock.Any(), int64(0), "").Return(&secretsmanager.ListSecretsOutput{}, expectedErr)
		ids, err := s.secretStore.List(ctx)

		assert.Nil(t, ids)
		assert.Equal(t, expectedErr, err)
	})
}

func (s *awsSecretStoreTestSuite) TestListDeleted() {
	s.T().Run("should fail with not implemented error", func(t *testing.T) {
		ctx := context.Background()
		expectedError := errors.ErrNotImplemented
		_, err := s.secretStore.ListDeleted(ctx)

		assert.Equal(t, err, expectedError)
	})
}

func ToSecretmanagerTags(tags map[string]string) []*secretsmanager.Tag {
	var fakeSecretsTags []*secretsmanager.Tag

	for key, value := range tags {
		k, v := key, value
		var in = secretsmanager.Tag{
			Key:   &k,
			Value: &v,
		}
		fakeSecretsTags = append(fakeSecretsTags, &in)
	}
	return fakeSecretsTags
}
