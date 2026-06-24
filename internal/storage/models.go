package storage

import (
	"time"

	"github.com/PCC-Grupo-11/TB2-Trabajo-final/internal/ml"
	"github.com/PCC-Grupo-11/TB2-Trabajo-final/internal/protocol"
	"go.mongodb.org/mongo-driver/v2/bson"
)

type ModelDocument struct {
	ml.Model `bson:"inline"`

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

type User struct {
	ID             bson.ObjectID `bson:"_id,omitempty"`
	Username       string        `bson:"username"`
	HashedPassword string        `bson:"hashed_password"`
	CreatedAt      time.Time     `bson:"created_at"`
}

type Prediction struct {
	ID        bson.ObjectID             `bson:"_id,omitempty"`
	UserID    string                    `bson:"user_id"`
	Input     protocol.PredictRequest   `bson:"input"`
	Response  protocol.PredictionResult `bson:"response"`
	CreatedAt time.Time                 `bson:"created_at"`
}

const (
	DatabaseName          = "TB2-TF"
	ModelsCollection      = "models"
	UsersCollection       = "users"
	PredictionsCollection = "predictions"
)
