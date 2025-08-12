package templates

import (
	"errors"
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"github.com/pelletier/go-toml/v2"
)

const manifestFilename = "Qore.toml"

// Manifest defines manifest data for Qore project.
type Manifest struct {
	Project      Project      `mapstructure:"project" toml:"project"`
	Dependencies Dependencies `mapstructure:"dependencies" toml:"dependencies"`
}

// Project defines project data in manifest.
type Project struct {
	Name        string `mapstructure:"name" toml:"name"`
	Version     string `mapstructure:"version" toml:"version"`
	Description string `mapstructure:"description" toml:"description"`
	Package     string `mapstructure:"package" toml:"package"`
	ModuleDir   string `mapstructure:"module-dir" toml:"module-dir"`
	Repository  string `mapstructure:"repository" toml:"repository"`
	CreatedAt   string `mapstructure:"created-at" toml:"created-at"`
}

func encodeManifest(manifest *Manifest) ([]byte, error) {
	if manifest == nil {
		return nil, errors.New("manifest is required")
	}

	project, err := toml.Marshal(struct {
		Project Project `toml:"project"`
	}{manifest.Project})
	if err != nil {
		return nil, fmt.Errorf("failed to encode manifest: %w", err)
	}

	var result strings.Builder
	result.Write(project)
	if len(manifest.Dependencies) > 0 {
		result.WriteString("\n[dependencies]\n")
		for name, dep := range manifest.Dependencies {
			result.WriteString(fmt.Sprintf(
				"%s = { version = '%s', path = '%s' }\n",
				name, dep.Version, dep.Path,
			))
		}
	}
	return []byte(result.String()), nil
}

func decodeManifest(data []byte) (*Manifest, error) {
	// Decode manifest.
	var manifest Manifest
	if err := toml.Unmarshal(data, &manifest); err != nil {
		return nil, fmt.Errorf("failed to decode manifest: %w", err)
	}
	return &manifest, nil
}

// ReadManifest reads manifest file in the project directory.
func ReadManifest(dir string) (*Manifest, error) {
	fpath := filepath.Join(dir, manifestFilename)

	// Read manifest file.
	data, err := os.ReadFile(fpath)
	if err != nil {
		return nil, fmt.Errorf("failed to read manifest file %s: %w", fpath, err)
	}

	// Decode manifest.
	return decodeManifest(data)
}

// WriteManifest writes manifest data to the project manifest file,
func WriteManifest(dir string, manifest *Manifest) error {
	b, err := encodeManifest(manifest)
	if err != nil {
		return err
	}

	fpath := filepath.Join(dir, manifestFilename)
	err = os.WriteFile(fpath, b, 0644)
	if err != nil {
		return fmt.Errorf("failed to write manifest %s: %w", fpath, err)
	}
	return nil
}
