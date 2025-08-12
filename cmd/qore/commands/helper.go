package commands

import (
	"errors"
	"fmt"
	"os"
	"path/filepath"
	"time"

	"github.com/qoinlyid/qore"
	"github.com/qoinlyid/qore/internal/command"
	"github.com/qoinlyid/qore/internal/templates"
)

// readManifest is helper that prompt project directory and read Qore manifest file.
func readManifest() (
	manifest *templates.Manifest,
	dir string,
	err error,
) {
	// Project dir default.
	dir, err = os.Getwd()
	if err != nil {
		return
	}
	verboseMessage(fmt.Sprintf("Working dir: %s", dir))

	// Command to get project manifest.
	_, err = command.NewCommandWizard().
		AddStep(command.WizardStep{
			ID:       "project_dir",
			Type:     command.StepPrompt,
			Question: "Project directory",
			Default:  dir,
			Validate: func(value any) error {
				val, ok := value.(string)
				if !ok {
					return errors.New("invalid project directory format")
				} else if qore.ValidationIsEmpty(val) {
					return errors.New("project directory is required")
				}
				return nil
			},
		}).
		AddProcess("project_manifest", "📁 Reading manifest...",
			func(results command.WizardResults, progress func(string)) (any, error) {
				d, ok := results["project_dir"].(string)
				if !ok {
					return nil, errors.New("invalid project directory value")
				}
				dir = d

				progress("📄 Reading project manifest file...")
				time.Sleep(time.Second)
				m, err := templates.ReadManifest(dir)
				if err != nil {
					return nil, fmt.Errorf("failed to read manifest file on %s: %w", dir, err)
				}
				manifest = m

				progress("⏳ Check manifest value...")
				time.Sleep(time.Second)
				if manifest.Project.Package == "" || manifest.Project.ModuleDir == "" {
					return nil, errors.New("project package nor module directory is empty")
				}

				progress("✅ Project manifest valid")
				return manifest, nil
			},
		).
		Run()
	return
}

// readModule is helper that collecting the existing module inside the project.
func readModule(manifest *templates.Manifest, dir string) (mods []string, err error) {
	_, err = command.NewCommandWizard().
		AddProcess("module_collect", "📥 Collecting module(s)...",
			func(results command.WizardResults, progress func(string)) (any, error) {
				modDir := filepath.Join(dir, manifest.Project.ModuleDir)

				progress("⏳ Collecting existing module(s)...")
				time.Sleep(time.Second)
				entries, err := os.ReadDir(modDir)
				if err != nil {
					return nil, fmt.Errorf("failed to read module directory %s: %w", modDir, err)
				}
				for _, entry := range entries {
					if entry.IsDir() {
						mods = append(mods, entry.Name())
					}
				}
				return "Success", nil
			},
		).
		Run()
	return
}
