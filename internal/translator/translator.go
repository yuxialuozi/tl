package translator

import (
	"context"
	"errors"
	"fmt"
	"regexp"
	"strings"
	"unicode"
)

// Language is a supported input or output language.
type Language string

const (
	English Language = "English"
	Chinese Language = "Chinese"
	Unknown Language = ""
)

// Output contains exactly one query result.
type Output struct {
	Vocabulary  *Vocabulary
	Translation *Translation
}

// Vocabulary is a learner-friendly dictionary entry.
type Vocabulary struct {
	Word         string
	PartOfSpeech string
	Category     string
	Definition   string
	Chinese      string
	Examples     []string
}

// Translation is a translated sentence or paragraph.
type Translation struct {
	SourceLanguage Language
	TargetLanguage Language
	Original       string
	Translation    string
}

// Completer is the model client required by Service.
type Completer interface {
	Complete(ctx context.Context, systemPrompt, userPrompt string) (string, error)
}

// Service classifies and processes user input.
type Service struct {
	client Completer
}

var englishWordPattern = regexp.MustCompile(`^[A-Za-z]+(?:['’-][A-Za-z]+)*$`)

// New creates a translator service.
func New(client Completer) *Service {
	return &Service{client: client}
}

// Run automatically selects vocabulary or translation mode.
func (s *Service) Run(ctx context.Context, input string) (Output, error) {
	input = strings.TrimSpace(input)
	if input == "" {
		return Output{}, errors.New("input is empty")
	}

	if IsSingleEnglishWord(input) {
		vocabulary, err := s.ExplainWord(ctx, input)
		if err != nil {
			return Output{}, err
		}
		return Output{Vocabulary: &vocabulary}, nil
	}

	source, ok := DetectLanguage(input)
	if !ok {
		return Output{}, errors.New("input must contain English or Chinese text")
	}
	target := Chinese
	if source == Chinese {
		target = English
	}

	translation, err := s.Translate(ctx, input, source, target)
	if err != nil {
		return Output{}, err
	}
	return Output{Translation: &translation}, nil
}

// IsSingleEnglishWord reports whether input is one English lexical word.
func IsSingleEnglishWord(input string) bool {
	return englishWordPattern.MatchString(strings.TrimSpace(input))
}

// DetectLanguage performs simple, predictable language detection for English and Chinese.
func DetectLanguage(input string) (Language, bool) {
	hasChinese := false
	hasLatin := false
	for _, r := range input {
		if unicode.Is(unicode.Han, r) {
			hasChinese = true
			continue
		}
		if unicode.Is(unicode.Latin, r) {
			hasLatin = true
		}
	}
	if hasChinese {
		return Chinese, true
	}
	if hasLatin {
		return English, true
	}
	return Unknown, false
}

// ExplainWord requests a dictionary entry for a single English word.
func (s *Service) ExplainWord(ctx context.Context, word string) (Vocabulary, error) {
	systemPrompt, userPrompt := BuildVocabularyPrompt(strings.TrimSpace(word))
	content, err := s.client.Complete(ctx, systemPrompt, userPrompt)
	if err != nil {
		return Vocabulary{}, err
	}
	vocabulary, err := ParseVocabularyJSON(content)
	if err != nil {
		return Vocabulary{}, fmt.Errorf("parse vocabulary response: %w", err)
	}
	if strings.TrimSpace(vocabulary.Word) == "" {
		vocabulary.Word = strings.TrimSpace(word)
	}
	return vocabulary, nil
}

// Translate requests a translation in the selected direction.
func (s *Service) Translate(ctx context.Context, text string, source, target Language) (Translation, error) {
	systemPrompt, userPrompt := BuildTranslationPrompt(text, source, target)
	content, err := s.client.Complete(ctx, systemPrompt, userPrompt)
	if err != nil {
		return Translation{}, err
	}
	translated, err := ParseTranslationJSON(content)
	if err != nil {
		return Translation{}, fmt.Errorf("parse translation response: %w", err)
	}
	return Translation{
		SourceLanguage: source,
		TargetLanguage: target,
		Original:       strings.TrimSpace(text),
		Translation:    translated,
	}, nil
}
