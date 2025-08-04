package commands

// import (
// 	"errors"
// 	"fmt"
// 	"os"
// 	"os/exec"
// 	"path"
// 	"strings"

// 	"github.com/qoinlyid/qore"
// 	"github.com/qoinlyid/qore/internal/templates"
// 	"github.com/spf13/cobra"
// )

// var (
// 	// Dependency root command.
// 	depCmd = &cobra.Command{
// 		Use:   "dep",
// 		Short: "Commands related to the Qore dependency",
// 		Long: `Commands related to the Qore dependency.
// See each sub-command's help for details on how to use the dependency command.`,
// 		Run: func(cmd *cobra.Command, args []string) { cmd.Help() },
// 	}

// 	// Dependency new command.
// 	depNewCmd = &cobra.Command{
// 		Use:   "new",
// 		Short: "Create new Qore dependency package",
// 		Long: `Create new Qore dependency package.
// Provide dependency name as an command argument, e.g new redis`,
// 		RunE: depNewRunE,
// 	}
// )

// const (
// 	depInitFlagIgnoreSuffix = "ignore-suffix"
// 	depInitFlagOut          = "out"
// 	depPkgUrl               = "pkg-url"
// )

// func init() {
// 	// Add dep init command.
// 	depNewCmd.Flags().Bool(depInitFlagIgnoreSuffix, false, "Do not add default 'dep' as package name suffix.")
// 	depNewCmd.Flags().String(depInitFlagOut, "", "Dependency out directory path.")
// 	depNewCmd.Flags().String(depPkgUrl, "", "Dependency package URL without package name (e.g. github.com/foo). If not present, only be local package without go module.")
// 	depCmd.AddCommand(depNewCmd)

// 	// Add to root.
// 	rootCmd.AddCommand(depCmd)
// }

// func depNewRunE(cmd *cobra.Command, args []string) error {
// 	// Check given argument is it valid or not.
// 	switch {
// 	case len(args) == 0:
// 		return errors.New("provide dependency name as an command argument")
// 	case qore.StringIsFirstCharNonAlphabet(args[0]):
// 		return errors.New("first char of dependency name cannot use non-alphabet")
// 	}
// 	verboseMessage("Dependency name valid")
// 	name := strings.ToLower(args[0])

// 	// Get flags.
// 	ignoreSuffixFlag, _ := cmd.Flags().GetBool(depInitFlagIgnoreSuffix)
// 	flagOut, _ := cmd.Flags().GetString(depInitFlagOut)
// 	flagPkgUrl, _ := cmd.Flags().GetString(depPkgUrl)

// 	// Dependency name.
// 	depname := qore.StringToCammelCase(name)
// 	depnamepkg := qore.StringRemoveNonAlphabet(name)
// 	if !ignoreSuffixFlag {
// 		depnamepkg += "dep"
// 	}
// 	verboseMessage(fmt.Sprintf(
// 		"Argument: %s, DependencyName: %s, PackageName: %s",
// 		name, depname, depnamepkg,
// 	))

// 	// Create directory.
// 	if len(strings.TrimSpace(flagOut)) == 0 {
// 		dir, err := os.Getwd()
// 		if err != nil {
// 			return fmt.Errorf("failed to get working directory: %w", err)
// 		}
// 		verboseMessage(fmt.Sprintf("Out directory: %s", dir))
// 		flagOut = dir
// 	}
// 	flagOut = path.Join(flagOut, depnamepkg)
// 	if err := os.MkdirAll(flagOut, 0755); err != nil {
// 		return fmt.Errorf("failed to create dependency directory: %w", err)
// 	}
// 	verboseMessage(fmt.Sprintf("Success create %s directory", flagOut))

// 	// Generate dependency file.
// 	if err := templates.MakeDependency(templates.Dependency{
// 		PkgUrl:    flagPkgUrl,
// 		StructTag: strings.ToUpper(qore.StringToSnakeCase(depnamepkg)),
// 		DepName:   depname,
// 		PkgName:   depnamepkg,
// 	}, flagOut); err != nil {
// 		return err
// 	}
// 	verboseMessage("Qore generated dependency has been created")

// 	// Pkg URL present means package need to be as go module.
// 	if !qore.ValidationIsEmpty(flagPkgUrl) {
// 		// Change dir.
// 		origin, err := os.Getwd()
// 		if err != nil {
// 			return fmt.Errorf("failed to get current dir: %w", err)
// 		}
// 		if err := os.Chdir(flagOut); err != nil {
// 			return fmt.Errorf("failed to change dir: %w", err)
// 		}
// 		verboseMessage("Directory has been changed to target")

// 		// Go mod init.
// 		comm := exec.Command("go", "mod", "init", fmt.Sprintf("%s/%s", flagPkgUrl, depnamepkg))
// 		comm.Stdout = os.Stdout
// 		comm.Stderr = os.Stderr
// 		if err := comm.Run(); err != nil {
// 			return fmt.Errorf("failed to initiate dependency with go mod init command: %w", err)
// 		}
// 		verboseMessage("Qore generated dependency has been initiated")

// 		// Go back.
// 		if err := os.Chdir(origin); err != nil {
// 			return fmt.Errorf("failed to change dir: %w", err)
// 		}
// 		verboseMessage("Directory has been changed to origin")
// 		return nil
// 	}

// 	return nil
// }
