package main

import (
	"context"
	"fmt"
	"os"
	"time"

	"github.com/PCC-Grupo-11/TB2-Trabajo-final/internal/config"
	"github.com/PCC-Grupo-11/TB2-Trabajo-final/internal/dataset"
	"github.com/PCC-Grupo-11/TB2-Trabajo-final/internal/logger"
	"github.com/PCC-Grupo-11/TB2-Trabajo-final/internal/ml"
	"github.com/PCC-Grupo-11/TB2-Trabajo-final/internal/storage"
)

func main() {
	cfg, err := config.Load()
	if err != nil {
		logger.Error("invalid config", "error", err)
		os.Exit(1)
	}

	logger.Info("trainer starting", "data_path", cfg.DataPath)

	start := time.Now()

	logger.Info("loading dataset")
	ds, err := dataset.Load(cfg.DataPath)
	if err != nil {
		logger.Error("failed to load dataset", "error", err)
		os.Exit(1)
	}

	logger.Info("dataset loaded", "records", ds.Len())

	trainSet, valSet := ds.Split(config.ValidationSplit)
	logger.Info("dataset split", "train", trainSet.Len(), "val", valSet.Len())

	datasetLoadingTime := time.Since(start).Seconds()
	logger.Info("dataset loading complete", "duration_seconds", datasetLoadingTime)

	logger.Info("training started")
	model := ml.NewModel()

	report := model.Train(trainSet, valSet)
	report.TrainingTimeSeconds = time.Since(start).Seconds() - datasetLoadingTime

	logger.Info("training complete",
		"epochs", report.EpochsTrained,
		"early_stopped", report.EarlyStopped,
		"val_loss", report.FinalValidationLoss,
		"accuracy", report.Accuracy,
		"duration_seconds", report.TrainingTimeSeconds,
	)

	logger.Info("saving model to MongoDB")
	if err := storage.SaveModelWithURI(context.Background(), cfg.MongoURI, model, report); err != nil {
		logger.Error("failed to save model", "error", err)
		os.Exit(1)
	}
	logger.Info("model saved successfully to MongoDB")

	fmt.Printf("\nTraining Report:\n")
	fmt.Printf("\tValidation Loss:    %.4f\n", report.FinalValidationLoss)
	fmt.Printf("\tAccuracy:           %.4f\n", report.Accuracy)
	fmt.Printf("\tMAE:                %.4f\n", report.MAE)
	fmt.Printf("\t+-1 Accuracy:       %.4f\n", report.AccuracyAt1)
	fmt.Printf("\tEpochs Trained:     %d\n", report.EpochsTrained)
	fmt.Printf("\tEarly Stopped:      %v\n", report.EarlyStopped)
	fmt.Printf("\tDuration:           %.2fs\n", report.TrainingTimeSeconds)

	logger.Info("trainer finished")
}
