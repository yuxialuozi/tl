package config

import (
	"os"
	"path/filepath"
	"strings"
	"testing"
	"time"
)

func clearTLEnvironment(t *testing.T) {
	t.Helper()
	t.Setenv("TL_BASE_URL", "")
	t.Setenv("TL_API_KEY", "")
	t.Setenv("TL_MODEL", "")
}

func TestLoadFromDefaultsWhenFileIsMissing(t *testing.T) {
	clearTLEnvironment(t)
	path := filepath.Join(t.TempDir(), "missing.toml")

	loaded, err := LoadFrom(path)
	if err != nil {
		t.Fatalf("LoadFrom() error = %v", err)
	}
	if loaded.FileExists {
		t.Fatal("FileExists = true, want false")
	}
	if loaded.Config.Provider.BaseURL != defaultBaseURL {
		t.Fatalf("BaseURL = %q, want %q", loaded.Config.Provider.BaseURL, defaultBaseURL)
	}
	if loaded.Config.Timeout() != 30*time.Second {
		t.Fatalf("Timeout = %v, want 30s", loaded.Config.Timeout())
	}
	if loaded.Sources.BaseURL != SourceDefault || loaded.Sources.APIKey != SourceDefault || loaded.Sources.Model != SourceDefault {
		t.Fatalf("Sources = %+v, want defaults", loaded.Sources)
	}
}

func TestLoadFromFileAndEnvironmentPriority(t *testing.T) {
	clearTLEnvironment(t)
	path := filepath.Join(t.TempDir(), "config.toml")
	content := `[provider]
base_url = "https://file.example/v1"
api_key = "file-key"
model = "file-model"

[app]
timeout = 12
enable_thinking = false
`
	if err := os.WriteFile(path, []byte(content), 0o600); err != nil {
		t.Fatal(err)
	}

	t.Setenv("TL_BASE_URL", "https://env.example/v1/")
	t.Setenv("TL_MODEL", "env-model")

	loaded, err := LoadFrom(path)
	if err != nil {
		t.Fatalf("LoadFrom() error = %v", err)
	}
	if !loaded.FileExists {
		t.Fatal("FileExists = false, want true")
	}
	if loaded.Config.Provider.BaseURL != "https://env.example/v1" {
		t.Fatalf("BaseURL = %q", loaded.Config.Provider.BaseURL)
	}
	if loaded.Config.Provider.APIKey != "file-key" {
		t.Fatalf("APIKey = %q", loaded.Config.Provider.APIKey)
	}
	if loaded.Config.Provider.Model != "env-model" {
		t.Fatalf("Model = %q", loaded.Config.Provider.Model)
	}
	if loaded.Config.App.Timeout != 12 {
		t.Fatalf("Timeout = %d", loaded.Config.App.Timeout)
	}
	if loaded.Config.App.EnableThinking == nil || *loaded.Config.App.EnableThinking {
		t.Fatalf("EnableThinking = %v, want false", loaded.Config.App.EnableThinking)
	}
	if loaded.Sources.BaseURL != SourceEnvironment || loaded.Sources.APIKey != SourceFile || loaded.Sources.Model != SourceEnvironment {
		t.Fatalf("Sources = %+v", loaded.Sources)
	}
	if err := loaded.Config.Validate(); err != nil {
		t.Fatalf("Validate() error = %v", err)
	}
}

func TestLoadFromEnvironmentWithoutFile(t *testing.T) {
	clearTLEnvironment(t)
	t.Setenv("TL_BASE_URL", "http://127.0.0.1:8080/v1")
	t.Setenv("TL_API_KEY", "env-key")
	t.Setenv("TL_MODEL", "env-model")

	loaded, err := LoadFrom(filepath.Join(t.TempDir(), "missing.toml"))
	if err != nil {
		t.Fatalf("LoadFrom() error = %v", err)
	}
	if loaded.Sources.BaseURL != SourceEnvironment || loaded.Sources.APIKey != SourceEnvironment || loaded.Sources.Model != SourceEnvironment {
		t.Fatalf("Sources = %+v", loaded.Sources)
	}
	if err := loaded.Config.Validate(); err != nil {
		t.Fatalf("Validate() error = %v", err)
	}
}

func TestEnvironmentAPIKeyOverridesFile(t *testing.T) {
	clearTLEnvironment(t)
	path := filepath.Join(t.TempDir(), "config.toml")
	content := `[provider]
base_url = "https://file.example/v1"
api_key = "file-key"
model = "file-model"
`
	if err := os.WriteFile(path, []byte(content), 0o600); err != nil {
		t.Fatal(err)
	}
	t.Setenv("TL_API_KEY", "env-key")

	loaded, err := LoadFrom(path)
	if err != nil {
		t.Fatalf("LoadFrom() error = %v", err)
	}
	if loaded.Config.Provider.APIKey != "env-key" {
		t.Fatalf("APIKey = %q, want env-key", loaded.Config.Provider.APIKey)
	}
	if loaded.Sources.APIKey != SourceEnvironment {
		t.Fatalf("APIKey source = %q", loaded.Sources.APIKey)
	}
}

func TestLoadFromMalformedFile(t *testing.T) {
	clearTLEnvironment(t)
	path := filepath.Join(t.TempDir(), "config.toml")
	if err := os.WriteFile(path, []byte("[provider\n"), 0o600); err != nil {
		t.Fatal(err)
	}

	_, err := LoadFrom(path)
	if err == nil || !strings.Contains(err.Error(), "parse configuration") {
		t.Fatalf("LoadFrom() error = %v, want parse error", err)
	}
}

func TestValidateReportsAllProblems(t *testing.T) {
	tests := []struct {
		name   string
		config Config
		want   []string
	}{
		{
			name:   "missing key and model",
			config: Config{Provider: Provider{BaseURL: "https://example.com/v1"}, App: App{Timeout: 30}},
			want:   []string{"API key", "model"},
		},
		{
			name:   "invalid URL",
			config: Config{Provider: Provider{BaseURL: "example.com", APIKey: "key", Model: "model"}, App: App{Timeout: 30}},
			want:   []string{"absolute http(s) URL"},
		},
		{
			name:   "invalid timeout",
			config: Config{Provider: Provider{BaseURL: "https://example.com/v1", APIKey: "key", Model: "model"}, App: App{Timeout: 0}},
			want:   []string{"timeout"},
		},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			err := test.config.Validate()
			if err == nil {
				t.Fatal("Validate() error = nil")
			}
			for _, part := range test.want {
				if !strings.Contains(err.Error(), part) {
					t.Fatalf("Validate() error = %q, want substring %q", err, part)
				}
			}
		})
	}
}
