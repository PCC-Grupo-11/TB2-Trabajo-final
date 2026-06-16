package storage

import "time"

type ModelDocument struct {
	Weights      []float32 `bson:"weights"`
	Biases       []float32 `bson:"biases"`
	FeatureCount int       `bson:"feature_count"`
	NumClasses   int       `bson:"num_classes"`
	Trained      bool      `bson:"trained"`

	FinalValidationLoss float32 `bson:"final_validation_loss"`
	Accuracy            float64 `bson:"accuracy"`
	MAE                 float64 `bson:"mae"`
	AccuracyAt1         float64 `bson:"accuracy_at_1"`
	EpochsTrained       int     `bson:"epochs_trained"`
	EarlyStopped        bool    `bson:"early_stopped"`
	TotalSamples        int     `bson:"total_samples"`
	TrainingTimeSeconds float64 `bson:"training_time_seconds"`
	LearningRate        float32 `bson:"learning_rate"`
	ShuffleSeed         int64   `bson:"shuffle_seed"`
	ValidationSplit     float64 `bson:"validation_split"`
	ConfusionMatrix     [][]int `bson:"confusion_matrix"`

	CreatedAt time.Time `bson:"created_at"`
}

const (
	ModelsCollection = "models"
	DatabaseName     = "TB2-TF"
)
