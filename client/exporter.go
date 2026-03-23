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

func (exporter *AnytypeExporter) ensureDir(path string) error {
	if _, err := os.Stat(path); os.IsNotExist(err) {
		err := os.Mkdir(path, 0750)
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

func (exporter *AnytypeExporter) exportMd(exportDir string, object Object) error {
	exportPath := exportDir + "/object.md"
	bytes := []byte(object.Markdown)
	err := exporter.writeFile(exportPath, bytes)
	if err != nil {
		return err
	}
	return nil
}

func (exporter *AnytypeExporter) exportJSON(exportDir string, object Object) error {
	exportPath := exportDir + "/object.meta.json"
	tmp := Object(object)
	tmp.Markdown = ""
	bytes, err := json.MarshalIndent(tmp, "", " ")
	if err != nil {
		exporter.logger.Error("Failed to marhsal json struct into byte slice", err)
		return err
	}
	err = exporter.writeFile(exportPath, bytes)
	if err != nil {
		return err
	}
	return nil
}

func (exporter *AnytypeExporter) exportObject(object Object) error {
	targetDir := fmt.Sprintf("/%s", object.Name)
	objectExportDir := exporter.exportDir + targetDir
	err := exporter.ensureDir(objectExportDir)

	err = exporter.exportMd(objectExportDir, object)
	if err != nil {
		return err
	}

	err = exporter.exportJSON(objectExportDir, object)
	if err != nil {
		return err
	}

	return nil
}

func (exporter *AnytypeExporter) ExportObjects(objects []Object) error {
	err := exporter.ensureDir(exporter.exportDir)
	if err != nil {
		return err
	}

	for _, object := range objects {
		err := exporter.exportObject(object)
		if err != nil {
			return err
		}
	}

	return nil
}
