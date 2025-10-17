package template

import (
	"context"
	"encoding/json"
	"io/fs"
	"log/slog"
	"os"
	"path/filepath"
	"regexp"
	"rph/state"
)

type TemplateOptions struct {
	Lang string
	ProjectType string
	Dir string
	Team uint64
	DesktopSupport *bool
}

type wpilibPreferences struct {
	CppIntellisense bool `json:"enableCppIntellisense"`
	Lang string `json:"currentLanguage"`
	Year string `json:"projectYear"`
	Team int `json:"teamNumber"`
}

// Fetch fetch the latest template zip that's distributed by vscode-wpilib
func Fetch(force bool, version string) {
	os.MkdirAll(filepath.Join(state.CachePath), 0755)
	getTemplateArchive(zipFile, force, version);
}

func GenerateProject(opts TemplateOptions) {
	opts, err := openConfigUi(opts)
	if err != nil {
		slog.Error("Failed to run interactive ui", "error", err)
		return
	}

	err = os.Mkdir(opts.Dir, 0755)
	if err != nil {
		slog.Error("Failed to create directory", "path", opts.Dir, "error", err)
		// TODO: how should we handle if the directory already exists?
	}

	fsys, err := OpenArchive(context.Background())
	if err != nil {
		slog.Error("Failed to open archive", "err", err)
		return
	}

	subFS, err := fs.Sub(fsys, filepath.Join(opts.Lang, opts.ProjectType))
	if err != nil {
		slog.Error("Unable to find project template", "template", opts.ProjectType, "error", err)
	}

	err = os.CopyFS(opts.Dir, subFS)
	if err != nil {
		slog.Error("Failed to copy template to destination", "template",
			opts.ProjectType, "destination", opts.Dir, "error", err)
		os.Exit(1)
	}

	// Configure the project

	err = os.Chmod(filepath.Join(opts.Dir, "gradlew"), 0755)
	if err != nil {
		slog.Error("Unable to make gradlew executable", "error", err)
	}

	{
		jsonFile := filepath.Join(opts.Dir, ".wpilib", "wpilib_preferences.json")
		file, err := os.Open(jsonFile)
		if err != nil {
			slog.Error(
				"Failed to open wpilib preferences file\n\nYou need to put your team number into " +
				opts.Dir + "/.wpilib/wpilib_preferences.json",
				"error",
				err,
				)
			goto PostJson
		}
		defer file.Close()

		{ // scope it so I can jump over it
			var p wpilibPreferences
			decoder := json.NewDecoder(file)
			err = decoder.Decode(&p)
			if err != nil {
				slog.Error("Failed to decode json", "error", err)
				goto PostJson
			}

			p.Team = int(opts.Team)

			jsonData, err := json.MarshalIndent(p, "", "  ")
			if err != nil {
				slog.Error("Error marshaling JSON:", "error", err)
				goto PostJson
			}

			err = os.WriteFile(jsonFile, jsonData, 0644)
			if err != nil {
				slog.Error("Error writing to file:", "error", err)
				goto PostJson
			}
		}

		PostJson:
	}

	// Enable desktop support
	{
		buildGradleFile := filepath.Join(opts.Dir, "build.gradle")
    in, err := os.ReadFile(buildGradleFile)
		if err != nil {
			slog.Error("Failed to open gradle build file", "error", err)
			goto PostGradle
		}

		{ // scope it so I can jump over it
			value := "false"
			if *opts.DesktopSupport {
				value = "true"
			}

			// this regex was translated from:
			// https://github.com/wpilibsuite/vscode-wpilib/blob/df7fc8bb9db453cbc9ccc32d3c5f81ef53f5e93a/vscode-wpilib/src/shared/generator.ts#L390
			re := regexp.MustCompile(`(?m)^(\s*def\s+includeDesktopSupport\s*=\s*)(true|false)\b`)
			out := re.ReplaceAllString(string(in), "${1}" + value)

			err = os.WriteFile(buildGradleFile, []byte(out), 0644)
			if err != nil {
				slog.Error("failed to write to gradle build file", "error", err)
				goto PostGradle
			}
		}

		PostGradle:
	}

	// Generate example deploy file
	{
		deployPath := filepath.Join(opts.Dir, "src", "main", "deploy")
		err := os.Mkdir(deployPath, 0755)
		if err != nil {
			slog.Error("Failed to make deploy directory", "error", err)
			goto PostExampleDeploy
		}

		{ // Scope it so we can jump over it
			var exampleTxt string
			switch (opts.Lang) {
			case "cpp":
				exampleTxt = `Files placed in this directory will be deployed to the RoboRIO into the
				'deploy' directory in the home folder. Use the 'frc::filesystem::GetDeployDirectory'
				function from the 'frc/Filesystem.h' header to get a proper path relative to the deploy
				directory.`
			case "java":
				exampleTxt = `Files placed in this directory will be deployed to the RoboRIO into the
				'deploy' directory in the home folder. Use the 'Filesystem.getDeployDirectory' wpilib function
				to get a proper path relative to the deploy directory.`
			default:
				exampleTxt = `Files placed in this directory will be deployed to the RoboRIO into the
				'deploy' directory in the home folder.`
			}

			os.WriteFile(deployPath, []byte(exampleTxt), 0644)
		}
		PostExampleDeploy:
	}
}
