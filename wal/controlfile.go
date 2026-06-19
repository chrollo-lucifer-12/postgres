package wal

import (
	"encoding/json"
	"os"
)

type ControlFile struct {
	LastLSN LSN `json:"last_lsn"`
}

func SaveControlFile(path string, cf ControlFile) error {
	data, err := json.Marshal(cf)
	if err != nil {
		return err
	}

	return os.WriteFile(path, data, 0644)
}

func LoadControlFile(path string) (ControlFile, error) {
	var cf ControlFile

	data, err := os.ReadFile(path)
	if err != nil {
		return cf, err
	}

	err = json.Unmarshal(data, &cf)

	return cf, err
}
