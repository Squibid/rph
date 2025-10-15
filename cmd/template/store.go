package template

import (
	"context"
	"errors"
	"io/fs"
	"log/slog"
	"net/http"
	"os"
	"path/filepath"
	"rph/state"

	"github.com/charmbracelet/bubbles/progress"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/mholt/archives"
)

const dataFile = "templates.bin"
const zipFile = "templates.zip"

func saveArchiveVersion(version string) error {
	return os.WriteFile(
		filepath.Join(state.CachePath, dataFile),
		[]byte(version),
		0644,
	)
}

func LoadArchiveVersion() (string, error) {
	data, err := os.ReadFile(filepath.Join(state.CachePath, dataFile))
	if err != nil {
		return "", err
	}
	return string(data), nil
}

func downloadFile(url string) error {
	var progressBar = true

	// Create the file
	out, err := os.Create(filepath.Join(state.CachePath, zipFile))
	if err != nil {
		return err
	}
	defer out.Close()

	// Get the data
	resp, err := http.Get(url)
	if err != nil {
		return err
	}
	defer resp.Body.Close()


	if resp.ContentLength <= 0 {
		slog.Warn("Can't parse content length, no progress bar will be shown.")
		progressBar = false
	}

	var p *tea.Program
	pw := &progressWriter{
		total:  int(resp.ContentLength),
		file:   out,
		reader: resp.Body,
		onProgress: func(ratio float64) {
			p.Send(progressMsg(ratio))
		},
	}

	m := downloadModel{
		pw:       pw,
		progress: progress.New(progress.WithDefaultGradient()),
	}
	p = tea.NewProgram(m)

	// start the download
	go pw.Start()

	if progressBar {
		if _, err := p.Run(); err != nil {
			slog.Error("Error starting the progress bar", "error", err)
		}
	}

	return err
}

func OpenArchive(ctx context.Context) (fsys fs.FS, err error) {
	fsys, err = archives.FileSystem(ctx, filepath.Join(state.CachePath, zipFile), nil)
	if err != nil {
		return nil, err
	}

	return fsys, nil
}

func GetLangs() ([]string, error) {
	var langs []string
	fsys, err := OpenArchive(context.Background())
	if err != nil {
		slog.Error("Failed to open archive", "err", err)
		return nil, err
	}

	entries, err := fs.ReadDir(fsys, ".")
	if err != nil {
		slog.Error("No files found in fsys", "err", err)
		return nil, err
	}

	for _, entry := range entries {
		if entry.IsDir() {
			langs = append(langs, entry.Name())
		}
	}

	return langs, nil
}

func GetProjects(lang string) ([]string, error) {
	var projects []string
	if lang == "" {
		return nil, errors.New("lang must be set")
	}

	fsys, err := OpenArchive(context.Background())
	if err != nil {
		slog.Error("Failed to open archive", "err", err)
		return nil, err
	}

	entries, err := fs.ReadDir(fsys, filepath.Join(".", lang))
	if err != nil {
		slog.Error("No files found in fsys", "err", err)
		return nil, err
	}

	for _, entry := range entries {
		if entry.IsDir() {
			projects = append(projects, entry.Name())
		}
	}

	return projects, nil
}
