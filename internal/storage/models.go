package storage

import (
	"time"

	"github.com/PCC-Grupo-11/TB2-Trabajo-final/internal/ml"
	"github.com/PCC-Grupo-11/TB2-Trabajo-final/internal/protocol"
	"go.mongodb.org/mongo-driver/v2/bson"
)

type ModelDocument struct {
	ml.Model

	FinalValidationLoss float32   `json:"final_validation_loss"`
	Accuracy            float64   `json:"accuracy"`
	MAE                 float64   `json:"mae"`
	AccuracyAt1         float64   `json:"accuracy_at_1"`
	EpochsTrained       int       `json:"epochs_trained"`
	EarlyStopped        bool      `json:"early_stopped"`
	TotalSamples        int       `json:"total_samples"`
	TrainingTimeSeconds float64   `json:"training_time_seconds"`
	LearningRate        float32   `json:"learning_rate"`
	ShuffleSeed         int64     `json:"shuffle_seed"`
	ValidationSplit     float64   `json:"validation_split"`
	ConfusionMatrix     [][]int   `json:"confusion_matrix"`
	CreatedAt           time.Time `json:"created_at"`
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
