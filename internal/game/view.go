package game

import (
	"fmt"
	"math"
	"strings"
	"time"

	"github.com/charmbracelet/lipgloss"
)

// View implements tea.Model.
func (m Model) View() string {
	if m.width > 0 && (m.width < 80 || m.height < 23) {
		return lipgloss.JoinVertical(
			lipgloss.Left,
			warnStyle.Render("TL // LEXICON RAID"),
			"",
			textStyle.Render("Terminal too small for the raid UI."),
			mutedStyle.Render(fmt.Sprintf("Need at least 80x23, current: %dx%d", m.width, m.height)),
			"",
			mutedStyle.Render("Resize the terminal or press Q to quit."),
		)
	}

	switch m.screen {
	case screenSelect:
		return m.selectView()
	case screenMode:
		return m.modeView()
	case screenPlaying:
		return m.playingView()
	case screenStudy:
		return m.studyView()
	case screenSettlement:
		return m.settlementView()
	default:
		return ""
	}
}

func (m Model) selectView() string {
	width := min(82, max(66, m.width-8))
	title := lipgloss.JoinVertical(
		lipgloss.Left,
		titleStyle.Render("████████╗██╗     "),
		titleStyle.Render("╚══██╔══╝██║     "),
		titleStyle.Render("   ██║   ██║     "),
		titleStyle.Render("   ██║   ██║     "),
		titleStyle.Render("   ██║   ███████╗"),
		titleStyle.Render("   ╚═╝   ╚══════╝"),
		"",
		accentStyle.Render("L E X I C O N   R A I D"),
		mutedStyle.Render("VOCABULARY DUNGEON // STUDY OR RAID"),
	)

	var bookLines []string
	for index, book := range m.books {
		line := fmt.Sprintf("%-22s %s  %s", book.Name, book.Description, mutedStyle.Render(fmt.Sprintf("%d WORDS", len(book.Words))))
		if index == m.selected {
			line = selectedStyle.Render("▶ " + line)
		} else {
			line = "  " + line
		}
		bookLines = append(bookLines, line)
	}
	menu := panelStyle.Width(width).Render(strings.Join(bookLines, "\n"))

	footer := lipgloss.JoinVertical(
		lipgloss.Left,
		"",
		textStyle.Render("↑/↓ 或 J/K 选择词书    ENTER 选择模式    Q 退出"),
		mutedStyle.Render("学习模式完整过词书 · 错词巩固 · 5 分钟词汇 Raid"),
	)

	return lipgloss.JoinVertical(lipgloss.Left, title, "", menu, footer)
}

func (m Model) playingView() string {
	if m.run == nil {
		return "raid state unavailable"
	}
	if m.run.paused {
		return m.pausedView()
	}
	header := m.playingHeader()
	left := m.mapPanel()
	if m.run.inCombat {
		left = m.encounterPanel()
	}
	rightWidth := max(40, m.width-lipgloss.Width(left)-4)
	right := m.explorePanel(rightWidth)
	footer := mutedStyle.Render(" WASD / ↑↓←→ 移动 · 碰撞怪物开战 · 踩到 ? 拾取 · Q 退出")
	if m.run.bagOpen {
		right = m.bagPanel(rightWidth)
		footer = mutedStyle.Render(" ↑↓ / J/K 选择 · ENTER 使用 · E / ESC 关闭 · P 暂停 · Q 退出")
	} else if m.run.awaitingSkill {
		right = m.skillPanel(rightWidth)
		footer = mutedStyle.Render(" 1-3 选择技能 · P 暂停 · Q 退出")
	} else if m.run.inCombat {
		right = m.battlePanel(rightWidth)
		if m.run.combatLocked {
			footer = mutedStyle.Render(" ENTER 继续 · P 暂停 · Q 退出")
		} else {
			footer = mutedStyle.Render(" A-D / 1-4 答题 · E 背包 · P 暂停 · Q 退出")
		}
	} else {
		footer = mutedStyle.Render(" WASD / ↑↓←→ 移动 · 碰撞怪物开战 · E 背包 · P 暂停 · Q 退出")
	}
	body := lipgloss.JoinHorizontal(lipgloss.Top, left, right)
	return lipgloss.JoinVertical(lipgloss.Left, header, body, footer)
}

func (m Model) playingHeader() string {
	run := m.run
	left := titleStyle.Render("TL // LEXICON RAID")
	timeLabel := "TIME " + formatRemaining(run.remaining)
	if run.combatLocked {
		timeLabel = "TIME HELD"
	}
	right := fmt.Sprintf(
		"%s   %s   %s   %s",
		accentStyle.Render(timeLabel),
		textStyle.Render("COMBO "+fmt.Sprintf("%02d", run.combo)),
		textStyle.Render(fmt.Sprintf("WORDS %d", run.correct)),
		mutedStyle.Render(run.book.Name),
	)
	gap := m.width - lipgloss.Width(left) - lipgloss.Width(right)
	if gap < 1 {
		gap = 1
	}
	return left + strings.Repeat(" ", gap) + right
}

func (m Model) mapPanel() string {
	run := m.run
	lines := []string{
		titleStyle.Render("SECTOR MAP") + "  " + mutedStyle.Render(run.sector),
	}

	pathPositions := make(map[point]bool, len(run.dungeon.path))
	for _, p := range run.dungeon.path {
		pathPositions[p] = true
	}

	for y, row := range run.dungeon.grid {
		var line strings.Builder
		for x, cell := range row {
			p := point{x: x, y: y}
			switch {
			case p == run.player:
				line.WriteString(playerStyle.Render("@"))
			case run.hasMonsterAt(p):
				line.WriteString(monsterStyle(run.monsterKindAt(p)).Render(run.monsterGlyphAt(p)))
			case run.hasItemAt(p):
				line.WriteString(lootStyle.Render("?"))
			case cell == '#':
				line.WriteString(wallStyle.Render("#"))
			case pathPositions[p]:
				line.WriteString(pathStyle.Render("·"))
			default:
				line.WriteString(mutedStyle.Render("."))
			}
		}
		lines = append(lines, line.String())
	}

	lines = append(
		lines,
		mutedStyle.Render(strings.Repeat("─", mapWidth)),
		textStyle.Render("PLAYER  ")+hpBar(run.playerHP, run.playerMaxHP, 17),
		textStyle.Render(fmt.Sprintf("COMBO %02d   %s", run.combo, weaponLabel(run))),
	)
	return panelStyle.Width(mapWidth + 2).Render(strings.Join(lines, "\n"))
}

func (m Model) explorePanel(width int) string {
	run := m.run
	lines := []string{
		accentStyle.Render("SECTOR // EXPLORATION"),
		textStyle.Render("WASD 移动 · 碰到怪物进入战斗"),
		textStyle.Render("HP ") + hpBar(run.playerHP, run.playerMaxHP, 16) + textStyle.Render(fmt.Sprintf(" %d/%d", run.playerHP, run.playerMaxHP)),
		textStyle.Render("WEAPON   ") + textStyle.Render(weaponLabel(run)),
		textStyle.Render("PASSIVE  ") + mutedStyle.Render(clipPlain(passiveSummary(run), max(12, width-12))),
		textStyle.Render("[E] 背包  ") + mutedStyle.Render(consumableSummary(run)),
		"",
		sectionStyle.Render("MONSTERS"),
		monsterLegend(run),
		mutedStyle.Render("? 道具   @ 玩家   · 通路"),
		"",
		mutedStyle.Render(strings.Repeat("─", max(20, width-2))),
		sectionStyle.Render("SYSTEM LOG"),
	}
	lines = append(lines, logLines(run, 5, width)...)
	return panelStyle.Width(width + 2).Render(strings.Join(lines, "\n"))
}

func (m Model) battlePanel(width int) string {
	run := m.run
	if run.combat == nil {
		return "combat state unavailable"
	}
	if run.combatLocked {
		return m.resultPanel(width)
	}
	lines := []string{
		accentStyle.Render("TARGET LOCK // 选择中文释义"),
		playerHealthLine(run, 18),
		textStyle.Render(fmt.Sprintf("COMBO %02d · %s", run.combo, weaponLabel(run))),
		"",
		wordHeaderLine(run.question.word),
	}
	if run.question.word.Phonetic != "" {
		lines = append(lines, mutedStyle.Render(run.question.word.Phonetic))
	}
	lines = append(lines, m.questionLines()...)
	lines = append(
		lines,
		"",
		textStyle.Render(clipPlain(inventoryLine(run), max(16, width-2))),
		mutedStyle.Render(strings.Repeat("─", max(20, width-2))),
		sectionStyle.Render("SYSTEM LOG"),
	)
	lines = append(lines, logLines(run, 2, width)...)
	return panelStyle.Width(width + 2).Render(strings.Join(lines, "\n"))
}

func (m Model) questionLines() []string {
	run := m.run
	lines := []string{textStyle.Render("选择正确的中文释义"), ""}
	labels := []string{"A", "B", "C", "D"}
	for index, option := range run.question.options {
		lines = append(lines, optionStyle.Render(fmt.Sprintf("[%s] %s", labels[index], option)))
	}
	return lines
}

func (m Model) resultPanel(width int) string {
	run := m.run
	word := run.question.word
	resultStyle := dangerStyle
	if run.lastCorrect {
		resultStyle = damageStyle
	}
	resultLine := fmt.Sprintf("%s // COMBO %02d", run.lastResult, run.combo)
	summary := strings.Join(run.lastNotes, " · ")
	if floater, ok := run.latestFloater(); ok && floater.value > 0 {
		label := floater.label
		if label == "" || label == "HIT" {
			label = "DAMAGE"
		}
		damage := fmt.Sprintf("%s %03d", label, floater.value)
		if floater.value >= maxCombatDamage {
			damage += " MAX"
		}
		if summary != "" {
			damage += " · " + summary
		}
		summary = damage
	}
	if summary == "" {
		summary = "READ THE CARD // PRESS ENTER"
	}
	lines := []string{
		resultStyle.Render(resultLine),
		playerHealthLine(run, 18),
		mutedStyle.Render(clipPlain(summary, max(16, width-2))),
		"",
		wordHeaderLine(word),
		mutedStyle.Render(word.Phonetic),
		"",
		sectionStyle.Render("Definition"),
	}
	lines = append(lines, wrappedLimited(word.Definition, max(16, width-4), 2, textStyle)...)
	lines = append(
		lines,
		"",
		sectionStyle.Render("中文"),
		textStyle.Render(word.Meaning),
		"",
		sectionStyle.Render("Examples"),
	)
	for _, example := range word.Examples[:min(1, len(word.Examples))] {
		lines = append(lines, wrappedLimited("- "+example, max(16, width-4), 2, textStyle)...)
	}
	if run.pendingKill && run.lastLoot != "" {
		lines = append(lines, warnStyle.Render("LOOT // "+clipPlain(run.lastLoot, max(16, width-2))))
	}
	lines = append(
		lines,
		"",
		accentStyle.Render("ENTER 继续   P 暂停"),
	)
	return panelStyle.Width(width + 2).Render(strings.Join(lines, "\n"))
}

func (m Model) skillPanel(width int) string {
	run := m.run
	lines := []string{
		warnStyle.Render("LEVEL UP // 三选一技能"),
		playerHealthLine(run, 18),
		"",
	}
	lines = append(lines, m.skillLines(width)...)
	lines = append(lines, accentStyle.Render("输入 1-3 安装模块"))
	return panelStyle.Width(width + 2).Render(strings.Join(lines, "\n"))
}

func (m Model) skillLines(width int) []string {
	lines := []string{
		mutedStyle.Render("选择一个强化，继续本次 Raid"),
		"",
	}
	for index, skill := range m.run.skillChoices {
		descriptionWidth := max(8, width-lipgloss.Width(skill.Name)-8)
		lines = append(
			lines,
			accentStyle.Render(fmt.Sprintf("[%d] %s", index+1, skill.Name))+"  "+mutedStyle.Render(clipPlain(skill.Description, descriptionWidth)),
		)
	}
	lines = append(lines, "", "")
	return lines
}

func (m Model) pausedView() string {
	run := m.run
	width := min(76, max(58, m.width-8))
	lines := []string{
		titleStyle.Render("TL // PAUSED"),
		mutedStyle.Render("RUN CLOCK AND ANIMATION FROZEN"),
		"",
		textStyle.Render("ARCHIVE  ") + titleStyle.Render(run.book.Name),
		textStyle.Render("TIME     ") + accentStyle.Render(formatRemaining(run.remaining)),
		playerHealthLine(run, 22),
	}
	if run.inCombat && run.combat != nil {
		lines = append(lines, textStyle.Render("ENEMY    ")+dangerStyle.Render(run.combat.enemy.label()))
	}
	lines = append(
		lines,
		textStyle.Render(fmt.Sprintf("COMBO    %02d   WORDS %d   KILLS %d", run.combo, run.correct, run.kills)),
		"",
		accentStyle.Render("P 继续    Q 退出"),
	)
	return panelStyle.Width(width).Render(strings.Join(lines, "\n"))
}

func wordHeaderLine(word Word) string {
	header := titleStyle.Render(word.Term)
	meta := make([]string, 0, 2)
	if word.PartOfSpeech != "" {
		meta = append(meta, word.PartOfSpeech)
	}
	if word.Category != "" {
		meta = append(meta, word.Category)
	}
	if len(meta) > 0 {
		header += mutedStyle.Render(" · " + strings.Join(meta, " · "))
	}
	return header
}

func wrappedLimited(text string, width, maxLines int, style lipgloss.Style) []string {
	lines := wrapText(text, width)
	if len(lines) == 0 {
		return nil
	}
	if len(lines) > maxLines {
		lines = lines[:maxLines]
		lines[len(lines)-1] = clipPlain(lines[len(lines)-1], max(1, width-1)) + "…"
	}
	for index := range lines {
		lines[index] = style.Render(lines[index])
	}
	return lines
}

func wrapText(text string, width int) []string {
	if width <= 0 || strings.TrimSpace(text) == "" {
		return nil
	}

	lines := make([]string, 0, 3)
	for _, paragraph := range strings.Split(text, "\n") {
		words := strings.Fields(paragraph)
		if len(words) == 0 {
			continue
		}
		line := ""
		for _, word := range words {
			if lipgloss.Width(word) > width {
				if line != "" {
					lines = append(lines, line)
					line = ""
				}
				parts := splitByWidth(word, width)
				if len(parts) > 1 {
					lines = append(lines, parts[:len(parts)-1]...)
				}
				line = parts[len(parts)-1]
				continue
			}
			candidate := word
			if line != "" {
				candidate = line + " " + word
			}
			if lipgloss.Width(candidate) <= width {
				line = candidate
				continue
			}
			lines = append(lines, line)
			line = word
		}
		if line != "" {
			lines = append(lines, line)
		}
	}
	return lines
}

func splitByWidth(text string, width int) []string {
	parts := make([]string, 0, 2)
	var current strings.Builder
	used := 0
	for _, char := range text {
		charWidth := lipgloss.Width(string(char))
		if used > 0 && used+charWidth > width {
			parts = append(parts, current.String())
			current.Reset()
			used = 0
		}
		current.WriteRune(char)
		used += charWidth
	}
	if current.Len() > 0 {
		parts = append(parts, current.String())
	}
	return parts
}

func (m Model) settlementView() string {
	run := m.run
	if run == nil {
		return "no run data"
	}
	outcomeStyle := titleStyle
	subtitle := "RAID TERMINATED"
	switch run.outcome {
	case outcomeVictory:
		outcomeStyle = titleStyle
		subtitle = "BOSS PURGED // SECTOR CLEARED"
	case outcomeTimeout:
		outcomeStyle = warnStyle
		subtitle = "FIVE MINUTES ELAPSED // LINK CLOSED"
	case outcomeDefeat:
		outcomeStyle = dangerStyle
		subtitle = "PLAYER SIGNAL LOST"
	}

	total := run.correct + run.wrong
	accuracy := 0.0
	if total > 0 {
		accuracy = float64(run.correct) / float64(total) * 100
	}
	const boxInner = 46
	boxTitle := "RUN COMPLETE // " + string(run.outcome)
	lines := []string{
		outcomeStyle.Render("╔" + strings.Repeat("═", boxInner) + "╗"),
		outcomeStyle.Render("║ " + fmt.Sprintf("%-*s", boxInner-2, boxTitle) + " ║"),
		outcomeStyle.Render("╚" + strings.Repeat("═", boxInner) + "╝"),
		"",
		accentStyle.Render(subtitle),
		"",
		textStyle.Render("ARCHIVE      " + run.book.Name),
		textStyle.Render(fmt.Sprintf("ENEMIES      %d", run.kills)),
		textStyle.Render(fmt.Sprintf("CORRECT      %d", run.correct)),
		textStyle.Render(fmt.Sprintf("WRONG        %d", run.wrong)),
		textStyle.Render(fmt.Sprintf("ACCURACY     %.1f%%", accuracy)),
		textStyle.Render(fmt.Sprintf("MAX COMBO    %d", run.maxCombo)),
		textStyle.Render("WEAPON       " + weaponLabel(run)),
		textStyle.Render(fmt.Sprintf("LOOT         %d items", len(run.loot))),
		textStyle.Render("ELAPSED      " + formatElapsed(run.elapsed)),
		"",
		mutedStyle.Render("ENTER 返回词书选择    Q 退出"),
	}
	return panelStyle.Width(min(76, max(62, m.width-8))).Render(strings.Join(lines, "\n"))
}

func (r *runState) hasMonsterAt(pos point) bool {
	return r.monsterAt(pos) >= 0
}

func (r *runState) monsterKindAt(pos point) monsterKind {
	index := r.monsterAt(pos)
	if index < 0 {
		return slimeKind
	}
	return r.monsters[index].kind
}

func (r *runState) monsterGlyphAt(pos point) string {
	index := r.monsterAt(pos)
	if index < 0 {
		return "?"
	}
	return r.monsters[index].spec().glyph
}

func (r *runState) hasItemAt(pos point) bool {
	return r.itemAt(pos) >= 0
}

func monsterStyle(kind monsterKind) lipgloss.Style {
	switch kind {
	case skeletonKind:
		return skeletonStyle
	case gargoyleKind:
		return gargoyleStyle
	case bossKind:
		return bossStyle
	default:
		return slimeStyle
	}
}

func monsterLegend(run *runState) string {
	counts := map[monsterKind]int{}
	for _, monster := range run.monsters {
		counts[monster.kind]++
	}
	return slimeStyle.Render("s") + mutedStyle.Render(fmt.Sprintf("史莱姆x%d ", counts[slimeKind])) +
		skeletonStyle.Render("k") + mutedStyle.Render(fmt.Sprintf("骷髅兵x%d ", counts[skeletonKind])) +
		gargoyleStyle.Render("g") + mutedStyle.Render(fmt.Sprintf("石像鬼x%d", counts[gargoyleKind]))
}

func inventoryLine(run *runState) string {
	return "[E] 背包 · " + consumableSummary(run) + " · SHIELD " + shieldLabel(run.shield)
}

func playerHealthLine(run *runState, barWidth int) string {
	return textStyle.Render("PLAYER HP  ") +
		hpBar(run.playerHP, run.playerMaxHP, barWidth) +
		textStyle.Render(fmt.Sprintf(" %d/%d", run.playerHP, run.playerMaxHP))
}

func consumableSummary(run *runState) string {
	return fmt.Sprintf("药%d 专%d 炸%d 卷%d",
		run.inventory["health_potion"],
		run.inventory["focus_potion"],
		run.inventory["bomb"],
		run.inventory["ward_scroll"],
	)
}

func weaponLabel(run *runState) string {
	if run.equippedWeapon == "" || run.equippedWeapon == "无" {
		if run.weaponBonus > 0 {
			return fmt.Sprintf("徒手 +%d", run.weaponBonus)
		}
		return "无"
	}
	return fmt.Sprintf("%s +%d", run.equippedWeapon, run.weaponBonus)
}

func passiveSummary(run *runState) string {
	parts := make([]string, 0, 5)
	if run.critChance > 0 {
		parts = append(parts, fmt.Sprintf("CRIT %d%%", run.critChance))
	}
	if run.lifesteal > 0 {
		parts = append(parts, fmt.Sprintf("LEECH %d", run.lifesteal))
	}
	if run.comboBonus > 3 {
		parts = append(parts, fmt.Sprintf("COMBO +%d", run.comboBonus-3))
	}
	if run.memoryRing {
		parts = append(parts, fmt.Sprintf("MEMORY +%d", run.memoryBonus))
	}
	if run.guardCharm {
		parts = append(parts, "GUARD")
	}
	if len(parts) == 0 {
		return "NONE"
	}
	return strings.Join(parts, " · ")
}

func logLines(run *runState, count, width int) []string {
	start := max(0, len(run.logs)-count)
	lines := make([]string, 0, count)
	for _, entry := range run.logs[start:] {
		lines = append(lines, mutedStyle.Render("> ")+textStyle.Render(clipPlain(entry, width-2)))
	}
	for len(lines) < count {
		lines = append(lines, mutedStyle.Render(">"))
	}
	return lines
}

func clipPlain(value string, width int) string {
	if width <= 0 {
		return ""
	}
	if lipgloss.Width(value) <= width {
		return value
	}
	const suffix = "..."
	if width <= len(suffix) {
		return suffix[:width]
	}
	limit := width - len(suffix)
	var result strings.Builder
	used := 0
	for _, char := range value {
		part := string(char)
		partWidth := lipgloss.Width(part)
		if used+partWidth > limit {
			break
		}
		result.WriteString(part)
		used += partWidth
	}
	return result.String() + suffix
}

func hpBar(value, maximum, width int) string {
	if maximum <= 0 {
		return mutedStyle.Render(strings.Repeat("░", width))
	}
	filled := int(math.Round(float64(max(value, 0)) / float64(maximum) * float64(width)))
	filled = min(width, max(0, filled))
	color := neonGreen
	if filled <= width/4 {
		color = bossRed
	} else if filled <= width/2 {
		color = amber
	}
	return lipgloss.NewStyle().Foreground(color).Render(strings.Repeat("█", filled)) +
		mutedStyle.Render(strings.Repeat("░", width-filled))
}

func shieldLabel(enabled bool) string {
	if enabled {
		return accentStyle.Render("ON")
	}
	return mutedStyle.Render("OFF")
}

func formatRemaining(duration time.Duration) string {
	if duration <= 0 {
		return "00:00"
	}
	seconds := int(math.Ceil(duration.Seconds()))
	return fmt.Sprintf("%02d:%02d", seconds/60, seconds%60)
}

func formatElapsed(duration time.Duration) string {
	if duration < 0 {
		duration = 0
	}
	seconds := int(duration.Seconds())
	return fmt.Sprintf("%02d:%02d", seconds/60, seconds%60)
}
