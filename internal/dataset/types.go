package dataset

import (
	"math/rand"

	"github.com/PCC-Grupo-11/TB2-Trabajo-final/internal/config"
)

type SparseFeature struct {
	Index uint32
	Value float32
}

type Record struct {
	Y        uint8
	Features [13]SparseFeature
}

type Dataset struct {
	Records []Record
}

func (ds *Dataset) Shuffle() {
	rng := rand.New(rand.NewSource(config.GlobalSeed))
	rng.Shuffle(len(ds.Records), func(i, j int) {
		ds.Records[i], ds.Records[j] = ds.Records[j], ds.Records[i]
	})
}

func (ds *Dataset) Split(valFraction float64) (train, val *Dataset) {
	if valFraction <= 0 || valFraction >= 1 {
		panic("valFraction must be between 0 and 1")
	}
	splitIdx := int(float64(len(ds.Records)) * (1 - valFraction))
	return &Dataset{Records: ds.Records[:splitIdx]}, &Dataset{Records: ds.Records[splitIdx:]}
}

func (ds *Dataset) Len() int {
	return len(ds.Records)
}
