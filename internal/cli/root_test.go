package cli

import (
	"bytes"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func TestQueryVocabularyEndToEnd(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path != "/v1/chat/completions" {
			t.Errorf("path = %s", r.URL.Path)
		}
		if got := r.Header.Get("Authorization"); got != "Bearer cli-test-key" {
			t.Errorf("Authorization = %q", got)
		}

		var request struct {
			Messages []struct {
				Content string `json:"content"`
			} `json:"messages"`
		}
		if err := json.NewDecoder(r.Body).Decode(&request); err != nil {
			t.Fatal(err)
		}
		if len(request.Messages) < 2 || !strings.Contains(request.Messages[1].Content, "keyring") {
			t.Errorf("messages = %+v", request.Messages)
		}

		_, _ = w.Write([]byte(`{"choices":[{"message":{"content":"{\"word\":\"keyring\",\"part_of_speech\":\"noun\",\"category\":\"Computing\",\"definition\":\"A secure place for passwords.\",\"chinese\":\"密钥环\",\"examples\":[\"The keyring stores the password.\",\"Linux uses a keyring.\"]}"}}]}`))
	}))
	defer server.Close()

	configDir := t.TempDir()
	t.Setenv("XDG_CONFIG_HOME", configDir)
	t.Setenv("TL_API_KEY", "")
	t.Setenv("TL_BASE_URL", "")
	t.Setenv("TL_MODEL", "")
	writeTestConfig(t, configDir, server.URL+"/v1", "cli-test-key")

	var stdout, stderr bytes.Buffer
	command := New("test-version", &stdout, &stderr)
	command.SetArgs([]string{"keyring"})
	if err := command.Execute(); err != nil {
		t.Fatalf("Execute() error = %v; stderr = %s", err, stderr.String())
	}

	for _, part := range []string{"keyring · noun · Computing", "A secure place for passwords.", "密钥环", "Linux uses a keyring."} {
		if !strings.Contains(stdout.String(), part) {
			t.Errorf("stdout missing %q:\n%s", part, stdout.String())
		}
	}
}

func TestConfigCommandRedactsAPIKey(t *testing.T) {
	configDir := t.TempDir()
	t.Setenv("XDG_CONFIG_HOME", configDir)
	t.Setenv("TL_API_KEY", "")
	t.Setenv("TL_BASE_URL", "")
	t.Setenv("TL_MODEL", "")
	writeTestConfig(t, configDir, "https://api.example.com/v1", "very-secret-key")

	var stdout, stderr bytes.Buffer
	command := New("test-version", &stdout, &stderr)
	command.SetArgs([]string{"config"})
	if err := command.Execute(); err != nil {
		t.Fatalf("Execute() error = %v", err)
	}

	output := stdout.String()
	if strings.Contains(output, "very-secret-key") {
		t.Fatalf("config output leaked API key:\n%s", output)
	}
	for _, part := range []string{"Path: ", "File: found", "Model: test-model", "API key: configured in config file", "Status: ready"} {
		if !strings.Contains(output, part) {
			t.Errorf("output missing %q:\n%s", part, output)
		}
	}
}

func TestVersionCommand(t *testing.T) {
	var stdout, stderr bytes.Buffer
	command := New("1.2.3", &stdout, &stderr)
	command.SetArgs([]string{"version"})
	if err := command.Execute(); err != nil {
		t.Fatalf("Execute() error = %v", err)
	}
	if got := stdout.String(); got != "tl 1.2.3\n" {
		t.Fatalf("stdout = %q", got)
	}
}

func TestHelpDoesNotListExtraCommands(t *testing.T) {
	var stdout, stderr bytes.Buffer
	command := New("test", &stdout, &stderr)
	command.SetArgs([]string{"--help"})
	if err := command.Execute(); err != nil {
		t.Fatalf("Execute() error = %v", err)
	}
	if strings.Contains(stdout.String(), "Available Commands") || strings.Contains(stdout.String(), "completion") {
		t.Fatalf("help lists unexpected commands:\n%s", stdout.String())
	}
	for _, part := range []string{"tl config", "tl version", "tl keyring", "--game"} {
		if !strings.Contains(stdout.String(), part) {
			t.Errorf("help missing %q:\n%s", part, stdout.String())
		}
	}
}

func TestGameFlagRejectsInputArguments(t *testing.T) {
	var stdout, stderr bytes.Buffer
	command := New("test", &stdout, &stderr)
	command.SetArgs([]string{"--game", "hello"})
	err := command.Execute()
	if err == nil || !strings.Contains(err.Error(), "--game does not accept input") {
		t.Fatalf("Execute() error = %v", err)
	}
}

func TestRootRejectsUnquotedMultipleArguments(t *testing.T) {
	var stdout, stderr bytes.Buffer
	command := New("test", &stdout, &stderr)
	command.SetArgs([]string{"hello", "world"})
	err := command.Execute()
	if err == nil || !strings.Contains(err.Error(), "quote text containing spaces") {
		t.Fatalf("Execute() error = %v", err)
	}
}

func writeTestConfig(t *testing.T, configDir, baseURL, apiKey string) {
	t.Helper()
	path := filepath.Join(configDir, "tl", "config.toml")
	if err := os.MkdirAll(filepath.Dir(path), 0o755); err != nil {
		t.Fatal(err)
	}
	content := "[provider]\nbase_url = \"" + baseURL + "\"\napi_key = \"" + apiKey + "\"\nmodel = \"test-model\"\n\n[app]\ntimeout = 5\n"
	if err := os.WriteFile(path, []byte(content), 0o600); err != nil {
		t.Fatal(err)
	}
}
