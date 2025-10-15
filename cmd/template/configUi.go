package template

import (
	"errors"
	"fmt"
	"log/slog"
	"os"
	"strconv"

	"github.com/charmbracelet/huh"
)

var _dummyGroup = huh.NewGroup(huh.NewNote().Title("Dummy"))

type _fieldWrapper struct {
	Visible func() (bool)
	Field huh.Field
}

func _buildGroups(groups ...*huh.Group) []*huh.Group {
	var visibleGroups []*huh.Group

	for _, g := range groups {
		if g != _dummyGroup {
			visibleGroups = append(visibleGroups, g)
		}
	}

	return visibleGroups
}

func _buildGroup(fields ..._fieldWrapper) *huh.Group {
	var visibleFields []huh.Field

	for _, f := range fields {
		if f.Visible() {
			visibleFields = append(visibleFields, f.Field)
		}
	}

	if len(visibleFields) > 0 {
		return huh.NewGroup(visibleFields...)
	}
	return _dummyGroup
}

func openConfigUi(opts TemplateOptions) (TemplateOptions, error) {
	var lang string = opts.Lang
	var projectType string = opts.ProjectType
	var dir string = opts.Dir
	var team string
	tmp := strconv.FormatUint(opts.Team, 10)
	if tmp != "0" {
		team = tmp
	}
	var desktopSupport bool
	if opts.DesktopSupport != nil {
		desktopSupport = *opts.DesktopSupport
	}

	// Display form with selected theme.
	err := huh.NewForm(
		_buildGroups(
			_buildGroup(
				_fieldWrapper{
					Visible: func() bool { return lang == "" },
					Field: huh.NewSelect[string]().
						OptionsFunc(func() []huh.Option[string] {
							langs, err := GetLangs()
							if err != nil {
								slog.Error("Unable to get languages", "error", err)
								os.Exit(1)
							}

							opts := make([]huh.Option[string], len(langs))
							for i, e := range langs {
								opts[i] = huh.Option[string]{Value: e, Key: e}
							}
							return opts
						}, nil).
						Description("Choose your desired language").
						Value(&lang).
						Height(4),
				},
				_fieldWrapper{
					Visible: func() bool { return projectType == "" },
					Field: huh.NewSelect[string]().
						OptionsFunc(func() []huh.Option[string] {
							if lang == "" {
								return []huh.Option[string]{}
							}

							projects, err := GetProjects(lang)
							if err != nil {
								slog.Error("Unable to get project types", "error", err)
								os.Exit(1)
							}

							opts := make([]huh.Option[string], len(projects))
							for i, e := range projects {
								opts[i] = huh.Option[string]{Value: e, Key: e}
							}
							return opts
						}, &lang).
						Description("Choose your project type").
						Value(&projectType),
				},
				).Title("Project Language & Type"),

			_buildGroup(
				_fieldWrapper{
					Visible: func() bool { return dir == "" },
					Field: huh.NewInput().
						Title("Project Path").
						Placeholder("This is the full path to the project").
						Value(&dir),
				},
				_fieldWrapper{
					Visible: func() bool { return team == "" },
					Field: huh.NewInput().
						Title("Team Number").
						Placeholder("Your teams number").
						Value(&team).
						Validate(func(s string) error {
							_, err := strconv.Atoi(s)
							if err != nil {
								return errors.New("must be a number")
							}
							return nil
						}),
				},
				_fieldWrapper{
					Visible: func() bool { return opts.DesktopSupport == nil },
					Field: huh.NewConfirm().
						Title("Enable Desktop Support?").
						Value(&desktopSupport),
				},
				).Title("Project Setup Information"),
			)...,
		).Run()

	if err != nil {
	if err == huh.ErrUserAborted {
		os.Exit(130)
	}
		fmt.Println(err)
		os.Exit(1)
	}

	teamnr, _ := strconv.ParseUint(team, 10, 64)

	return TemplateOptions{
		Lang: lang,
		ProjectType: projectType,
		Dir: dir,
		Team: teamnr,
		DesktopSupport: &desktopSupport,
	}, nil
}
