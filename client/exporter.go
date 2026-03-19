package client

import (
	"encoding/json"
	"fmt"
	"os"

	"github.com/allofthenamesaretaken/anytype-exporter/utils"
)

type AnytypeExporter struct {
	logger    *utils.Logger
	exportDir string
}

func NewAnytypeExporter(logger *utils.Logger) *AnytypeExporter {
	return &AnytypeExporter{
		logger:    logger,
		exportDir: os.Getenv("EXPORT_DIR"),
	}
}

// func (exporter *AnytypeExporter) gobEncoder(clob string) ([]byte, error) {
// 	var buffer bytes.Buffer
//
// 	err := gob.NewEncoder(&buffer).Encode(clob)
// 	if err != nil {
// 		exporter.logger.Error("Failed to encode clob string into byte slice", err)
// 		return nil, err
// 	}
//
// 	bytes := buffer.Bytes()
//
// 	return bytes, nil
// }

func (exporter *AnytypeExporter) JSONMarshal(data any) ([]byte, error) {
	var bytes []byte

	bytes, err := json.MarshalIndent(data, "", " ")
	if err != nil {
		exporter.logger.Error("Failed to marhsal json struct into byte slice", err)
		return nil, err
	}

	return bytes, nil
}

func (exporter *AnytypeExporter) ensureDir() error {
	if _, err := os.Stat(exporter.exportDir); os.IsNotExist(err) {
		err := os.Mkdir(exporter.exportDir, 0750)
		if err != nil {
			exporter.logger.Error("Failed to make export dir", err)
			return err
		}
	}

	return nil
}

func (exporter *AnytypeExporter) writeFile(path string, bytes []byte) error {
	err := os.WriteFile(path, bytes, 0644)
	if err != nil {
		exporter.logger.Error("Failed to write bytes to export dir", err)
		return err
	}

	return nil
}

func (exporter *AnytypeExporter) ExportObject(object Object) error {
	err := exporter.ensureDir()
	if err != nil {
		return err
	}

	var exportPath string
	var bytes []byte

	exportPath = exporter.exportDir + fmt.Sprintf("/object-%s.md", object.Name)
	bytes = []byte(object.Markdown)

	err = exporter.writeFile(exportPath, bytes)
	if err != nil {
		return err
	}

	tmp := Object(object)
	tmp.Markdown = ""

	exportPath = exporter.exportDir + fmt.Sprintf("/object-%s.json", object.Name)
	bytes, err = exporter.JSONMarshal(tmp)
	if err != nil {
		return err
	}

	err = exporter.writeFile(exportPath, bytes)
	if err != nil {
		return err
	}

	return nil
}
