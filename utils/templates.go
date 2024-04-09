package utils

import (
	"embed"
	"os"
	"path/filepath"
)

// Templates directory
var Templates embed.FS

// CreateTemplates create docker-compose files
func CreateTemplates(overwrite bool) error {
	templateDir := TemplateDir()

	// delete existing directory
	if overwrite {
		err := RemoveDirectory(templateDir)
		if err != nil {
			return err
		}
	}

	err := CreateDirectory(templateDir)
	if err != nil {
		return err
	}

	entries, err := Templates.ReadDir("config-files")
	if err != nil {
		return err
	}

	for _, entry := range entries {
		out, err := os.Create(filepath.Join(templateDir, entry.Name()))
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
