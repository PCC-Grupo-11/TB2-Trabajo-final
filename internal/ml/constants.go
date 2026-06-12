package ml

const (
	LearningRate          = 3.0
	Epochs                = 500
	EarlyStoppingPatience = 15
	MinImprovement        = 1e-3
)

const (
	L2Lambda             float32 = 1e-4
	weightInitScale      float32 = 0.05
	probFloor            float64 = 1e-15
	accuracyAt1Threshold int     = 1
)
