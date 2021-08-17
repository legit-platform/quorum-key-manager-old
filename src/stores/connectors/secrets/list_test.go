package secrets

import (
	"context"
	"testing"

	"github.com/consensys/quorum-key-manager/pkg/errors"
	"github.com/consensys/quorum-key-manager/src/infra/log/testutils"
	mock2 "github.com/consensys/quorum-key-manager/src/stores/database/mock"
	"github.com/consensys/quorum-key-manager/src/stores/entities"
	testutils2 "github.com/consensys/quorum-key-manager/src/stores/entities/testutils"
	"github.com/consensys/quorum-key-manager/src/stores/mock"
	"github.com/golang/mock/gomock"
	"github.com/stretchr/testify/assert"
)

func TestListSecret(t *testing.T) {
	ctx := context.Background()
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	store := mock.NewMockSecretStore(ctrl)
	db := mock2.NewMockSecrets(ctrl)
	logger := testutils.NewMockLogger(ctrl)

	connector := NewConnector(store, db, nil, logger)

	t.Run("should list deleted secret successfully", func(t *testing.T) {
		secretOne := testutils2.FakeSecret()
		secretTwo := testutils2.FakeSecret()

		db.EXPECT().GetAll(gomock.Any()).Return([]*entities.Secret{secretOne, secretTwo}, nil)

		secretIDs, err := connector.List(ctx)

		assert.NoError(t, err)
		assert.Equal(t, secretIDs, []string{secretOne.ID, secretTwo.ID})
	})

	t.Run("should fail to list deleted secret if db fails", func(t *testing.T) {
		expectedErr := errors.PostgresError("cannot connect")

		db.EXPECT().GetAll(gomock.Any()).Return(nil, expectedErr)

		_, err := connector.List(ctx)

		assert.Error(t, err)
		assert.Equal(t, err, expectedErr)
	})
}

func TestListDeletedSecret(t *testing.T) {
	ctx := context.Background()
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	store := mock.NewMockSecretStore(ctrl)
	db := mock2.NewMockSecrets(ctrl)
	logger := testutils.NewMockLogger(ctrl)

	connector := NewConnector(store, db, nil, logger)

	t.Run("should list deleted secret successfully", func(t *testing.T) {
		secretOne := testutils2.FakeSecret()
		secretTwo := testutils2.FakeSecret()

		db.EXPECT().GetAllDeleted(gomock.Any()).Return([]*entities.Secret{secretOne, secretTwo}, nil)

		secretIDs, err := connector.ListDeleted(ctx)

		assert.NoError(t, err)
		assert.Equal(t, secretIDs, []string{secretOne.ID, secretTwo.ID})
	})

	t.Run("should fail to list deleted secret if db fails", func(t *testing.T) {
		expectedErr := errors.PostgresError("cannot connect")

		db.EXPECT().GetAllDeleted(gomock.Any()).Return(nil, expectedErr)

		_, err := connector.ListDeleted(ctx)

		assert.Error(t, err)
		assert.Equal(t, err, expectedErr)
	})
}