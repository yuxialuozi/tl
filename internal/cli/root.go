package cli

import (
	"context"
	"fmt"
	"io"
	"os"
	"strings"

	"github.com/spf13/cobra"

	"github.com/yuxialuozi/tl/internal/config"
	"github.com/yuxialuozi/tl/internal/llm"
	"github.com/yuxialuozi/tl/internal/renderer"
	"github.com/yuxialuozi/tl/internal/translator"
)

// New builds the root command.
func New(version string, stdout, stderr io.Writer) *cobra.Command {
	root := &cobra.Command{
		Use:   "tl [word or text]",
		Short: "A small terminal tool for English vocabulary and translation",
		Long: `tl looks up one English word or translates English and Chinese text.

English input is translated to Chinese. Chinese input is translated to English.

Commands:
  tl config    Show configuration path and status
  tl version   Print the tl version`,
		Example: `  tl keyring
  tl "The daemon runs in the background."
  tl config
  tl version`,
		Args:          cobra.ArbitraryArgs,
		SilenceUsage:  true,
		SilenceErrors: true,
		RunE: func(cmd *cobra.Command, args []string) error {
			if len(args) == 0 {
				return cmd.Help()
			}
			if args[0] == "config" || args[0] == "version" {
				if len(args) != 1 {
					return fmt.Errorf("%s does not accept arguments", args[0])
				}
				if args[0] == "config" {
					return showConfig(stdout)
				}
				_, err := fmt.Fprintf(stdout, "tl %s\n", version)
				return err
			}
			if len(args) != 1 {
				return fmt.Errorf("expected one word or sentence; quote text containing spaces")
			}
			return runQuery(cmd.Context(), args[0], stdout)
		},
	}
	root.SetOut(stdout)
	root.SetErr(stderr)
	root.CompletionOptions.DisableDefaultCmd = true

	return root
}

// Execute runs the CLI with cancellation support.
func Execute(ctx context.Context, version string, stdout, stderr io.Writer) error {
	return New(version, stdout, stderr).ExecuteContext(ctx)
}

func runQuery(ctx context.Context, input string, stdout io.Writer) error {
	loaded, err := config.Load()
	if err != nil {
		return err
	}
	if err := loaded.Config.Validate(); err != nil {
		return fmt.Errorf("configuration is incomplete:\n%w", err)
	}

	client := llm.New(
		loaded.Config.Provider.BaseURL,
		loaded.Config.Provider.APIKey,
		loaded.Config.Provider.Model,
		loaded.Config.Timeout(),
		loaded.Config.App.EnableThinking,
	)
	service := translator.New(client)
	output, err := service.Run(ctx, input)
	if err != nil {
		return err
	}

	render := renderer.New(colorEnabled(stdout))
	switch {
	case output.Vocabulary != nil:
		_, err = fmt.Fprint(stdout, render.Vocabulary(*output.Vocabulary))
	case output.Translation != nil:
		_, err = fmt.Fprint(stdout, render.Translation(*output.Translation))
	default:
		return fmt.Errorf("internal error: no result produced")
	}
	return err
}

func showConfig(stdout io.Writer) error {
	loaded, err := config.Load()
	if err != nil {
		return err
	}

	status := renderer.ConfigStatus{
		Path:         loaded.Path,
		FileExists:   loaded.FileExists,
		BaseURL:      displayValue(loaded.Config.Provider.BaseURL),
		Model:        loaded.Config.Provider.Model,
		APIKeyStatus: apiKeyStatus(loaded.Config.Provider.APIKey, loaded.Sources.APIKey),
		Timeout:      loaded.Config.Timeout(),
	}
	if err := loaded.Config.Validate(); err != nil {
		status.Issues = strings.Split(err.Error(), "\n")
	}

	_, err = fmt.Fprint(stdout, renderer.New(colorEnabled(stdout)).Config(status))
	return err
}

func apiKeyStatus(key string, source config.Source) string {
	if strings.TrimSpace(key) == "" {
		return "not configured"
	}
	switch source {
	case config.SourceEnvironment:
		return "configured via TL_API_KEY"
	case config.SourceFile:
		return "configured in config file"
	default:
		return "not configured"
	}
}

func displayValue(value string) string {
	if strings.TrimSpace(value) == "" {
		return "(not set)"
	}
	return value
}

func colorEnabled(w io.Writer) bool {
	if os.Getenv("NO_COLOR") != "" || strings.EqualFold(os.Getenv("TERM"), "dumb") {
		return false
	}
	file, ok := w.(*os.File)
	if !ok {
		return false
	}
	info, err := file.Stat()
	if err != nil {
		return false
	}
	return info.Mode()&os.ModeCharDevice != 0
}
