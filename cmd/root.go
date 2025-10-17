package cmd

import (
	"io/fs"
	"log/slog"
	"os"
	"path/filepath"
	"rph/state"
	"rph/utils"

	"github.com/spf13/cobra"
)

var projectFs fs.FS
var projectDir string

var rootCmd = &cobra.Command{
	Use: state.Name,
	Short: "Manage your FRC robot code the UNIX way.",
	Long: `rph (Robot Pits Helper) is a command line utility with the goal of
giving FRC teams a simple way to interact with wpilib robot code.

rph is cross platform, and should work everywhere wpilib is supported. To
actually run your robot code you will still need to install wpilib.`,
	PersistentPreRunE: func(cmd *cobra.Command, args []string) error {
		dir, err := cmd.Flags().GetString("project-dir")
		if err != nil {
			slog.Error("Unable to set projectFs or projectDir", "error", err)
			os.Exit(1)
		}

		// if the user specified a directory we'll trust them
		if dir != "." {
			projectFs = os.DirFS(dir)
			projectDir = dir
			return nil
		}

		// otherwise we should go find the .wpilib folder in the parent directories
		path, err := utils.FindEntryDirInParents(dir, ".wpilib")
		if err != nil {
			slog.Error("Unable to find project directory", "error", err)
			return err
		}

		projectFs = os.DirFS(path)
		projectDir = path
		return nil
	},
}

func Execute() {
	rootCmd.PersistentFlags().String("project-dir", ".", "Set the project directory.")

	err := rootCmd.Execute()
	if err != nil {
		os.Exit(1)
	}
}

func superPersistentPreRun(cmd *cobra.Command, args []string) {
	if parent := cmd.Parent(); parent != nil {
		if parent.PersistentPreRunE != nil {
			if err := parent.PersistentPreRunE(parent, args); err != nil {
				return
			}
		} else if parent.PersistentPreRun != nil {
			// Fallback to PersistentPreRun if PersistentPreRunE isn't set
			parent.PersistentPreRun(parent, args)
		}
	}
}

// inProjectDir handles the log message for you
func inProjectDir() bool {
	_, err := os.Stat(filepath.Join(projectDir, ".wpilib", "wpilib_preferences.json"))
	if err != nil {
		slog.Error("Are you in a project directory?")
		return false
	}

	return true
}
