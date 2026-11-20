package translator

import "fmt"

const vocabularySystemPrompt = `You are a concise English learner's dictionary.
Return only one valid JSON object with this exact shape:
{"word":"keyring","part_of_speech":"noun","category":"Computing","definition":"A system that securely stores passwords and other secret information.","chinese":"密钥环，用于安全存储密码、密钥等敏感信息。","examples":["The application stores the password in the keyring.","Linux uses a keyring to manage credentials."]}

Rules:
- definition must be simple, natural English that an intermediate learner can understand.
- chinese must be a concise Simplified Chinese meaning, not an explanation.
- examples must contain exactly two natural English sentences.
- category is an optional short subject label such as Computing or Biology.
- Do not use Markdown, commentary, or extra keys.`

const translationSystemPrompt = `You are a precise translator.
Return only one valid JSON object with this exact shape:
{"translation":"translated text"}

Rules:
- Translate according to the requested source and target languages.
- Preserve meaning, tone, names, numbers, and paragraph breaks.
- Do not explain grammar, define words, or add advice.
- Do not use Markdown, commentary, or extra keys.`

// BuildVocabularyPrompt returns the system and user prompts for a word lookup.
func BuildVocabularyPrompt(word string) (string, string) {
	return vocabularySystemPrompt, fmt.Sprintf("Explain this English word: %s", word)
}

// BuildTranslationPrompt returns the system and user prompts for translation.
func BuildTranslationPrompt(text string, source, target Language) (string, string) {
	userPrompt := fmt.Sprintf("Source language: %s\nTarget language: %s\n\nText:\n%s", source, target, text)
	return translationSystemPrompt, userPrompt
}
