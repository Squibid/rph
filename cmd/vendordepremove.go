package cmd

import (
	"log/slog"
	"os"
	"path/filepath"
	"rph/cmd/vendordep"

	"github.com/spf13/cobra"
)

// vendordepremoveCmd represents the vendordep remove command
var vendordepremoveCmd = &cobra.Command{
	Use: "remove",
	Short: "Remove a vendordep",
	Long: `Remove a vendordep, by default this will move the vendordep file into
a safe place just incase you wish to undo this action to make sure it's been
removed from your disk you may use the -f flag.`,
	Aliases: []string{ "rm" },
	Args: cobra.MinimumNArgs(1),
	ValidArgsFunction: vendorDepsComp,
	RunE: func(cmd *cobra.Command, args []string) error {
		if !inProjectDir() { return nil }

		force, err := cmd.Flags().GetBool("force")
		if err != nil { return err }

		for _, n := range args {
			dep, err := vendordep.FindVendorDepFromName(args[0], projectFs)
			if err != nil {
				slog.Error("Failed to get vendor dep from name", "name", n, "error", err)
				return err
			}

			if force {
				err = os.Remove(filepath.Join(projectDir, "vendordeps", dep.FileName))
				if err != nil {
					slog.Error("Failed to remove vendor dep", "error", err)
					return err
				}
			} else {
				err = vendordep.Trash(filepath.Join(projectDir, "vendordeps", dep.FileName))
				if err != nil {
					slog.Error("Failed to move vendor dep to trash", "error", err)
					return err
				}
			}
		}

		return nil
	},
}

func init() {
	vendordepCmd.AddCommand(vendordepremoveCmd)

	vendordepremoveCmd.Flags().BoolP("force", "f", false, "Forcefully remove a vendor dependency.")
}
