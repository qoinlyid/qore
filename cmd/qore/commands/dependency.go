package commands

import (
	"errors"
	"maps"
	"path"

	"github.com/qoinlyid/qore"
	"github.com/qoinlyid/qore/internal/templates"
	"github.com/spf13/cobra"
)

var (
	// Dependency root command.
	depCmd = &cobra.Command{
		Use:   "dep",
		Short: "Commands related to the Qore dependency",
		Long: `Commands related to the Qore dependency.
See each sub-command's help for details on how to use the dependency command.`,
		Run: func(cmd *cobra.Command, args []string) { cmd.Help() },
	}

	depAddCmd = &cobra.Command{
		Use:   "add",
		Short: "Add registers new dependency into manifest",
		Long:  `Add registers new dependency into manifest. E.g: add <package_uri_or_path>`,
		RunE:  depAddRunE,
	}
)

func init() {
	// Add sub command.
	depCmd.AddCommand(depAddCmd)

	// Add to root.
	rootCmd.AddCommand(depCmd)
}

func depAddRunE(cmd *cobra.Command, args []string) error {
	// Validate.
	if len(args) == 0 {
		return errors.New("argument package path required")
	}
	pkg := path.Base(args[0])
	name := qore.StringRemoveNonAlphabet(pkg)

	// Read manifest.
	manifest, dir, err := readManifest()
	if err != nil {
		return err
	}

	// Re-write manifest.
	deps := templates.Dependencies{name: templates.Dependency{
		Version: "",
		Path:    pkg,
	}}
	if len(manifest.Dependencies) == 0 {
		manifest.Dependencies = deps
	} else {
		maps.Copy(manifest.Dependencies, deps)
	}
	return templates.WriteManifest(dir, manifest)
}
