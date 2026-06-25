package storage

import (
	"context"
	"fmt"
	"time"

	"github.com/PCC-Grupo-11/TB2-Trabajo-final/internal/protocol"
	"go.mongodb.org/mongo-driver/v2/bson"
	"go.mongodb.org/mongo-driver/v2/mongo"
	"go.mongodb.org/mongo-driver/v2/mongo/options"
)

func Connect(ctx context.Context, uri string) (*mongo.Client, error) {
	serverAPI := options.ServerAPI(options.ServerAPIVersion1)
	opts := options.Client().ApplyURI(uri).SetServerAPIOptions(serverAPI)

	client, err := mongo.Connect(opts)
	if err != nil {
		return nil, fmt.Errorf("failed to connect to MongoDB: %w", err)
	}

	if err := client.Ping(ctx, nil); err != nil {
		_ = client.Disconnect(context.Background())
		return nil, fmt.Errorf("failed to ping MongoDB: %w", err)
	}

	return client, nil
}

type Repository struct {
	client *mongo.Client
	db     *mongo.Database
}

func NewRepository(ctx context.Context, uri string) (*Repository, error) {
	client, err := Connect(ctx, uri)
	if err != nil {
		return nil, err
	}

	db := client.Database(DatabaseName)

	coll := db.Collection(UsersCollection)
	index := mongo.IndexModel{
		Keys:    bson.D{{Key: "username", Value: 1}},
		Options: options.Index().SetUnique(true),
	}

	if _, err := coll.Indexes().CreateOne(ctx, index); err != nil {
		client.Disconnect(context.Background())
		return nil, fmt.Errorf("create unique username index: %w", err)
	}

	return &Repository{
		client: client,
		db:     db,
	}, nil
}

func (r *Repository) Close() error {
	return r.client.Disconnect(context.Background())
}

func (r *Repository) Ping(ctx context.Context) error {
	return r.client.Ping(ctx, nil)
}

func (r *Repository) CreateUser(ctx context.Context, username, hashedPassword string) error {
	coll := r.db.Collection(UsersCollection)
	_, err := coll.InsertOne(ctx, User{
		Username:       username,
		HashedPassword: hashedPassword,
		CreatedAt:      time.Now(),
	})
	if err != nil {
		return fmt.Errorf("create user %q: %w", username, err)
	}
	return nil
}

func (r *Repository) FindUser(ctx context.Context, username string) (*User, error) {
	coll := r.db.Collection(UsersCollection)

	var user User
	err := coll.FindOne(ctx, bson.M{"username": username}).Decode(&user)
	if err != nil {
		return nil, fmt.Errorf("find user %q: %w", username, err)
	}
	return &user, nil
}

func (r *Repository) SavePrediction(ctx context.Context, userID string, input protocol.PredictRequest, response protocol.PredictionResult) error {
	coll := r.db.Collection(PredictionsCollection)
	_, err := coll.InsertOne(ctx, Prediction{
		UserID:    userID,
		Input:     input,
		Response:  response,
		CreatedAt: time.Now(),
	})
	return err
}
