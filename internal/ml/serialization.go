package ml

import (
	"encoding/json"
	"fmt"
	"os"
)

func (m *Model) Save(path string) error {
	f, err := os.Create(path)
	if err != nil {
		return err
	}
	defer f.Close()
	return json.NewEncoder(f).Encode(m)
}

func LoadModel(path string) (*Model, error) {
	f, err := os.Open(path)
	if err != nil {
		return nil, err
	}
	defer f.Close()
	var m Model
	if err := json.NewDecoder(f).Decode(&m); err != nil {
		return nil, err
	}
	return &m, nil
}

func formatConfusionMatrix(mat [][]int) string {
	if len(mat) == 0 {
		return ""
	}
	nc := len(mat)
	s := ""
	// header
	s += "      "
	for j := 0; j < nc; j++ {
		s += fmt.Sprintf("%4d ", j)
	}
	s += "\n"
	// rows
	for i := 0; i < nc; i++ {
		s += fmt.Sprintf("%4d: ", i)
		for j := 0; j < nc; j++ {
			s += fmt.Sprintf("%4d ", mat[i][j])
		}
		s += "\n"
	}
	return s[:len(s)-1] // trim trailing newline
}

func SaveReport(report *TrainingReport, path string) error {
	data, err := json.MarshalIndent(report, "", "  ")
	if err != nil {
		return err
	}

	return os.WriteFile(path, data, 0644)
}
