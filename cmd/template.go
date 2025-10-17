package cmd

import (
	"fmt"
	"log/slog"
	"rph/cmd/template"
	"rph/utils"
	"slices"

	"github.com/spf13/cobra"
)

var desktopSupportFlag utils.BoolFlag

// templateCmd represents the template command
var templateCmd = &cobra.Command{
	Use: "template",
	Short: "Generate a new WPILIB project from a template.",
	Long: `Generates a new WPILib robot project from a template archive.
You can pass flags or leave them out to be prompted interactively.

If you wish to skip the interactive ui then you must pass all of your
options in using the following flags:
--lang, --type, --dir, --team, --desktopSupport

Example:
rph template --lang=java --type=commandbased --dir=MyRobot --team=5438 --desktopSupport=false`,

	PersistentPreRun: func(cmd *cobra.Command, args []string) {
		// This is a noop to stop the root command from preventing us from making
		// a new robot project
	},
	PreRun: func(cmd *cobra.Command, args []string) {
		template.Fetch(false, "keep")
	},
	RunE: func(cmd *cobra.Command, args []string) error {
		lang, err := cmd.Flags().GetString("lang")
		if err != nil { return err }
		projectType, err := cmd.Flags().GetString("type")
		if err != nil { return err }
		types, err := cmd.Flags().GetBool("types")
		if err != nil { return err }
		dir, err := cmd.Flags().GetString("dir")
		if err != nil { return err }
		team, err := cmd.Flags().GetUint64("team")
		if err != nil { return err }

		// by default desktopSupport is nil to allow the interactive ui to show
		var desktopSupport *bool
		if desktopSupportFlag.IsSet {
			desktopSupport = &desktopSupportFlag.Value
		}

		var langs []string
		var projectTypes []string

		if types || lang != "" || projectType != "" {
			langs, err = template.GetLangs();
			if err != nil {
				slog.Error("Unable to get langs", "error", err)
				return err
			}

			if lang != "" {
				projectTypes, err = template.GetProjects(lang);
				if err != nil {
					slog.Error("Unable to get project types", "error", err)
					return err
				}
			}
		}

		if types {
			if lang != "" {
				for _, e := range projectTypes {
					fmt.Println(e)
				}
			} else {
				for _, e := range langs {
					fmt.Println(e)
				}
			}
		}

		// ensure that lang and projectType are valid
		if lang != "" {
			if !slices.Contains(langs, lang) {
				slog.Error("Language is not valid", "language", lang)
				return nil
			}
		}
		if projectType != "" {
			if !slices.Contains(projectTypes, projectType) {
				slog.Error("Project type is not valid", "type", projectType)
				return nil
			}
		}

		template.GenerateProject(template.TemplateOptions{
			Lang: lang,
			ProjectType: projectType,
			Dir: dir,
			Team: team,
			DesktopSupport: desktopSupport,
		})

		return err
	},
}

func init() {
	rootCmd.AddCommand(templateCmd)
	templateCmd.Flags().StringP("lang", "l", "", "The language of the project")
	templateCmd.Flags().StringP("type", "t", "", "The type of the project")
	templateCmd.Flags().Bool("types", false, "List the languages available or if lang is specified the types of projects for that lang")
	templateCmd.Flags().StringP("dir", "d", "", "The directory which will contain the contents of your new project")
	templateCmd.Flags().Uint64P("team", "n", 0, "Your team number")
	templateCmd.Flags().VarP(&desktopSupportFlag, "desktopSupport", "s", "Enable desktop simulation support")
}
