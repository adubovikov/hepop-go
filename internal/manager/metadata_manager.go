package manager

import (
	"encoding/json"
	"os"
	"path/filepath"
)

type Metadata struct {
	WriterID string `json:"writer_id"`
	NextID   int    `json:"next_id"`
}

type MetadataManager struct {
	baseDir string
}

func NewMetadataManager(baseDir string) *MetadataManager {
	return &MetadataManager{baseDir: baseDir}
}

func (m *MetadataManager) LoadMetadata() (*Metadata, error) {
	filePath := filepath.Join(m.baseDir, "metadata.json")
	file, err := os.Open(filePath)
	if err != nil {
		if os.IsNotExist(err) {
			return &Metadata{WriterID: "default", NextID: 0}, nil
		}
		return nil, err
	}
	defer file.Close()

	var metadata Metadata
	if err := json.NewDecoder(file).Decode(&metadata); err != nil {
		return nil, err
	}
	return &metadata, nil
}

func (m *MetadataManager) SaveMetadata(metadata *Metadata) error {
	filePath := filepath.Join(m.baseDir, "metadata.json")
	file, err := os.Create(filePath)
	if err != nil {
		return err
	}
	defer file.Close()

	return json.NewEncoder(file).Encode(metadata)
}
