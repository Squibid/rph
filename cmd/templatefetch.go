package cmd

import (
	"fmt"
	"rph/cmd/template"

	"github.com/spf13/cobra"
)

// templatefetchCmd represents the template fetch command
var templatefetchCmd = &cobra.Command{
	Use: "fetch",
	Short: "Fetch the template archive which is used for generating templates.",
	Long: `Fetch the template archive which is used for generating templates.

Examples:
  rph template fetch -v latest # Install the latest version
  rph template fetch -v v2025.1.1 # Switch to a specific version`,

	RunE: func(cmd *cobra.Command, args []string) error {
		force, err := cmd.Flags().GetBool("force")
		if err != nil { return err }
		version, err := cmd.Flags().GetString("version")
		if err != nil { return err }
		list, err := cmd.Flags().GetUint8("list")
		if err != nil { return err }
		getversion, err := cmd.Flags().GetBool("get-version")
		if err != nil { return err }

		if getversion {
			v, err := template.LoadArchiveVersion()
			if err != nil {
				return err
			}
			fmt.Println(v)
			return nil
		}

		if list > 0 {
			versions := template.ListTemplateArchiveVersions(list)
			for _, v := range versions {
				fmt.Println(v)
			}
			return nil
		}

		template.Fetch(force, version)
		return nil
	},
}

func init() {
	templateCmd.AddCommand(templatefetchCmd)

	templatefetchCmd.Flags().Uint8P("list", "l", 0, "List all available template versions.")
	templatefetchCmd.Flags().Lookup("list").NoOptDefVal = "10"
	templatefetchCmd.Flags().BoolP("force", "f", false, "Force refetch the template archive.")
	templatefetchCmd.Flags().StringP("version", "v", "keep", "Change the version of the template archive.")
	templatefetchCmd.Flags().BoolP("get-version", "g", false, "Get the currently installed version of the template archive.")
}
