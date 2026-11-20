package config

import (
	"errors"
	"fmt"
	"net/url"
	"os"
	"path/filepath"
	"strings"
	"time"

	"github.com/pelletier/go-toml/v2"
)

const (
	defaultBaseURL = "https://api.openai.com/v1"
	defaultTimeout = 30
)

// Source records where a configuration value came from.
type Source string

const (
	SourceDefault     Source = "default"
	SourceFile        Source = "config file"
	SourceEnvironment Source = "environment"
)

// Provider contains the OpenAI-compatible endpoint settings.
type Provider struct {
	BaseURL string `toml:"base_url"`
	APIKey  string `toml:"api_key"`
	Model   string `toml:"model"`
}

// App contains application settings.
type App struct {
	Timeout int `toml:"timeout"`
}

// Config is the complete tl configuration.
type Config struct {
	Provider Provider `toml:"provider"`
	App      App      `toml:"app"`
}

// Sources records the effective source of provider settings.
type Sources struct {
	BaseURL Source
	APIKey  Source
	Model   Source
}

// Result is a loaded configuration and its metadata.
type Result struct {
	Config     Config
	Path       string
	FileExists bool
	Sources    Sources
}

type rawProvider struct {
	BaseURL *string `toml:"base_url"`
	APIKey  *string `toml:"api_key"`
	Model   *string `toml:"model"`
}

type rawApp struct {
	Timeout *int `toml:"timeout"`
}

type rawConfig struct {
	Provider rawProvider `toml:"provider"`
	App      rawApp      `toml:"app"`
}

// Default returns the built-in configuration.
func Default() Config {
	return Config{
		Provider: Provider{
			BaseURL: defaultBaseURL,
		},
		App: App{
			Timeout: defaultTimeout,
		},
	}
}

// DefaultPath returns the platform-specific configuration path.
func DefaultPath() (string, error) {
	dir, err := os.UserConfigDir()
	if err != nil {
		return "", fmt.Errorf("find configuration directory: %w", err)
	}
	return filepath.Join(dir, "tl", "config.toml"), nil
}

// Load reads the default configuration file and applies environment overrides.
func Load() (Result, error) {
	path, err := DefaultPath()
	if err != nil {
		return Result{}, err
	}
	return LoadFrom(path)
}

// LoadFrom reads a TOML configuration file and applies environment overrides.
// A missing file is not an error; built-in defaults remain active.
func LoadFrom(path string) (Result, error) {
	result := Result{
		Config: Default(),
		Path:   filepath.Clean(path),
		Sources: Sources{
			BaseURL: SourceDefault,
			APIKey:  SourceDefault,
			Model:   SourceDefault,
		},
	}

	data, err := os.ReadFile(path)
	switch {
	case err == nil:
		result.FileExists = true
		var raw rawConfig
		if err := toml.Unmarshal(data, &raw); err != nil {
			return Result{}, fmt.Errorf("parse configuration %s: %w", path, err)
		}
		applyFile(&result, raw)
	case errors.Is(err, os.ErrNotExist):
		// Missing files are allowed. Environment variables and defaults still apply.
	default:
		return Result{}, fmt.Errorf("read configuration %s: %w", path, err)
	}

	applyEnvironment(&result)
	result.Config.Provider.BaseURL = strings.TrimRight(strings.TrimSpace(result.Config.Provider.BaseURL), "/")
	result.Config.Provider.APIKey = strings.TrimSpace(result.Config.Provider.APIKey)
	result.Config.Provider.Model = strings.TrimSpace(result.Config.Provider.Model)

	return result, nil
}

func applyFile(result *Result, raw rawConfig) {
	if raw.Provider.BaseURL != nil {
		result.Config.Provider.BaseURL = *raw.Provider.BaseURL
		result.Sources.BaseURL = SourceFile
	}
	if raw.Provider.APIKey != nil {
		result.Config.Provider.APIKey = *raw.Provider.APIKey
		result.Sources.APIKey = SourceFile
	}
	if raw.Provider.Model != nil {
		result.Config.Provider.Model = *raw.Provider.Model
		result.Sources.Model = SourceFile
	}
	if raw.App.Timeout != nil {
		result.Config.App.Timeout = *raw.App.Timeout
	}
}

func applyEnvironment(result *Result) {
	if value := strings.TrimSpace(os.Getenv("TL_BASE_URL")); value != "" {
		result.Config.Provider.BaseURL = value
		result.Sources.BaseURL = SourceEnvironment
	}
	if value := strings.TrimSpace(os.Getenv("TL_API_KEY")); value != "" {
		result.Config.Provider.APIKey = value
		result.Sources.APIKey = SourceEnvironment
	}
	if value := strings.TrimSpace(os.Getenv("TL_MODEL")); value != "" {
		result.Config.Provider.Model = value
		result.Sources.Model = SourceEnvironment
	}
}

// Timeout returns the configured request timeout.
func (c Config) Timeout() time.Duration {
	return time.Duration(c.App.Timeout) * time.Second
}

// Validate checks whether the configuration can be used for a request.
func (c Config) Validate() error {
	var problems []error

	baseURL := strings.TrimSpace(c.Provider.BaseURL)
	if baseURL == "" {
		problems = append(problems, errors.New("base URL is not configured"))
	} else if parsed, err := url.Parse(baseURL); err != nil || (parsed.Scheme != "http" && parsed.Scheme != "https") || parsed.Host == "" {
		problems = append(problems, fmt.Errorf("base URL %q must be an absolute http(s) URL", baseURL))
	}
	if strings.TrimSpace(c.Provider.APIKey) == "" {
		problems = append(problems, errors.New("API key is not configured; set TL_API_KEY or provider.api_key"))
	}
	if strings.TrimSpace(c.Provider.Model) == "" {
		problems = append(problems, errors.New("model is not configured; set TL_MODEL or provider.model"))
	}
	if c.App.Timeout <= 0 {
		problems = append(problems, errors.New("timeout must be greater than zero"))
	}

	return errors.Join(problems...)
}
