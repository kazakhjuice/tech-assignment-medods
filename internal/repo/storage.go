package repository

import (
	"context"
	"errors"
	"log"
	"time"

	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/mongo"
)

type Token struct {
	UUID      string    `bson:"uuid"`
	Token     string    `bson:"hashRefreshToken"`
	ExpiresAt time.Time `bson:"expiresAt"`
}

type Repo struct {
	db *mongo.Collection
}

func NewRepo(db *mongo.Collection) *Repo {
	return &Repo{
		db: db,
	}
}

func (r *Repo) UploadToken(hashedToken string, uuid string) error {

	token := Token{
		UUID:      uuid,
		Token:     hashedToken,
		ExpiresAt: time.Now().Add(31 * 24 * time.Hour),
	}

	filter := bson.M{"uuid": uuid}
	count, err := r.db.CountDocuments(context.Background(), filter)
	if err != nil {
		return err
	}
	if count > 0 {
		return errors.New("token with the same UUID already exists")
	}

	_, err = r.db.InsertOne(context.Background(), token)

	if err != nil {
		return err
	}
	log.Print("succesffully added token to uuid: ", uuid)
	return nil
}

func (r *Repo) UpdateToken(hashedToken string, uuid string) error {
	filter := bson.M{"uuid": uuid}

	update := bson.M{
		"$set": bson.M{
			"hashRefreshToken": hashedToken,
			"expiresAt":        time.Now().Add(31 * 24 * time.Hour),
		},
	}

	_, err := r.db.UpdateOne(context.Background(), filter, update)
	if err != nil {
		return err
	}
	return nil
}

func (r *Repo) GetToken(UUID string) (*Token, error) {
	filter := bson.M{"uuid": UUID}

	var token Token
	err := r.db.FindOne(context.Background(), filter).Decode(&token)
	if err != nil {
		if err == mongo.ErrNoDocuments {
			return nil, nil
		}
		return nil, err
	}
	return &token, nil
}
