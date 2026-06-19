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
