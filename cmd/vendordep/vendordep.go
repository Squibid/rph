package vendordep

import (
	"encoding/json"
	"errors"
	"io"
	"io/fs"
	"log/slog"
	"rph/utils"
)

type JavaDepedency struct {
	GroupId string `json:"groupId"`
	ArtifactId string `json:"artifactId"`
	Version string `json:"version"`
}

type JniDependency struct {
	GroupId string `json:"groupId"`
	ArtifactId string `json:"artifactId"`
	Version string `json:"version"`
	IsJar bool `json:"isJar"`
	SkipInvalidPlatforms bool `json:"skipInvalidPlatforms"`
	ValidPlatforms []string `json:"validPlatforms"`
	SimMode string `json:"simMode"`
}

type CppDependency struct {
	GroupId string `json:"groupId"`
	ArtifactId string `json:"artifactId"`
	Version string `json:"version"`
	LibName string `json:"libName"`
	HeaderClassifier string `json:"headerClassifier"`
	SharedLibrary bool `json:"sharedLibrary"`
	SkipInvalidPlatforms bool `json:"skipInvalidPlatforms"`
	BinaryPlatforms []string `json:"binaryPlatforms"`
	SimMode string `json:"simMode"`
}

type Vendordep struct {
	FileName string `json:"filename"`
	Name string `json:"name"`
	Version string `json:"version"`
	FrcYear utils.StringOrNumber `json:"frcYear"`
	UUID string `json:"uuid"`
	MavenUrls []string `json:"mavenUrls"`
	JsonUrl string `json:"jsonUrl"`
	JavaDependencies []JavaDepedency `json:"javaDependencies"`
	JniDependencies []JniDependency `json:"jniDependencies"`
	CppDependencies []CppDependency `json:"cppDependencies"`
}

// Matches check if one vendor dep matches another in any real useful way
func (v *Vendordep) Matches(other Vendordep, strict bool) bool {
	if other.Name != "" && v.Name == other.Name {
		if other.Version != "" && v.Version == other.Version {
			return true
		}
	}
	if other.UUID != "" && v.UUID == other.UUID { return true }
	if !strict {
		if other.JsonUrl != "" && v.JsonUrl == other.JsonUrl { return true }
	}

	return false
}

func Parse(vendordepFile io.Reader) (*Vendordep, error) {
	var vendordep Vendordep

	if err := json.NewDecoder(vendordepFile).Decode(&vendordep); err != nil {
		slog.Error("Invalid vendor dependency Error decoding JSON", "error", err)
		return nil, err
	}

	return &vendordep, nil
}

func ListVendorDeps(projectFs fs.FS) ([]Vendordep, error) {
	var out []Vendordep

	err := fs.WalkDir(projectFs, "vendordeps", func(path string, d fs.DirEntry,
		err error) error {
			if err != nil { return err }
			if d.IsDir() { return nil }

			file, err := projectFs.Open(path)
			if err != nil {
				slog.Error("unable to open vendor directory", "error", err)
				return err
			}
			defer file.Close()

			dep, err := Parse(file)
			if err != nil {
				slog.Error("Unable to parse vendordep file", "file", d.Name, "error", err)
				return err
			}

			out = append(out, *dep)
			return nil
		})

	if err != nil {
		slog.Error("Unable to read vendordep fs", "error", err)
		return nil, err
	}

	return out, nil
}

func FindVendorDepFromName(name string, fs fs.FS) (*Vendordep, error) {
	deps, err := ListVendorDeps(fs);
	if err != nil {
		slog.Error("Unable to find vendor deps", "error", err)
		return nil, err
	}

	for _, dep := range deps {
		if dep.Name == name {
			return &dep, nil
		}
	}

	return nil, errors.New("Vendordep not found")
}

func ShowInfo(Vendordep) {
}
