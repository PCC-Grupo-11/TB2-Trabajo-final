package storage

import (
	"context"
	"time"

	"github.com/PCC-Grupo-11/TB2-Trabajo-final/internal/ml"
	"go.mongodb.org/mongo-driver/v2/bson"
	"go.mongodb.org/mongo-driver/v2/mongo"
	"go.mongodb.org/mongo-driver/v2/mongo/options"
)

func Connect(ctx context.Context, uri string) (*mongo.Client, error) {
	client, err := mongo.Connect(options.Client().ApplyURI(uri))
	if err != nil {
		return nil, err
	}
	if err := client.Ping(ctx, nil); err != nil {
		return nil, err
	}
	return client, nil
}

func LoadLatestModel(ctx context.Context, client *mongo.Client) (*ml.Model, error) {
	coll := client.Database(DatabaseName).Collection(ModelsCollection)
	opts := options.FindOne().SetSort(bson.D{{Key: "created_at", Value: -1}})

	var doc ModelDocument

	if err := coll.FindOne(ctx, bson.D{}, opts).Decode(&doc); err != nil {
		return nil, err
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
	return err
}

func LoadLatestModelFromURI(ctx context.Context, uri string) (*ml.Model, error) {
	client, err := Connect(ctx, uri)
	if err != nil {
		return nil, err
	}
	defer client.Disconnect(ctx)

	return LoadLatestModel(ctx, client)
}

func SaveModelWithURI(ctx context.Context, uri string, model *ml.Model, report ml.TrainingReport) error {
	client, err := Connect(ctx, uri)
	if err != nil {
		return err
	}
	defer client.Disconnect(ctx)

	return SaveModel(ctx, client, model, report)
}
