package storage

import (
	"context"
	"time"

	"github.com/PCC-Grupo-11/TB2-Trabajo-final/internal/protocol"
	"go.mongodb.org/mongo-driver/v2/bson"
	"go.mongodb.org/mongo-driver/v2/mongo"
)

type Repository struct {
	client *mongo.Client
	db     *mongo.Database
}

func NewRepository(ctx context.Context, uri string) (*Repository, error) {
	client, err := Connect(ctx, uri)
	if err != nil {
		return nil, err
	}
	return &Repository{
		client: client,
		db:     client.Database(DatabaseName),
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
	return err
}

func (r *Repository) FindUser(ctx context.Context, username string) (*User, error) {
	coll := r.db.Collection(UsersCollection)
	var user User
	err := coll.FindOne(ctx, bson.M{"username": username}).Decode(&user)
	if err != nil {
		return nil, err
	}
	return &user, nil
}

func (r *Repository) SavePrediction(ctx context.Context, userID string, input protocol.PredictRequest, response protocol.PredictionResult, latencyMs float64) error {
	coll := r.db.Collection(PredictionsCollection)
	_, err := coll.InsertOne(ctx, Prediction{
		UserID:    userID,
		Input:     input,
		Response:  response,
		LatencyMs: latencyMs,
		CreatedAt: time.Now(),
	})
	return err
}
