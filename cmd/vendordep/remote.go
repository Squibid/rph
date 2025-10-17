package vendordep

import (
	"errors"
	"io/fs"
	"log/slog"
	"regexp"
	"rph/cmd/vendordep/artifactory"
	"time"
)

type OnlineVendordep struct {
	VendordepName string
	Version string
	FileName string
	LastModTime time.Time
}

func ListAvailableOnlineDeps(year string) (map[string][]OnlineVendordep, error) {
	fsys := artifactory.New(artifactory.DefaultVendorDepArtifactoryUrl)
	path := "vendordeps/vendordep-marketplace/" + year

	entries, err := fs.ReadDir(fsys, path)
	if err != nil {
		slog.Error("Failed to readdir from artifactory", "dir", path, "error", err)
		return nil, err
	}

	type fileEntry struct {
		Name string
		ModTime time.Time
	}

	files := make([]fileEntry, len(entries))
	for i, e := range entries {
		info, err := e.Info()
		if err != nil {
			return nil, err
		}

		files[i] = fileEntry{ Name: e.Name(), ModTime: info.ModTime() }
	}

	allDeps := make(map[string][]OnlineVendordep, len(entries))

	// breaks a vendordep file name into "name-of-library" and "v2025.9.28"
	re := regexp.MustCompile(`^(.+)-v?(\d+\.\d+(?:\.\d+)?)`)
	for _, file := range files {
		matches := re.FindStringSubmatch(file.Name)
		if len(matches) > 2 {
			baseName := matches[1]
			version := matches[2]

			allDeps[baseName] = append(allDeps[baseName], OnlineVendordep{
				VendordepName: baseName,
				Version: version,
				FileName: file.Name,
				LastModTime: file.ModTime,
			})
		} else {
			return nil, errors.New("Vendordep file name does not match format '" + file.Name + "'")
		}
	}

	return allDeps, nil
}
