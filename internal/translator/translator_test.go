package translator

import (
	"context"
	"errors"
	"strings"
	"testing"
)

type completerFunc func(ctx context.Context, systemPrompt, userPrompt string) (string, error)

func (f completerFunc) Complete(ctx context.Context, systemPrompt, userPrompt string) (string, error) {
	return f(ctx, systemPrompt, userPrompt)
}

func TestIsSingleEnglishWord(t *testing.T) {
	tests := []struct {
		input string
		want  bool
	}{
		{input: "keyring", want: true},
		{input: "API", want: true},
		{input: "don't", want: true},
		{input: "state-of-the-art", want: true},
		{input: "hello world", want: false},
		{input: "Hello!", want: false},
		{input: "你好", want: false},
		{input: "", want: false},
	}

	for _, test := range tests {
		if got := IsSingleEnglishWord(test.input); got != test.want {
			t.Errorf("IsSingleEnglishWord(%q) = %t, want %t", test.input, got, test.want)
		}
	}
}

func TestDetectLanguage(t *testing.T) {
	tests := []struct {
		input  string
		want   Language
		wantOK bool
	}{
		{input: "The daemon runs in the background.", want: English, wantOK: true},
		{input: "守护进程在后台运行。", want: Chinese, wantOK: true},
		{input: "使用 Go 编写 daemon", want: Chinese, wantOK: true},
		{input: "12345", want: Unknown, wantOK: false},
	}

	for _, test := range tests {
		got, ok := DetectLanguage(test.input)
		if got != test.want || ok != test.wantOK {
			t.Errorf("DetectLanguage(%q) = (%q, %t), want (%q, %t)", test.input, got, ok, test.want, test.wantOK)
		}
	}
}

func TestPromptBuilders(t *testing.T) {
	system, user := BuildVocabularyPrompt("keyring")
	if !strings.Contains(system, `"part_of_speech"`) || !strings.Contains(system, `"examples"`) {
		t.Fatalf("vocabulary system prompt missing schema: %s", system)
	}
	if user != "Explain this English word: keyring" {
		t.Fatalf("vocabulary user prompt = %q", user)
	}

	system, user = BuildTranslationPrompt("hello", English, Chinese)
	if !strings.Contains(system, `{"translation":"translated text"}`) {
		t.Fatalf("translation system prompt missing schema: %s", system)
	}
	for _, part := range []string{"Source language: English", "Target language: Chinese", "Text:\nhello"} {
		if !strings.Contains(user, part) {
			t.Fatalf("translation user prompt %q missing %q", user, part)
		}
	}
}

func TestParseVocabularyJSON(t *testing.T) {
	content := "```json\n" + `{
  "word": "keyring",
  "part_of_speech": "noun",
  "category": "Computing",
  "definition": "A secure place for passwords.",
  "chinese": "密钥环",
  "examples": ["The keyring stores the password.", "Linux uses a keyring."]
}` + "\n```"

	got, err := ParseVocabularyJSON(content)
	if err != nil {
		t.Fatalf("ParseVocabularyJSON() error = %v", err)
	}
	if got.Word != "keyring" || got.PartOfSpeech != "noun" || got.Category != "Computing" {
		t.Fatalf("unexpected vocabulary: %+v", got)
	}
	if len(got.Examples) != 2 {
		t.Fatalf("examples = %v", got.Examples)
	}
}

func TestParseVocabularyJSONRejectsIncompleteData(t *testing.T) {
	tests := []string{
		`not json`,
		`{"part_of_speech":"noun","definition":"x","chinese":"y","examples":["only one"]}`,
		`{"part_of_speech":"","definition":"x","chinese":"y","examples":["one","two"]}`,
		`{"part_of_speech":"noun","definition":"","chinese":"y","examples":["one","two"]}`,
	}
	for _, content := range tests {
		if _, err := ParseVocabularyJSON(content); err == nil {
			t.Errorf("ParseVocabularyJSON(%q) error = nil", content)
		}
	}
}

func TestParseTranslationJSON(t *testing.T) {
	got, err := ParseTranslationJSON("Here is the result:\n{\"translation\":\"你好\"}\nDone.")
	if err != nil {
		t.Fatalf("ParseTranslationJSON() error = %v", err)
	}
	if got != "你好" {
		t.Fatalf("ParseTranslationJSON() = %q", got)
	}
	if _, err := ParseTranslationJSON(`{"translation":""}`); err == nil {
		t.Fatal("ParseTranslationJSON() error = nil for empty translation")
	}
}

func TestServiceRunVocabulary(t *testing.T) {
	response := `{"word":"keyring","part_of_speech":"noun","category":"Computing","definition":"A secure place for passwords.","chinese":"密钥环","examples":["The keyring stores the password.","Linux uses a keyring."]}`
	service := New(completerFunc(func(_ context.Context, system, user string) (string, error) {
		if !strings.Contains(system, `"chinese"`) {
			t.Errorf("system prompt = %q", system)
		}
		if user != "Explain this English word: keyring" {
			t.Errorf("user prompt = %q", user)
		}
		return response, nil
	}))

	output, err := service.Run(context.Background(), "keyring")
	if err != nil {
		t.Fatalf("Run() error = %v", err)
	}
	if output.Vocabulary == nil || output.Translation != nil {
		t.Fatalf("output = %+v", output)
	}
	if output.Vocabulary.Word != "keyring" {
		t.Fatalf("word = %q", output.Vocabulary.Word)
	}
}

func TestServiceRunTranslation(t *testing.T) {
	tests := []struct {
		name       string
		input      string
		wantSource Language
		wantTarget Language
	}{
		{name: "English to Chinese", input: "The daemon runs in the background.", wantSource: English, wantTarget: Chinese},
		{name: "Chinese to English", input: "守护进程在后台运行。", wantSource: Chinese, wantTarget: English},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			service := New(completerFunc(func(_ context.Context, _ string, user string) (string, error) {
				if !strings.Contains(user, "Target language: "+string(test.wantTarget)) {
					t.Errorf("user prompt = %q", user)
				}
				return `{"translation":"translated"}`, nil
			}))

			output, err := service.Run(context.Background(), test.input)
			if err != nil {
				t.Fatalf("Run() error = %v", err)
			}
			if output.Translation == nil || output.Vocabulary != nil {
				t.Fatalf("output = %+v", output)
			}
			if output.Translation.Original != test.input {
				t.Fatalf("original = %q", output.Translation.Original)
			}
			if output.Translation.SourceLanguage != test.wantSource || output.Translation.TargetLanguage != test.wantTarget {
				t.Fatalf("languages = %s -> %s", output.Translation.SourceLanguage, output.Translation.TargetLanguage)
			}
		})
	}
}

func TestServiceRunPropagatesClientError(t *testing.T) {
	wantErr := errors.New("network failed")
	service := New(completerFunc(func(context.Context, string, string) (string, error) {
		return "", wantErr
	}))

	_, err := service.Run(context.Background(), "keyring")
	if !errors.Is(err, wantErr) {
		t.Fatalf("Run() error = %v, want %v", err, wantErr)
	}
}
