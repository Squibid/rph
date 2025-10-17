package utils

import (
	"fmt"
	"log/slog"
	"os"
	"path/filepath"
	"strings"
)

func FindEntryDirInParents(startPath string, lookingFor string) (string, error) {
	absPath, err := filepath.Abs(startPath)
	if err != nil {
		slog.Error("Unable to get the absolute path", "path", startPath, "error", err)
		return "", err
	}

	// Find our depth so we don't have to do more checking than necessary
	components := strings.Split(filepath.Clean(absPath), string(os.PathSeparator))

	// Start searching from the current directory upwards
	currentPath := absPath
	for i := len(components); i > 0; i-- {
		lookingForPath := filepath.Join(currentPath, lookingFor)

		_, err := os.Stat(lookingForPath)
		if err == nil {
			realLookingForPath, err := filepath.Abs(currentPath)
			if err != nil {
				slog.Error("Unable to get the absolute path", "path", lookingForPath, "error", err)
				return "", err
			}
			return realLookingForPath, nil
		}

		// Go up one directory level
		currentPath = filepath.Dir(currentPath)
	}

	return "", fmt.Errorf(lookingFor + " not found")
}
