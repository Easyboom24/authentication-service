package repository

import (
	"context"
	"encoding/json"
	"errors"
	"go-test/internal/apperror"
	"go-test/internal/domain"
	"go-test/pkg/logging"

	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/bson/primitive"
	"go.mongodb.org/mongo-driver/mongo"
)

func NewUserStorage(datebase *mongo.Database, logger logging.Logger) *DB {
	
	return &DB{
		collection: datebase.Collection("users"),
		logger: logger,
	}
}

func (userDB *DB) Create(ctx context.Context, user domain.User) (string, error) {
	userDB.logger.Debug("Create User")
	result, err := userDB.collection.InsertOne(ctx, user)
	if err != nil {
		return "", err
	}

	userDB.logger.Debug("Convert insertedID to ObjectID")
	oid, ok := result.InsertedID.(primitive.ObjectID)
	if ok {
		return oid.Hex(), nil
	}

	userDB.logger.Trace(user)
	return "", err
}

func (userDB *DB) GetByGUID(ctx context.Context, guid string) (u domain.User, err error) {

	oid, err := primitive.ObjectIDFromHex(guid)
	if err != nil {
		return u, err
	}
	filter := bson.M{"_id": oid}

	userDB.logger.Debug("Find user by GUID")
	result := userDB.collection.FindOne(ctx, filter)
	if result.Err() != nil {
		if errors.Is(result.Err(), mongo.ErrNoDocuments) {

			return u, apperror.ErrNotFound
		}
		return u, err
	}
	userDB.logger.Debug("User find result decode")
	if err = result.Decode(&u); err != nil {
		return u, err
	}

	r, err := json.Marshal(u)
	var a domain.User
	json.Unmarshal(r,&a)
	return a, nil
}

func (userDB *DB) DeleteSession(ctx context.Context, user domain.User) error {
	userDB.logger.Debugf("Deleting a session from a user (GUID = %s) with a fingerprint %s",user.GUID,user.Sessions[0].FingerPrint)
	oid, err := primitive.ObjectIDFromHex(user.GUID)
	if err != nil {
		return err
	}
	filter := bson.M{"_id":oid}
	update := bson.M{"$pull": bson.M{"sessions": bson.M{"fingerprint":user.Sessions[0].FingerPrint}}}
	_, err = userDB.collection.UpdateOne(ctx, filter, update)
	if err != nil {
		return err
	}
	
	return nil
}

func (userDB *DB) CreateSession(ctx context.Context, user domain.User) error {	
	userDB.logger.Debugf("Creating a user session with GIOD %s",user.GUID)
	
	objectID, err := primitive.ObjectIDFromHex(user.GUID)
	if err != nil {
		return err
	}
		
	if err != nil {
		return err
	}
	filter := bson.M{"_id": objectID}
	update := bson.M{"$push": bson.M{"sessions": user.Sessions[len(user.Sessions)-1]}}
	result := userDB.collection.FindOneAndUpdate(ctx, filter, update)
	
	if result.Err() != nil {
		return err
	}

	return nil
}

func (userDB *DB) GetByFingerPrint(ctx context.Context, fingerPrint string) (domain.User, error) {
	userDB.logger.Debug("Find user by fingerprint")
	var results []domain.User
	filter := bson.A{
		bson.D{
			{Key: "$project", 
				Value: bson.D{
					{Key: "username", Value: 1},
					{Key: "sessions", 
						Value: bson.D{
							{Key: "$filter",
								Value: bson.D{
									{Key: "input", Value: "$sessions"},
									{Key: "as", Value: "s"},
									{Key: "limit", Value: 1},
									{Key: "cond", 
										Value: bson.D{
											{Key: "$eq",
												Value: bson.A{"$$s.fingerprint", fingerPrint},
											},
										},
									},
								},
							},
						},
					},
				},
			},
		},
	}

	cursor, err := userDB.collection.Aggregate(ctx, filter)
	if cursor.Err() != nil {
		if errors.Is(cursor.Err(), mongo.ErrNoDocuments) {

			return domain.User{}, apperror.ErrNotFound
		}
		return domain.User{}, err
	}
	
	err = cursor.All(ctx, &results)
	if err != nil {
		return domain.User{}, err
	}
	
	return results[0], nil
}