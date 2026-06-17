package storage

import (
	"context"
	"fmt"
	"time"

	"github.com/PCC-Grupo-11/TB2-Trabajo-final/internal/ml"
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

func LoadLatestModel(ctx context.Context, client *mongo.Client) (*ml.Model, error) {
	coll := client.Database(DatabaseName).Collection(ModelsCollection)
	opts := options.FindOne().SetSort(bson.D{{Key: "created_at", Value: -1}})

	var doc ModelDocument

	if err := coll.FindOne(ctx, bson.D{}, opts).Decode(&doc); err != nil {
		if err == mongo.ErrNoDocuments {
			return nil, fmt.Errorf("no trained model found in database %q collection %q", DatabaseName, ModelsCollection)
		}
		return nil, fmt.Errorf("failed to load model: %w", err)
	}

	return &ml.Model{
		FeatureCount: doc.FeatureCount,
		NumClasses:   doc.NumClasses,
		Weights:      doc.Weights,
		Biases:       doc.Biases,
		Trained:      doc.Trained,
	}, nil
}

func SaveModel(ctx context.Context, client *mongo.Client, model *ml.Model, report ml.TrainingReport) error {
	coll := client.Database(DatabaseName).Collection(ModelsCollection)

	doc := ModelDocument{
		Weights:             model.Weights,
		Biases:              model.Biases,
		FeatureCount:        model.FeatureCount,
		NumClasses:          model.NumClasses,
		Trained:             model.Trained,
		FinalValidationLoss: report.FinalValidationLoss,
		Accuracy:            report.Accuracy,
		MAE:                 report.MAE,
		AccuracyAt1:         report.AccuracyAt1,
		EpochsTrained:       report.EpochsTrained,
		EarlyStopped:        report.EarlyStopped,
		TotalSamples:        report.TotalSamples,
		TrainingTimeSeconds: report.TrainingTimeSeconds,
		LearningRate:        report.LearningRate,
		ShuffleSeed:         report.ShuffleSeed,
		ValidationSplit:     report.ValidationSplit,
		ConfusionMatrix:     report.ConfusionMatrix,
		CreatedAt:           time.Now(),
	}

	_, err := coll.InsertOne(ctx, doc)
	if err != nil {
		return fmt.Errorf("failed to save model: %w", err)
	}
	return nil
}

func LoadLatestModelFromURI(uri string) (*ml.Model, error) {
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	client, err := Connect(ctx, uri)
	if err != nil {
		return nil, err
	}
	defer client.Disconnect(context.Background())

	return LoadLatestModel(ctx, client)
}

func SaveModelWithURI(uri string, model *ml.Model, report ml.TrainingReport) error {
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	client, err := Connect(ctx, uri)
	if err != nil {
		return err
	}
	defer client.Disconnect(context.Background())

	return SaveModel(ctx, client, model, report)
}
