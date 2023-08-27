package repository

import (
	"context"
	"go-test/internal/domain"
	"go-test/pkg/logging"

	"go.mongodb.org/mongo-driver/mongo"
)

type UserStorage interface {
	Create(ctx context.Context, user domain.User) (string, error)
	CreateSession(ctx context.Context, user domain.User) error
	DeleteSession(ctx context.Context, user domain.User) error
	GetByGUID(ctx context.Context, guid string) (domain.User, error)
	GetByFingerPrint(ctx context.Context, fingerPrint string) (domain.User, error)
}


type Storages struct {
	UserStorage UserStorage
}

type DB struct {
	collection *mongo.Collection
	logger logging.Logger
}

func NewStorages(db *mongo.Database, logger logging.Logger) *Storages {
	return &Storages{
		UserStorage: NewUserStorage(db, logger),
	}
}


