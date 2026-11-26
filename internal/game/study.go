package game

import (
	"strings"

	tea "github.com/charmbracelet/bubbletea"
)

type gameMode struct {
	Name        string
	Description string
}

var gameModes = []gameMode{
	{Name: "RAID", Description: "5 分钟 Roguelike 词汇地牢"},
	{Name: "STUDY", Description: "按顺序完整学习当前词书"},
	{Name: "REVIEW", Description: "只复习标记为不认识的单词"},
}

type studyMode int

const (
	studyFull studyMode = iota
	studyReview
)

type studyStage int

const (
	studyPrompt studyStage = iota
	studyReveal
	studyDone
)

type studyState struct {
	book    Wordbook
	mode    studyMode
	words   []Word
	index   int
	stage   studyStage
	known   int
	unknown int
}

func (m Model) startMode() (tea.Model, tea.Cmd) {
	switch m.mode {
	case 1:
		return m.startStudy(studyFull), nil
	case 2:
		return m.startStudy(studyReview), nil
	default:
		return m.startRun()
	}
}

func (m Model) startStudy(mode studyMode) Model {
	book := m.books[m.selected]
	words := append([]Word(nil), book.Words...)
	if mode == studyReview {
		words = m.reviewWords(book)
	}
	for index := range words {
		words[index].Term = strings.ToLower(words[index].Term)
	}

	stage := studyPrompt
	if len(words) == 0 {
		stage = studyDone
	}
	m.study = &studyState{
		book:  book,
		mode:  mode,
		words: words,
		stage: stage,
	}
	m.screen = screenStudy
	return m
}

func (m Model) updateStudyKey(key string) (tea.Model, tea.Cmd) {
	if m.study == nil {
		m.screen = screenMode
		return m, nil
	}

	switch key {
	case "q":
		return m, tea.Quit
	case "esc", "backspace":
		m.screen = screenMode
		return m, nil
	}

	switch m.study.stage {
	case studyPrompt:
		switch key {
		case "enter", " ":
			m.study.stage = studyReveal
		}
	case studyReveal:
		switch key {
		case "1", "f":
			m.markStudyWord(false)
		case "2", "j", "enter", " ":
			m.markStudyWord(true)
		}
	case studyDone:
		switch key {
		case "enter", " ":
			m.screen = screenMode
		case "r":
			return m.startStudy(studyReview), nil
		}
	}
	return m, nil
}

func (m *Model) markStudyWord(known bool) {
	study := m.study
	if study == nil || study.index < 0 || study.index >= len(study.words) {
		return
	}
	term := strings.ToLower(study.words[study.index].Term)
	if m.reviewTerms == nil {
		m.reviewTerms = map[string]map[string]bool{}
	}
	if m.reviewTerms[study.book.ID] == nil {
		m.reviewTerms[study.book.ID] = map[string]bool{}
	}
	if known {
		study.known++
		delete(m.reviewTerms[study.book.ID], term)
	} else {
		study.unknown++
		m.reviewTerms[study.book.ID][term] = true
	}

	study.index++
	study.stage = studyPrompt
	if study.index >= len(study.words) {
		study.stage = studyDone
	}
}

func (m Model) reviewWords(book Wordbook) []Word {
	terms := m.reviewTerms[book.ID]
	words := make([]Word, 0, len(terms))
	for _, word := range book.Words {
		if terms[strings.ToLower(word.Term)] {
			words = append(words, word)
		}
	}
	return words
}

func (m Model) reviewCount(book Wordbook) int {
	return len(m.reviewTerms[book.ID])
}
