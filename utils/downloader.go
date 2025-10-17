package utils

import (
	"io"
	"log/slog"
	"net/http"
	"os"
	"strings"
	"time"

	"github.com/charmbracelet/bubbles/progress"
	tea "github.com/charmbracelet/bubbletea"
)

const (
	padding = 2
	maxWidth = 80
)

func DownloadFile(url string, outpath string) error {
	var progressBar = true

	// Create the file
	out, err := os.Create(outpath)
	if err != nil { return err }
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
		total: int(resp.ContentLength),
		file: out,
		reader: resp.Body,
		originUrl: url,
		onProgress: func(ratio float64) {
			p.Send(progressMsg(ratio))
		},
	}

	p = tea.NewProgram(downloadModel{
		pw: pw,
		progress: progress.New(progress.WithDefaultGradient()),
	})


	if progressBar {
		// start the download
		go pw.Start()
		if _, err := p.Run(); err != nil {
			slog.Error("Error starting the progress bar", "error", err)
		}
	} else {
		// we need to block the file and stream from closing
		pw.Start()
	}

	return err
}

type progressMsg float64
type progressErrMsg struct{ err error }

func finalPause() tea.Cmd {
	return tea.Tick(time.Millisecond*750, func(_ time.Time) tea.Msg {
		return nil
	})
}

type progressWriter struct {
	total int
	downloaded int
	file *os.File
	reader io.Reader
	originUrl string
	onProgress func(float64)
}

func (pw *progressWriter) Start() {
	// TeeReader calls pw.Write() each time a new response is received
	_, err := io.Copy(pw.file, io.TeeReader(pw.reader, pw))
	if err != nil {
		slog.Error("Error in progress writer", "error", progressErrMsg{err})
	}
}

func (pw *progressWriter) Write(p []byte) (int, error) {
	pw.downloaded += len(p)
	if pw.total > 0 && pw.onProgress != nil {
		pw.onProgress(float64(pw.downloaded) / float64(pw.total))
	}
	return len(p), nil
}

type downloadModel struct {
	pw *progressWriter
	progress progress.Model
	err error
}

func (m downloadModel) Init() tea.Cmd {
	return nil
}

func (m downloadModel) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch msg := msg.(type) {
	case tea.KeyMsg:
		return m, tea.Quit

	case tea.WindowSizeMsg:
		m.progress.Width = min(msg.Width - padding * 2 - 4, maxWidth)
		return m, nil

	case progressErrMsg:
		m.err = msg.err
		return m, tea.Quit

	case progressMsg:
		var cmds []tea.Cmd

		if msg >= 1.0 {
			cmds = append(cmds, tea.Sequence(finalPause(), tea.Quit))
		}

		cmds = append(cmds, m.progress.SetPercent(float64(msg)))
		return m, tea.Batch(cmds...)

	// FrameMsg is sent when the progress bar wants to animate itself
	case progress.FrameMsg:
		progressModel, cmd := m.progress.Update(msg)
		m.progress = progressModel.(progress.Model)
		return m, cmd

	default:
		return m, nil
	}
}

func (m downloadModel) View() string {
	if m.err != nil {
		return "Error downloading: " + m.err.Error() + "\n"
	}

	pad := strings.Repeat(" ", padding)
	return "\n" + pad + "Downloading " + m.pw.originUrl + " to " + m.pw.file.Name() + "\n\n" + pad + m.progress.View() + "\n\n" + pad + "Press any key to quit"
}
