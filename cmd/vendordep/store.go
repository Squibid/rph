package vendordep

import (
	"log/slog"
	"os"
	"path/filepath"
	"rph/state"
)

const vendordepDir = "vendordeps"

func MkCacheDir() {
	os.MkdirAll(filepath.Join(state.CachePath, vendordepDir), 0755);
}

func Trash(path string) error {
	_, err := os.Stat(path)
	if err != nil {
		slog.Error("Can't trash file, path must be a valid file", "path", path, "error", err)
		return err
	}

	file, err := os.Open(path)
	if err != nil {
		slog.Error("Failed to open vendordep file", "error", err)
		return err
	}
	defer file.Close()

	dep, err := Parse(file)
	if err != nil {
		slog.Error("Failed to parse vendordep file", "error", err)
	}

	err = os.Rename(path, filepath.Join(state.CachePath, vendordepDir, dep.FileName))
	if err != nil {
		slog.Info("Failed to move vendor dep", "error", err)
		return err
	}

	return nil
}

// hasVendorDepOnDisk check if a vendor dep is already on your disk, this is only
// useful if you've got the uuid of the vendordep you would like to install or
// are in a very percarious situation where you have no internet and any version
// of your vendordep will do.
func hasVendorDepOnDisk(dep Vendordep, strict bool) (bool, error) {
	path := filepath.Join(state.CachePath, vendordepDir)

	// by default we're not matching
	matches := false

	err := filepath.WalkDir(path, func(path string, d os.DirEntry, err error) error {
		if err != nil { return err }
		if d.IsDir() { return nil }

		file, err := os.Open(filepath.Join(path, d.Name()))
		if err != nil {
			slog.Error("Failed to open vendordep file", "error", err)
			return err
		}
		defer file.Close()

		new_dep, err := Parse(file)
		if err != nil {
			slog.Error("Failed to parse vendordep file", "error", err)
		}

		matches = new_dep.Matches(dep, strict)
		return nil
	})

	if err != nil {
		slog.Error("Failed to walk the vendordep directory", "error", err)
		return false, err
	}

	return matches, nil
}
