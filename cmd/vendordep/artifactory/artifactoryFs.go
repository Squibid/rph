package artifactory

import (
	"bytes"
	"encoding/json"
	"errors"
	"io"
	"io/fs"
	"net/http"
	"path"
	"strings"
	"time"
)

type ArtifactoryFS struct {
	BaseURL string
	Client *http.Client
}

func New(baseURL string) ArtifactoryFS {
	return ArtifactoryFS{
		BaseURL: strings.TrimRight(baseURL, "/") + "/",
		Client: &http.Client{Timeout: 10 * time.Second},
	}
}

func (afs ArtifactoryFS) GetUrl(name string) string {
	cleanName := path.Clean(name)
	return afs.BaseURL + cleanName
}

func (afs ArtifactoryFS) Open(name string) (fs.File, error) {
	cleanName := path.Clean(name)
	url := afs.BaseURL + cleanName

	// Fetch metadata via storage API
	metaURL := afs.BaseURL + "api/storage/" + cleanName
	if cleanName == "." {
		metaURL = afs.BaseURL + "api/storage/"
	}

	resp, err := afs.Client.Get(metaURL)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	if resp.StatusCode == 404 {
		return nil, fs.ErrNotExist
	} else if resp.StatusCode != 200 {
		return nil, errors.New("unexpected status: " + resp.Status)
	}

	var meta struct {
		Repo string `json:"repo"`
		Path string `json:"path"`
		Children []struct {
			URI string `json:"uri"`
			Folder bool `json:"folder"`
		} `json:"children"`
		Size string `json:"size"`
	}

	if err := json.NewDecoder(resp.Body).Decode(&meta); err != nil {
		return nil, err
	}

	// It's a directory
	if len(meta.Children) > 0 {
		return &artifactoryDir{
			entries: meta.Children,
			pos: 0,
			name: cleanName,
		}, nil
	}

	// It's a file
	contentResp, err := afs.Client.Get(url)
	if err != nil {
		return nil, err
	}

	if contentResp.StatusCode == 404 {
		return nil, fs.ErrNotExist
	} else if contentResp.StatusCode != 200 {
		return nil, errors.New("unexpected status: " + contentResp.Status)
	}

	data, err := io.ReadAll(contentResp.Body)
	contentResp.Body.Close()
	if err != nil {
		return nil, err
	}

	return &artifactoryFile{
		data: bytes.NewReader(data),
		name: cleanName,
		size: int64(len(data)),
	}, nil
}

type artifactoryFile struct {
	data *bytes.Reader
	name string
	size int64
}

func (f *artifactoryFile) Stat() (fs.FileInfo, error) {
	return &fileInfo{
		name: path.Base(f.name),
		size: f.size,
		mode: 0444,
	}, nil
}

func (f *artifactoryFile) Read(p []byte) (int, error) {
	return f.data.Read(p)
}

func (f *artifactoryFile) Close() error {
	return nil
}

type artifactoryDir struct {
	entries []struct {
		URI    string `json:"uri"`
		Folder bool   `json:"folder"`
	}
	pos  int
	name string
}

func (d *artifactoryDir) Stat() (fs.FileInfo, error) {
	return &fileInfo{
		name: path.Base(d.name),
		mode: fs.ModeDir | 0555,
	}, nil
}

func (d *artifactoryDir) Read([]byte) (int, error) {
	return 0, errors.New("cannot read directory")
}

func (d *artifactoryDir) Close() error {
	return nil
}

func (d *artifactoryDir) ReadDir(n int) ([]fs.DirEntry, error) {
	if d.pos >= len(d.entries) && n > 0 {
		return nil, io.EOF
	}

	var entries []fs.DirEntry
	max := len(d.entries)
	if n > 0 && d.pos+n < max {
		max = d.pos + n
	}

	for ; d.pos < max; d.pos++ {
		e := d.entries[d.pos]
		entries = append(entries, &dirEntry{
			name: strings.TrimPrefix(e.URI, "/"),
			isDir: e.Folder,
		})
	}

	return entries, nil
}

type fileInfo struct {
	name string
	size int64
	mode fs.FileMode
}

func (fi *fileInfo) Name() string { return fi.name }
func (fi *fileInfo) Size() int64 { return fi.size }
func (fi *fileInfo) Mode() fs.FileMode { return fi.mode }
func (fi *fileInfo) ModTime() time.Time { return time.Time{} }
func (fi *fileInfo) IsDir() bool { return fi.mode.IsDir() }
func (fi *fileInfo) Sys() any { return nil }

type dirEntry struct {
	name  string
	isDir bool
}

func (de *dirEntry) Name() string { return de.name }
func (de *dirEntry) IsDir() bool { return de.isDir }
func (de *dirEntry) Type() fs.FileMode {
	if de.isDir {
		return fs.ModeDir
	} else {
		return 0
	}
}
func (de *dirEntry) Info() (fs.FileInfo, error) {
	return &fileInfo{
		name: de.name,
		mode: de.Type(),
	}, nil
}
