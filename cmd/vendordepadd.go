package cmd

import (
	"encoding/json"
	"log/slog"
	"net/http"
	"os"
	"path/filepath"
	"rph/cmd/template"
	"rph/cmd/vendordep"
	"rph/cmd/vendordep/artifactory"
	"rph/utils"
	"strings"

	"github.com/spf13/cobra"
)

// vendordepaddCmd represents the vendordep add command
var vendordepaddCmd = &cobra.Command{
	Use: "add",
	Short: "Add a new vendordep",
	Long: `Add a new vendordep. You may pass in as many urls or vendordep names
as you wish. The vendordep names are determined by what's found at
https://frcmaven.wpi.edu/ui/native/vendordeps/`,
	Args: cobra.MinimumNArgs(1),
	ValidArgsFunction: func(cmd *cobra.Command, args []string, toComplete string) ([]cobra.Completion, cobra.ShellCompDirective) {
		year, err := cmd.Flags().GetString("year")
		if err != nil {
			return nil, cobra.ShellCompDirectiveNoFileComp
		}

		if year == "" {
			file, err := os.Open(filepath.Join(projectDir, ".wpilib", "wpilib_preferences.json"))
			if err != nil {
				return nil, cobra.ShellCompDirectiveNoFileComp
			}
			defer file.Close()

			var wpilibPrefs template.WpilibPreferences
			if err := json.NewDecoder(file).Decode(&wpilibPrefs); err != nil {
				return nil, cobra.ShellCompDirectiveNoFileComp
			}

			year = wpilibPrefs.Year
		}

		// TODO: refactor this into it's own func, cache it and then we can make
		// less api calls
		validVendordeps, err := vendordep.ListAvailableOnlineDeps(year)
		if err != nil {
			return nil, cobra.ShellCompDirectiveNoFileComp
		}

		var completions []string
		for k, deps := range validVendordeps {
			for _, dep := range deps {
				comp := k + "-" + dep.Version
				if strings.HasPrefix(comp, toComplete) {
					completions = append(completions, comp)
				}
			}
		}

		return completions, cobra.ShellCompDirectiveNoFileComp
	},
	RunE: func(cmd *cobra.Command, args []string) error {
		if !inProjectDir() { return nil }

		year, err := cmd.Flags().GetString("year")
		if err != nil { return err }

		if year == "" {
			file, err := os.Open(filepath.Join(projectDir, ".wpilib", "wpilib_preferences.json"))
			if err != nil {
				slog.Error("Failed to open wpilib_preferences.json", "error", err)
				return err
			}
			defer file.Close()

			var wpilibPrefs template.WpilibPreferences
			if err := json.NewDecoder(file).Decode(&wpilibPrefs); err != nil {
				slog.Error("Failed to decode wpilib_preferences.json", "error", err)
				return err
			}

			year = wpilibPrefs.Year
		}

		fsys := artifactory.New(artifactory.DefaultVendorDepArtifactoryUrl)
		path := "vendordeps/vendordep-marketplace/" + year

		// make sure the vendordep directory exists in the current project
		os.MkdirAll(filepath.Join(projectDir, "vendordeps"), 0755);

		for _, arg := range args {
			if strings.HasPrefix(arg, "http") {
				resp, err := http.Get(arg)
				if err != nil {
					slog.Error("Failed to download file", "url", arg, "error", err)
					return err
				}
				defer resp.Body.Close()

				dep, err := vendordep.Parse(resp.Body)
				if err != nil {
					slog.Error("Failed to parse vendor dep", "error", err)
					return err
				}

				err = utils.DownloadFile(arg, filepath.Join(projectDir, "vendordeps", dep.FileName))
				if err != nil {
					slog.Error("Failed to download vendordep", "error", err)
				}
			} else {
				vendordeps, err := vendordep.ListAvailableOnlineDeps(year)
				if err != nil {
					return err
				}

				for k, deps := range vendordeps {
					for _, dep := range deps {
						d := k + "-" + dep.Version
						if d == arg {
							url := fsys.GetUrl(path + "/" + dep.FileName)
							err := utils.DownloadFile(url, filepath.Join(projectDir, "vendordeps", dep.FileName))
							if err != nil {
								slog.Error("Failed to copy the file to the filesystem", "error", err)
								return err
							}

							break
						}
					}
				}
			}
		}

		// TODO: tell the user to gradle build

		return nil
	},
}

func init() {
	vendordepCmd.AddCommand(vendordepaddCmd)
	vendordepaddCmd.Flags().StringP("year", "y", "", "override the year to search for dependencies in frcmaven.")
}
