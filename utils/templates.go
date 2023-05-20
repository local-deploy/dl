package utils

import (
	"embed"
	"os"
	"path/filepath"

	"github.com/local-deploy/dl/helper"
)

// Templates directory
var Templates embed.FS

// CreateTemplates create docker-compose files
func CreateTemplates(overwrite bool) error {
	templateDir := helper.TemplateDir()
	configDir := filepath.Join(templateDir, "config-files")

	// delete existing directory
	if overwrite {
		err := helper.RemoveDirectory(configDir)
		if err != nil {
			return err
		}
	}

	err := helper.CreateDirectory(configDir)
	if err != nil {
		return err
	}

	entries, err := Templates.ReadDir("config-files")
	if err != nil {
		return err
	}

	for _, entry := range entries {
		out, err := os.Create(filepath.Join(configDir, entry.Name()))
		if err != nil {
			return err
		}

		data, err := Templates.ReadFile(filepath.Join("config-files", entry.Name()))
		if err != nil {
			return err
		}

		_, err = out.Write(data)
		if err != nil {
			return err
		}
	}

	return nil
}
