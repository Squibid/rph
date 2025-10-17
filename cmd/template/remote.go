package template

import (
	"encoding/json"
	"io"
	"log/slog"
	"net/http"
	"os"
	"path/filepath"
	"strconv"

	"rph/state"
	"rph/utils"
)

type asset struct {
	Name               string `json:"name"`
	BrowserDownloadURL string `json:"browser_download_url"`
}

type release struct {
	TagName string  `json:"tag_name"`
	Assets  []asset `json:"assets"`
}

func getTemplateArchive(filename string, force bool, version string) {
	const url = "https://api.github.com/repos/wpilibsuite/vscode-wpilib/releases/"
	path := filepath.Join(state.CachePath, filename)

	currentVersion, err := LoadArchiveVersion()
	// default the version to the latest version if no version is currently
	// installed and the user wants to keep the current version
	if err != nil && version == "keep" {
		version = "latest"
	} else if currentVersion != "" && version == "keep" {
		version = currentVersion
	}

	// use tags to select the version when we're not just getting the latest one
	if version != "latest" {
		version = "tags/" + version
	}

	resp, err := http.Get(url + version)
	if err != nil {
		slog.Error("Error fetching release:", "error", err)
		os.Exit(1)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		body, _ := io.ReadAll(resp.Body)
		slog.Error("GitHub API error", "status", resp.Status, "body", string(body))
		os.Exit(1)
	}

	var release release
	if err := json.NewDecoder(resp.Body).Decode(&release); err != nil {
		slog.Error("Error decoding JSON", "error", err)
		os.Exit(1)
	}

	_, ferr := os.Stat(path)
	currentVersion, err = LoadArchiveVersion()
	if !force && err == nil && currentVersion == release.TagName && ferr == nil {
		slog.Info("Template archive is already installed", "version", currentVersion)
		slog.Info("If you would like to install a different version try: rph template fetch -h")
		return
	}

	var downloadURL string
	for _, asset := range release.Assets {
		if asset.Name == "templates.zip" {
			downloadURL = asset.BrowserDownloadURL
			break
		}
	}

	if downloadURL == "" {
		slog.Warn("templates.zip not found in release version.", "version", version)
		return
	}

	err = utils.DownloadFile(downloadURL, filepath.Join(state.CachePath, zipFile))
	if err != nil {
		slog.Error("Error downloading archive file", "error", err)
		os.Exit(1)
	}

	err = saveArchiveVersion(release.TagName)
	if err != nil {
		slog.Warn("Failed to save version information", "error", err)
	} else {
		slog.Info("Downloaded new template file", "version", release.TagName)
	}
}

func ListTemplateArchiveVersions(results uint8) []string {
	const url = "https://api.github.com/repos/wpilibsuite/vscode-wpilib/releases?per_page="
	resp, err := http.Get(url + strconv.Itoa(int(results)))
	if err != nil {
		slog.Error("Error fetching release:", "error", err)
		os.Exit(1)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		body, _ := io.ReadAll(resp.Body)
		slog.Error("GitHub API error", "status", resp.Status, "body", string(body))
		os.Exit(1)
	}

	var releases []release
	if err := json.NewDecoder(resp.Body).Decode(&releases); err != nil {
		slog.Error("Error decoding JSON", "error", err)
		os.Exit(1)
	}

	var versions []string
	for _, r := range releases {
		for _, a := range r.Assets {
			if a.Name == "templates.zip" {
				versions = append(versions, r.TagName)
			}
		}
	}

	return versions
}
