package cmd

import (
	"os"
	"rph/state"

	"github.com/spf13/cobra"
)

var rootCmd = &cobra.Command{
	Use: state.Name,
	Short: "Manage your FRC robot code the UNIX way.",
	Long: `rph (Robot Pits Helper) is a command line utility with the goal of
giving FRC teams a simple way to interact with wpilib robot code.

rph is cross platform, and should work everywhere wpilib is supported. To
actually run your robot code you will still need to install wpilib.`,
}

func Execute() {
	err := rootCmd.Execute()
	if err != nil {
		os.Exit(1)
	}
}
