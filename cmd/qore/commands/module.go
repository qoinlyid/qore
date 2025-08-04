package commands

import (
	"errors"
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"time"

	"github.com/qoinlyid/qore"
	"github.com/qoinlyid/qore/internal/command"
	"github.com/qoinlyid/qore/internal/templates"
	"github.com/spf13/cobra"
)

var (
	// Module root command.
	modCmd = &cobra.Command{
		Use:   "mod",
		Short: "Commands related to the Qore module",
		Long: `Commands related to the Qore module.
See each sub-command's help for details on how to use the module command.`,
		Run: func(cmd *cobra.Command, args []string) { cmd.Help() },
	}

	// Module new command.
	modNewCmd = &cobra.Command{
		Use:   "new",
		Short: "Add new module into the project",
		RunE:  modNewRunE,
	}

	// Module remove command.
	modRemoveCmd = &cobra.Command{
		Use:   "remove",
		Short: "Remove existing module from the project",
		RunE:  modRemoveRunE,
	}
)

func init() {
	// Add sub command.
	modCmd.AddCommand(modNewCmd)
	modCmd.AddCommand(modRemoveCmd)

	// Add to root.
	rootCmd.AddCommand(modCmd)
}

func modNewRunE(cmd *cobra.Command, args []string) error {
	// Project dir default.
	dir, err := os.Getwd()
	if err != nil {
		return fmt.Errorf("failed to get working directory: %w", err)
	}
	verboseMessage(fmt.Sprintf("Working dir: %s", dir))

	// Command to get project manifest.
	var manifest *templates.Manifest
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
	if err != nil {
		return fmt.Errorf("failed to get project manifest: %w", err)
	}

	// Create command wizard.
	wizard := command.NewCommandWizard(command.WizardConfig{
		Title:         "🧩 Module creator",
		Description:   "Add new module into your application project",
		ShowProgress:  true,
		ClearScreen:   true,
		ClearHistory:  true,
		ResultColor:   command.ColorCyan,
		ShowFinish:    true,
		FinishMessage: "✨ Module addedd successfully!",
	})

	// Module name
	wizard.AddStep(command.WizardStep{
		ID:       "module_name",
		Type:     command.StepPrompt,
		Question: "🔐 Module name",
		Required: true,
		Validate: func(value any) error {
			val, ok := value.(string)
			if !ok {
				return errors.New("invalid module name format")
			} else if qore.StringIsFirstCharNonAlphabet(val) {
				return errors.New("first char of module name cannot use non-alphabet")
			}
			return nil
		},
		Transform: func(value any) any {
			if val, ok := value.(string); ok {
				return strings.ToLower(val)
			}
			return value
		},
	})

	// Using dependencies.
	var deps []command.Option
	for name := range manifest.Dependencies {
		deps = append(deps, command.Option{
			Key:   name,
			Value: qore.StringToCammelCase(name),
		})
	}
	if len(deps) > 0 {
		wizard.AddStep(command.WizardStep{
			ID:       "dependencies",
			Type:     command.StepMultiChoice,
			Question: "🧬 Which dependencies that would you like to add to your project",
			Required: false,
			Options:  deps,
		})
	}

	// Project directory & manifest.
	// wizard.AddStep(command.WizardStep{
	// 	ID:       "project_dir",
	// 	Type:     command.StepPrompt,
	// 	Question: "Project directory",
	// 	Default:  dir,
	// 	Validate: func(value any) error {
	// 		val, ok := value.(string)
	// 		if !ok {
	// 			return errors.New("invalid project directory format")
	// 		} else if qore.ValidationIsEmpty(val) {
	// 			return errors.New("project directory is required")
	// 		}
	// 		return nil
	// 	},
	// }).AddProcess("project_manifest", "📁 Reading manifest...",
	// 	func(results command.WizardResults, progress func(string)) (any, error) {
	// 		dir, ok := results["project_dir"].(string)
	// 		if !ok {
	// 			return nil, errors.New("invalid project directory value")
	// 		}

	// 		progress("📄 Reading project manifest file...")
	// 		time.Sleep(time.Second)
	// 		manifest, err := templates.ReadManifest(dir)
	// 		if err != nil {
	// 			return nil, fmt.Errorf("failed to read manifest file on %s: %w", dir, err)
	// 		}

	// 		progress("⏳ Check manifest value...")
	// 		time.Sleep(time.Second)
	// 		if manifest.Project.Package == "" || manifest.Project.ModuleDir == "" {
	// 			return nil, errors.New("project package nor module directory is empty")
	// 		}

	// 		progress("✅ Project manifest valid")
	// 		return manifest, nil
	// 	},
	// )

	// Module publicity & process generate module.
	wizard.
		AddConfirm("module_flag_public", "👀 Make public", false).
		AddProcess("module_generate", "📦 Generating module...",
			func(results command.WizardResults, progress func(string)) (any, error) {
				progress("⏳ Get command value of module name & manifest...")
				time.Sleep(time.Second)
				name, ok := results["module_name"].(string)
				if !ok {
					return nil, errors.New("invalid module name value")
				}
				public, _ := results["module_flag_public"].(bool)
				modname := qore.StringRemoveNonAlphabet(name)

				// Get desired dependencies that want to use.
				var useddeps templates.Dependencies
				depchoices, _ := results["dependencies"].([]command.Option)
				if len(depchoices) > 0 {
					useddeps = make(templates.Dependencies, len(depchoices))
					for _, v := range depchoices {
						name := fmt.Sprintf("%v", v.Key)
						if dep, ok := manifest.Dependencies[name]; ok {
							useddeps[name] = dep
						}
					}
				}

				progress("🚧 Generate module code...")
				time.Sleep(time.Second)
				if err := templates.NewModule(&templates.Module{
					// Module.
					Name:       modname,
					Struct:     qore.StringToCammelCase(name),
					Pkg:        modname,
					Dir:        filepath.Join(dir, manifest.Project.ModuleDir, modname),
					MakePublic: public,
					// Project.
					RootProjectDir: dir,
					ProjectPackage: manifest.Project.Package,
					ModuleDir:      manifest.Project.ModuleDir,
					Dependencies:   useddeps,
				}); err != nil {
					return nil, err
				}
				return "Success", nil
			},
		)

	// Run command wizard.
	_, err = wizard.Run()
	if err != nil {
		return fmt.Errorf("failed add module: %w", err)
	}
	return nil
}

func modRemoveRunE(cmd *cobra.Command, args []string) error {
	// Project dir default.
	dir, err := os.Getwd()
	if err != nil {
		return fmt.Errorf("failed to get working directory: %w", err)
	}
	verboseMessage(fmt.Sprintf("Working dir: %s", dir))

	// Command to collecting existing module(s).
	var mods []command.Option
	var manifest *templates.Manifest
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
		AddProcess("module_collect", "📥 Collecting module(s)...",
			func(results command.WizardResults, progress func(string)) (any, error) {
				d, ok := results["project_dir"].(string)
				if !ok {
					return nil, errors.New("invalid project directory value")
				}
				dir = d
				manifest, ok := results["project_manifest"].(*templates.Manifest)
				if !ok {
					return nil, errors.New("invalid manifest name value")
				}
				modDir := filepath.Join(dir, manifest.Project.ModuleDir)

				progress("⏳ Collecting existing module(s)...")
				time.Sleep(time.Second)
				entries, err := os.ReadDir(modDir)
				if err != nil {
					return nil, fmt.Errorf("failed to read module directory %s: %w", modDir, err)
				}
				for _, entry := range entries {
					if entry.IsDir() {
						mods = append(mods, command.Option{
							Key:   entry.Name(),
							Value: qore.StringToCammelCase(entry.Name()),
						})
					}
				}
				return "Success", nil
			},
		).
		Run()
	if err != nil {
		return fmt.Errorf("failed collect existing module(s): %w", err)
	}

	// Command to execute remove existing module(s).
	if len(mods) > 0 {
		wizard := command.NewCommandWizard(command.WizardConfig{
			Title:        "🗑️  Module remover",
			Description:  "Remove existing module from your application project",
			ShowProgress: true,
			ClearScreen:  true,
			ClearHistory: true,
			ResultColor:  command.ColorCyan,
		}).
			AddMultiChoice("modules", "📦 Which module(s) you want to remove", mods).
			AddConfirm("remove_confirm", "Are you sure to remove module(s)", false).
			AddProcess("remove", "❌ Removing module(s)...",
				func(results command.WizardResults, progress func(string)) (any, error) {
					confirm, _ := results["remove_confirm"].(bool)
					if !confirm {
						return "Canceled", nil
					}
					modules, _ := results["modules"].([]command.Option)
					if len(modules) == 0 {
						return "Noting selected", nil
					}

					var removes []string
					for _, mod := range modules {
						progress(fmt.Sprintf("❌ Removing module '%s'...", mod.Value))
						time.Sleep(time.Second)
						removes = append(removes, fmt.Sprintf("%v", mod.Key))
					}
					if err := templates.RemoveModule(&templates.Module{
						// Project.
						RootProjectDir: dir,
						ProjectPackage: manifest.Project.Package,
						ModuleDir:      manifest.Project.ModuleDir,
					}, removes); err != nil {
						return nil, err
					}
					return "Success", nil
				},
			)

		// Run.
		if _, err := wizard.Run(); err != nil {
			return fmt.Errorf("failed remove module: %w", err)
		}
		confirmed, _ := wizard.GetBoolResult("remove_confirm")
		if !confirmed {
			fmt.Println("\n❌ Process cancelled!")
		} else {
			fmt.Println("\n🔥 Module removed successfully!")
		}
	}
	return nil
}
