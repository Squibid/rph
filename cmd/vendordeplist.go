package cmd

import (
	"fmt"
	"log/slog"
	"rph/cmd/vendordep"

	"github.com/spf13/cobra"
)

// vendordeplistCmd represents the vendordep list command
var vendordeplistCmd = &cobra.Command{
	Use: "list",
	Short: "List out your installed vendordeps",
	Long: `List out your installed vendordeps.`,
	Aliases: []string{ "ls" },
	RunE: func(cmd *cobra.Command, args []string) error {
		if !inProjectDir() { return nil }

		deps, err := vendordep.ListVendorDeps(projectFs)
		if err != nil {
			slog.Error("Unable to list vendor deps", "error", err)
			return err
		}

		for _, dep := range deps {
			fmt.Println(dep.Name)
		}

		return nil
	},
}

func init() {
	vendordepCmd.AddCommand(vendordeplistCmd)
}
