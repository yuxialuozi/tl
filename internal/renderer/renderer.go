package renderer

import (
	"strings"
	"time"

	"github.com/charmbracelet/lipgloss"

	"github.com/yuxialuozi/tl/internal/translator"
)

var (
	headerStyle  = lipgloss.NewStyle().Bold(true)
	headingStyle = lipgloss.NewStyle().Bold(true).Foreground(lipgloss.Color("39"))
)

// Renderer formats query and configuration output.
type Renderer struct {
	color bool
}

// ConfigStatus contains redacted configuration details safe for terminal output.
type ConfigStatus struct {
	Path         string
	FileExists   bool
	BaseURL      string
	Model        string
	APIKeyStatus string
	Timeout      time.Duration
	Issues       []string
}

// New creates a renderer. Color is enabled only when supported by the output.
func New(color bool) *Renderer {
	return &Renderer{color: color}
}

// Vocabulary renders a dictionary entry.
func (r *Renderer) Vocabulary(v translator.Vocabulary) string {
	headerParts := []string{v.Word, v.PartOfSpeech}
	if v.Category != "" {
		headerParts = append(headerParts, v.Category)
	}
	header := strings.Join(headerParts, " · ")

	var builder strings.Builder
	builder.WriteString(r.style(headerStyle, header))
	builder.WriteString("\n\n")
	builder.WriteString(r.style(headingStyle, "Definition"))
	builder.WriteString("\n")
	builder.WriteString(v.Definition)
	builder.WriteString("\n\n")
	builder.WriteString(r.style(headingStyle, "中文"))
	builder.WriteString("\n")
	builder.WriteString(v.Chinese)
	builder.WriteString("\n\n")
	builder.WriteString(r.style(headingStyle, "Examples"))
	builder.WriteString("\n")
	for _, example := range v.Examples {
		builder.WriteString("- ")
		builder.WriteString(example)
		builder.WriteString("\n")
	}
	return builder.String()
}

// Translation renders the original text and its translation.
func (r *Renderer) Translation(t translator.Translation) string {
	var builder strings.Builder
	builder.WriteString(r.style(headingStyle, "Original"))
	builder.WriteString("\n")
	builder.WriteString(t.Original)
	builder.WriteString("\n\n")
	builder.WriteString(r.style(headingStyle, "Translation"))
	builder.WriteString("\n")
	builder.WriteString(t.Translation)
	builder.WriteString("\n")
	return builder.String()
}

// Config renders safe configuration metadata without the API key.
func (r *Renderer) Config(status ConfigStatus) string {
	fileStatus := "not found"
	if status.FileExists {
		fileStatus = "found"
	}
	resultStatus := "ready"
	if len(status.Issues) > 0 {
		resultStatus = "incomplete"
	}

	var builder strings.Builder
	builder.WriteString(r.style(headerStyle, "Configuration"))
	builder.WriteString("\n")
	builder.WriteString("Path: ")
	builder.WriteString(status.Path)
	builder.WriteString("\n")
	builder.WriteString("File: ")
	builder.WriteString(fileStatus)
	builder.WriteString("\n")
	builder.WriteString("Base URL: ")
	builder.WriteString(status.BaseURL)
	builder.WriteString("\n")
	builder.WriteString("Model: ")
	builder.WriteString(displayValue(status.Model))
	builder.WriteString("\n")
	builder.WriteString("API key: ")
	builder.WriteString(status.APIKeyStatus)
	builder.WriteString("\n")
	builder.WriteString("Timeout: ")
	builder.WriteString(status.Timeout.String())
	builder.WriteString("\n")
	builder.WriteString("Status: ")
	builder.WriteString(resultStatus)
	builder.WriteString("\n")
	for _, issue := range status.Issues {
		builder.WriteString("Issue: ")
		builder.WriteString(issue)
		builder.WriteString("\n")
	}
	return builder.String()
}

func (r *Renderer) style(style lipgloss.Style, text string) string {
	if !r.color {
		return text
	}
	return style.Render(text)
}

func displayValue(value string) string {
	if strings.TrimSpace(value) == "" {
		return "(not set)"
	}
	return value
}
