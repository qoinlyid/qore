package commands

import (
	"errors"
	"fmt"
	"maps"
	"path"
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

	// Module add dependency command
	modDepAddCmd = &cobra.Command{
		Use:   "add-dependency",
		Short: "Add dependency to the module",
		RunE:  modDepAddRunE,
	}
)

var modNewDep bool

func init() {
	// Add sub command.
	modCmd.AddCommand(modNewCmd)
	modCmd.AddCommand(modRemoveCmd)

	modDepAddCmd.Flags().BoolVarP(&modNewDep, "new", "n", false, "--new to add new dependency into module")
	modCmd.AddCommand(modDepAddCmd)

	// Add to root.
	rootCmd.AddCommand(modCmd)
}

// modNewRunE is runner for module new command.
func modNewRunE(cmd *cobra.Command, args []string) error {
	// Read manifest.
	manifest, dir, err := readManifest()
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

// modRemoveRunE is runner for module remove command.
func modRemoveRunE(cmd *cobra.Command, args []string) error {
	// Read manifest.
	manifest, dir, err := readManifest()
	if err != nil {
		return fmt.Errorf("failed to get project manifest: %w", err)
	}

	// Read existing modules.
	var modchoices []command.Option
	mods, err := readModule(manifest, dir)
	if err != nil {
		return fmt.Errorf("failed to get existing modules: %w", err)
	}
	for _, mod := range mods {
		modchoices = append(modchoices, command.Option{
			Key:   mod,
			Value: qore.StringToCammelCase(mod),
		})
	}

	// Command to execute remove existing module(s).
	if len(modchoices) > 0 {
		wizard := command.NewCommandWizard(command.WizardConfig{
			Title:        "🗑️  Module remover",
			Description:  "Remove existing module from your application project",
			ShowProgress: true,
			ClearScreen:  true,
			ClearHistory: true,
			ResultColor:  command.ColorCyan,
		}).
			AddMultiChoice("modules", "📦 Which module(s) you want to remove", modchoices).
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

func modDepAddRunE(cmd *cobra.Command, arg []string) error {
	// Read manifest.
	manifest, dir, err := readManifest()
	if err != nil {
		return fmt.Errorf("failed to get project manifest: %w", err)
	}

	// Read existing modules.
	var modchoices []command.Option
	mods, err := readModule(manifest, dir)
	if err != nil {
		return fmt.Errorf("failed to get existing modules: %w", err)
	}
	for _, mod := range mods {
		modchoices = append(modchoices, command.Option{
			Key:   mod,
			Value: qore.StringToCammelCase(mod),
		})
	}
	if len(modchoices) == 0 {
		return errors.New("does not have any module")
	}

	// Command wizard.
	wizard := command.NewCommandWizard(command.WizardConfig{
		Title:        "🧩 Add Dependency",
		Description:  "Adding dependency into existing module",
		ShowProgress: true,
		ClearScreen:  true,
		ClearHistory: true,
		ResultColor:  command.ColorCyan,
	}).AddMultiChoice("modules", "📦 Which module(s) you want to add dependency", modchoices)
	if modNewDep {
		wizard = wizard.AddPrompt("dep", "Dependency package path", true)
	} else {
		var depchoice []command.Option
		for name, v := range manifest.Dependencies {
			depchoice = append(depchoice, command.Option{
				Key:   v.Path,
				Value: name,
			})
		}
		if len(depchoice) == 0 {
			return errors.New("existing dependency not found, use command flag --new to input new dependency package")
		}
		wizard = wizard.AddChoice("dep", "🧩 Chose dependency", depchoice, true)
	}
	wizard.AddProcess("process", "🚜 Processing...", func(results command.WizardResults, progress func(string)) (any, error) {
		// Get dependency.
		var depname, deppkg string
		progress("🧩 Get dependency...")
		time.Sleep(time.Second)
		switch v := results["dep"].(type) {
		case string:
			depname = path.Base(v)
			deppkg = v
		case command.Option:
			depname = v.Value
			deppkg = fmt.Sprintf("%v", v.Key)
		default:
			return nil, errors.New("dependency is required")
		}

		// Write manifest.
		progress("📝 Writing manifest...")
		time.Sleep(time.Second)
		var deps templates.Dependencies = templates.Dependencies{
			depname: templates.Dependency{Path: deppkg},
		}
		if len(manifest.Dependencies) == 0 {
			manifest.Dependencies = deps
		} else {
			maps.Copy(manifest.Dependencies, deps)
		}
		if err := templates.WriteManifest(dir, manifest); err != nil {
			return nil, err
		}

		// Get module(s).
		moduleopts, _ := results["modules"].([]command.Option)
		if len(moduleopts) == 0 {
			return nil, errors.New("module is required")
		}

		// Generating code.
		progress("⚙️ Generating code...")
		time.Sleep(time.Second)
		var modules []*templates.Module
		for _, opt := range moduleopts {
			modules = append(modules, &templates.Module{
				RootProjectDir: dir,
				Dir:            manifest.Project.ModuleDir,
				Name:           strings.ToLower(opt.Value),
				Pkg:            fmt.Sprintf("%v", opt.Key),
				Dependencies:   deps,
			})
		}
		if err := templates.UpdateModule(modules); err != nil {
			return nil, err
		}
		return "Success", nil
	})

	// Run.
	_, err = wizard.Run()
	return err
}
