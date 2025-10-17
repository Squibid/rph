package cmd

import (
	"log/slog"
	"rph/cmd/vendordep"
	"strings"

	"github.com/spf13/cobra"
)

func vendorDepsComp(cmd *cobra.Command, args []string, toComplete string) ([]cobra.Completion, cobra.ShellCompDirective) {
	validVendordeps, err := vendordep.ListVendorDeps(projectFs)
	if err != nil {
		slog.Error("Unable to find vendor deps", "error", err)
		return nil, cobra.ShellCompDirectiveNoFileComp
	}

	var completions []string
	for _, dep := range validVendordeps {
		if strings.HasPrefix(dep.Name, toComplete) {
			completions = append(completions, dep.Name)
		}
	}

	return completions, cobra.ShellCompDirectiveNoFileComp
}

// vendordepCmd represents the vendordep command
var vendordepCmd = &cobra.Command{
	Use: "vendordep",
	Aliases: []string{ "vend" },
	Short: "Mange your WPILIB projects vendordeps",
	Long: `Mange your WPILIB projects vendordeps`,
	PersistentPreRun: func(cmd *cobra.Command, args []string) {
		superPersistentPreRun(cmd, args)
		vendordep.MkCacheDir()
	},
}

func init() {
	rootCmd.AddCommand(vendordepCmd)
}
