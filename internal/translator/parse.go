package translator

import (
	"encoding/json"
	"errors"
	"fmt"
	"strings"
)

type vocabularyJSON struct {
	Word         string   `json:"word"`
	PartOfSpeech string   `json:"part_of_speech"`
	Category     string   `json:"category"`
	Definition   string   `json:"definition"`
	Chinese      string   `json:"chinese"`
	Examples     []string `json:"examples"`
}

type translationJSON struct {
	Translation string `json:"translation"`
}

// ParseVocabularyJSON parses and validates a structured vocabulary response.
func ParseVocabularyJSON(content string) (Vocabulary, error) {
	object, err := extractJSONObject(content)
	if err != nil {
		return Vocabulary{}, err
	}

	var decoded vocabularyJSON
	if err := json.Unmarshal(object, &decoded); err != nil {
		return Vocabulary{}, err
	}

	decoded.Word = strings.TrimSpace(decoded.Word)
	decoded.PartOfSpeech = strings.TrimSpace(decoded.PartOfSpeech)
	decoded.Category = strings.TrimSpace(decoded.Category)
	decoded.Definition = strings.TrimSpace(decoded.Definition)
	decoded.Chinese = strings.TrimSpace(decoded.Chinese)

	examples := make([]string, 0, len(decoded.Examples))
	for _, example := range decoded.Examples {
		if example = strings.TrimSpace(example); example != "" {
			examples = append(examples, example)
		}
	}
	if len(examples) < 2 {
		return Vocabulary{}, errors.New("expected at least two examples")
	}
	if decoded.PartOfSpeech == "" {
		return Vocabulary{}, errors.New("part_of_speech is empty")
	}
	if decoded.Definition == "" {
		return Vocabulary{}, errors.New("definition is empty")
	}
	if decoded.Chinese == "" {
		return Vocabulary{}, errors.New("chinese is empty")
	}

	return Vocabulary{
		Word:         decoded.Word,
		PartOfSpeech: decoded.PartOfSpeech,
		Category:     decoded.Category,
		Definition:   decoded.Definition,
		Chinese:      decoded.Chinese,
		Examples:     examples[:2],
	}, nil
}

// ParseTranslationJSON parses and validates a structured translation response.
func ParseTranslationJSON(content string) (string, error) {
	object, err := extractJSONObject(content)
	if err != nil {
		return "", err
	}

	var decoded translationJSON
	if err := json.Unmarshal(object, &decoded); err != nil {
		return "", err
	}
	translated := strings.TrimSpace(decoded.Translation)
	if translated == "" {
		return "", errors.New("translation is empty")
	}
	return translated, nil
}

func extractJSONObject(content string) ([]byte, error) {
	trimmed := strings.TrimSpace(content)
	if strings.HasPrefix(trimmed, "```") {
		if newline := strings.IndexByte(trimmed, '\n'); newline >= 0 {
			trimmed = trimmed[newline+1:]
		}
		trimmed = strings.TrimSuffix(strings.TrimSpace(trimmed), "```")
		trimmed = strings.TrimSpace(trimmed)
	}

	start := strings.IndexByte(trimmed, '{')
	end := strings.LastIndexByte(trimmed, '}')
	if start < 0 || end < start {
		return nil, fmt.Errorf("response does not contain a JSON object")
	}
	return []byte(trimmed[start : end+1]), nil
}
