package main

import (
	"fmt"
	"log/slog"
	"os"
	"time"

	_ "net/http/pprof"

	"github.com/PCC-Grupo-11/TB2-Trabajo-final/internal/config"
	"github.com/PCC-Grupo-11/TB2-Trabajo-final/internal/dataset"
	"github.com/PCC-Grupo-11/TB2-Trabajo-final/internal/logger"
	"github.com/PCC-Grupo-11/TB2-Trabajo-final/internal/ml"
)

func main() {
	cfg := config.Load()
	logger.Init(slog.LevelInfo)

	logger.Info("trainer starting", "data_path", cfg.DataPath)

	start := time.Now()

	logger.Info("loading dataset")
	ds, err := dataset.Load(cfg.DataPath)
	if err != nil {
		logger.Error("failed to load dataset", "error", err)
		os.Exit(1)
	}

	logger.Info("dataset loaded", "records", ds.Len())

	trainSet, valSet := ds.Split(ml.ValidationSplit)
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

	logger.Info("saving model", "path", cfg.ModelOutputPath)
	if err := model.Save(cfg.ModelOutputPath); err != nil {
		logger.Error("failed to save model", "error", err)
		os.Exit(1)
	}

	logger.Info("saving metadata", "path", cfg.MetadataPath)
	if err := ml.SaveReport(report, cfg.MetadataPath); err != nil {
		logger.Error("failed to save metadata", "error", err)
		os.Exit(1)
	}

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
