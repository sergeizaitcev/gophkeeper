package mongodb

import (
	"context"
	"errors"

	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/bson/primitive"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"

	"github.com/sergeizaitcev/gophkeeper/internal/router"
	"github.com/sergeizaitcev/gophkeeper/pkg/passwords"
)

const collectionName = "users"

var _ router.Storage = (*Users)(nil)

type User struct {
	ID           primitive.ObjectID `json:"_id"           bson:"_id"`
	Username     string             `json:"username"      bson:"username"`
	HashPassword string             `json:"hash_password" bson:"hash_password"`
	Filepath     string             `json:"filepath"      bson:"filepath"`
}

// Users определяет хранилище данных пользователей.
type Users struct {
	db *mongo.Collection
}

// New возвращает новый экземпляр Users.
func New(db *mongo.Database) *Users {
	return &Users{db: db.Collection(collectionName)}
}

// MigrateIndex создает индекс в коллекции.
func (s *Users) MigrateIndex(ctx context.Context) error {
	index := mongo.IndexModel{
		Keys:    bson.D{{Key: "username", Value: 1}},
		Options: options.Index().SetUnique(true),
	}
	_, err := s.db.Indexes().CreateOne(ctx, index)
	return err
}

func (u *Users) Register(ctx context.Context, username, password string) (string, error) {
	token, err := u.Login(ctx, username, password)
	if err == nil {
		return token, nil
	}
	if !errors.Is(err, mongo.ErrNoDocuments) {
		return "", err
	}

	hash, err := passwords.Hash(password)
	if err != nil {
		return "", err
	}

	id := primitive.NewObjectID()

	doc := bson.D{
		{Key: "_id", Value: id},
		{Key: "username", Value: username},
		{Key: "hash_password", Value: hash},
	}

	if _, err = u.db.InsertOne(ctx, doc); err != nil {
		return "", err
	}

	return id.Hex(), nil
}

func (u *Users) Login(ctx context.Context, username, password string) (string, error) {
	var user User

	filter := bson.D{{Key: "username", Value: username}}
	if err := u.db.FindOne(ctx, filter).Decode(&user); err != nil {
		return "", err
	}

	if !passwords.Compare(user.HashPassword, password) {
		return "", errors.New("password is invalid")
	}

	return user.ID.Hex(), nil
}

func (u *Users) Check(ctx context.Context, token string) error {
	id, err := primitive.ObjectIDFromHex(token)
	if err != nil {
		return err
	}
	filter := bson.D{{Key: "_id", Value: id}}
	return u.db.FindOne(ctx, filter).Err()
}

func (u *Users) Save(ctx context.Context, token, filepath string) error {
	id, err := primitive.ObjectIDFromHex(token)
	if err != nil {
		return err
	}

	filter := bson.D{{Key: "_id", Value: id}}
	update := bson.D{{
		Key:   "$set",
		Value: bson.D{{Key: "filepath", Value: filepath}},
	}}

	if _, err = u.db.UpdateOne(ctx, filter, update); err != nil {
		return err
	}

	return nil
}

func (u *Users) Get(ctx context.Context, token string) (string, error) {
	id, err := primitive.ObjectIDFromHex(token)
	if err != nil {
		return "", err
	}

	var user User
	filter := bson.D{{Key: "_id", Value: id}}

	if err = u.db.FindOne(ctx, filter).Decode(&user); err != nil {
		return "", err
	}

	return user.Filepath, nil
}
