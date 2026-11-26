package game

import (
	"fmt"
	"strings"
)

func (m Model) modeView() string {
	book := m.books[m.selected]
	width := min(76, max(62, m.width-8))
	lines := []string{
		accentStyle.Render("SELECT MODE // 选择学习方式"),
		textStyle.Render("ARCHIVE  ") + titleStyle.Render(book.Name) + mutedStyle.Render(fmt.Sprintf("  (%d WORDS)", len(book.Words))),
		"",
	}
	for index, mode := range gameModes {
		label := fmt.Sprintf("%-8s %s", mode.Name, mode.Description)
		if index == 1 {
			label = fmt.Sprintf("%-8s %s  %d WORDS", mode.Name, mode.Description, len(book.Words))
		} else if index == 2 {
			label = fmt.Sprintf("%-8s %s  %d WORDS", mode.Name, mode.Description, m.reviewCount(book))
		}
		if index == m.mode {
			lines = append(lines, selectedStyle.Render("▶ "+label))
		} else {
			lines = append(lines, textStyle.Render("  "+label))
		}
	}
	lines = append(
		lines,
		"",
		mutedStyle.Render("学习模式按原顺序覆盖整本词书；错词会自动进入巩固列表。"),
		mutedStyle.Render("↑/↓ 或 J/K 选择    ENTER 开始    ESC 返回    Q 退出"),
	)
	return panelStyle.Width(width).Render(strings.Join(lines, "\n"))
}

func (m Model) studyView() string {
	study := m.study
	if study == nil {
		return mutedStyle.Render("study state unavailable")
	}
	width := min(76, max(62, m.width-8))
	modeLabel := "FULL WORD BOOK"
	if study.mode == studyReview {
		modeLabel = "MISTAKE REVIEW"
	}
	header := []string{
		titleStyle.Render("TL // STUDY LINK"),
		mutedStyle.Render("ARCHIVE  " + study.book.Name + "  //  " + modeLabel),
		"",
	}

	if study.stage == studyDone {
		return m.studyResultView(width, header)
	}

	current := study.words[study.index]
	progress := fmt.Sprintf("%d/%d", study.index+1, len(study.words))
	lines := append([]string{}, header...)
	lines = append(
		lines,
		textStyle.Render("进度 ")+accentStyle.Render(progress)+textStyle.Render(fmt.Sprintf("  认识 %d  不认识 %d", study.known, study.unknown)),
		"",
		wordHeaderLine(current),
		mutedStyle.Render(current.Phonetic),
		"",
	)
	if study.stage == studyPrompt {
		lines = append(
			lines,
			textStyle.Render("先回忆中文释义和英文解释，再显示答案。"),
			"",
			warnStyle.Render("ENTER / SPACE  显示释义"),
		)
	} else {
		contentWidth := max(24, width-6)
		lines = append(
			lines,
			sectionStyle.Render("Definition"),
		)
		lines = append(lines, wrappedLimited(current.Definition, contentWidth, 2, textStyle)...)
		lines = append(
			lines,
			"",
			sectionStyle.Render("中文"),
			textStyle.Render(current.Meaning),
			"",
			sectionStyle.Render("Examples"),
		)
		for _, example := range current.Examples[:min(1, len(current.Examples))] {
			lines = append(lines, wrappedLimited("- "+example, contentWidth, 2, textStyle)...)
		}
		lines = append(
			lines,
			"",
			textStyle.Render("1  还不认识（加入错词）    2  ENTER  记住了"),
		)
	}
	lines = append(lines, "", mutedStyle.Render("ESC 返回模式选择    Q 退出"))
	return panelStyle.Width(width).Render(strings.Join(lines, "\n"))
}

func (m Model) studyResultView(width int, header []string) string {
	study := m.study
	lines := append([]string{}, header...)
	if len(study.words) == 0 {
		lines = append(
			lines,
			warnStyle.Render("NO WORDS TO REVIEW"),
			textStyle.Render("当前词书还没有标记为不认识的单词。"),
			"",
			textStyle.Render("先完整学习一遍词书，错词会自动收集到这里。"),
		)
	} else {
		lines = append(
			lines,
			titleStyle.Render("SESSION COMPLETE"),
			textStyle.Render(fmt.Sprintf("已遍历  %d WORDS", len(study.words))),
			textStyle.Render(fmt.Sprintf("认识    %d", study.known)),
			textStyle.Render(fmt.Sprintf("不认识  %d", study.unknown)),
			textStyle.Render(fmt.Sprintf("待巩固  %d", m.reviewCount(study.book))),
		)
	}
	lines = append(
		lines,
		"",
		mutedStyle.Render("ENTER 返回模式选择    R 再复习错词    Q 退出"),
	)
	return panelStyle.Width(width).Render(strings.Join(lines, "\n"))
}
