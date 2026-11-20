package renderer

import (
	"strings"
	"testing"
	"time"

	"github.com/yuxialuozi/tl/internal/translator"
)

func TestVocabularyPlainText(t *testing.T) {
	output := New(false).Vocabulary(translator.Vocabulary{
		Word:         "keyring",
		PartOfSpeech: "noun",
		Category:     "Computing",
		Definition:   "A secure place for passwords.",
		Chinese:      "密钥环",
		Examples:     []string{"The keyring stores the password.", "Linux uses a keyring."},
	})

	for _, part := range []string{
		"keyring · noun · Computing",
		"Definition\nA secure place for passwords.",
		"中文\n密钥环",
		"Examples\n- The keyring stores the password.\n- Linux uses a keyring.",
	} {
		if !strings.Contains(output, part) {
			t.Errorf("output missing %q:\n%s", part, output)
		}
	}
	if strings.Contains(output, "\x1b[") {
		t.Fatalf("plain output contains ANSI escapes: %q", output)
	}
}

func TestTranslationPlainText(t *testing.T) {
	output := New(false).Translation(translator.Translation{
		Original:    "The daemon runs in the background.",
		Translation: "该守护进程在后台运行。",
	})
	want := "Original\nThe daemon runs in the background.\n\nTranslation\n该守护进程在后台运行。\n"
	if output != want {
		t.Fatalf("output = %q, want %q", output, want)
	}
}

func TestConfigDoesNotExposeSecret(t *testing.T) {
	output := New(false).Config(ConfigStatus{
		Path:         "/home/user/.config/tl/config.toml",
		FileExists:   true,
		BaseURL:      "https://api.example.com/v1",
		Model:        "example-model",
		APIKeyStatus: "configured via TL_API_KEY",
		Timeout:      30 * time.Second,
	})
	for _, part := range []string{"Path: /home/user/.config/tl/config.toml", "API key: configured via TL_API_KEY", "Status: ready"} {
		if !strings.Contains(output, part) {
			t.Errorf("output missing %q:\n%s", part, output)
		}
	}
	if strings.Contains(output, "secret") {
		t.Fatalf("config output leaked a secret: %s", output)
	}
}
