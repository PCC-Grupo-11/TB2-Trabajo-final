package ml

import (
	"encoding/json"
	"os"
	"path/filepath"
)

func ensureDir(path string) error {
	return os.MkdirAll(filepath.Dir(path), 0755)
}

func (m *Model) Save(path string) error {
	if err := ensureDir(path); err != nil {
		return err
	}
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

func SaveReport(report TrainingReport, path string) error {
	if err := ensureDir(path); err != nil {
		return err
	}
	data, err := json.MarshalIndent(report, "", "  ")
	if err != nil {
		return err
	}

	return os.WriteFile(path, data, 0644)
}
