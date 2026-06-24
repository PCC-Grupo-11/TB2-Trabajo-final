package storage

import (
	"encoding/json"
	"fmt"
	"os"
	"time"

	"github.com/PCC-Grupo-11/TB2-Trabajo-final/internal/ml"
)

func SaveModelToFile(model *ml.Model, report ml.TrainingReport, path string) error {
	doc := ModelDocument{
		Model:               *model,
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

	data, err := json.MarshalIndent(doc, "", "  ")
	if err != nil {
		return fmt.Errorf("failed to marshal model: %w", err)
	}

	if err := os.WriteFile(path, data, 0644); err != nil {
		return fmt.Errorf("failed to write model file %q: %w", path, err)
	}

	return nil
}
