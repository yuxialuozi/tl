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
	case screenPlaying:
		return m.playingView()
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
		mutedStyle.Render("ROGUELIKE WORD DUNGEON // 5 MINUTE RUN"),
	)

	var bookLines []string
	for index, book := range m.books {
		marker := "  "
		name := book.Name
		if index == m.selected {
			marker = "▶ "
			name = titleStyle.Render(book.Name)
		}
		description := mutedStyle.Render(book.Description)
		count := mutedStyle.Render(fmt.Sprintf("%d WORDS", len(book.Words)))
		bookLines = append(bookLines, fmt.Sprintf("%s%s  %s", marker, name, description), "    "+count, "")
	}
	menu := panelStyle.Width(width).Render(strings.Join(bookLines, "\n"))

	footer := lipgloss.JoinVertical(
		lipgloss.Left,
		"",
		textStyle.Render("↑/↓ 或 J/K 选择词书    ENTER 开始    Q 退出"),
		mutedStyle.Render("探索随机地牢 · 碰撞怪物开战 · 击败怪物掉落装备"),
	)

	return lipgloss.JoinVertical(lipgloss.Left, title, "", menu, footer)
}

func (m Model) playingView() string {
	if m.run == nil {
		return "raid state unavailable"
	}
	header := m.playingHeader()
	left := m.mapPanel()
	rightWidth := max(40, m.width-lipgloss.Width(left)-4)
	right := m.explorePanel(rightWidth)
	footer := mutedStyle.Render(" WASD / ↑↓←→ 移动 · 碰撞怪物开战 · 踩到 ? 拾取 · Q 退出")
	if m.run.inCombat {
		right = m.battlePanel(rightWidth)
		footer = mutedStyle.Render(" A-D / 1-4 答题 · H 药水 · F 专注 · X 炸弹 · R 卷轴 · Q 退出")
	}
	body := lipgloss.JoinHorizontal(lipgloss.Top, left, right)
	return lipgloss.JoinVertical(lipgloss.Left, header, body, footer)
}

func (m Model) playingHeader() string {
	run := m.run
	left := titleStyle.Render("TL // LEXICON RAID")
	right := fmt.Sprintf(
		"%s   %s   %s   %s",
		accentStyle.Render("TIME "+formatRemaining(run.remaining)),
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

	weapon := run.equippedWeapon
	if run.weaponBonus > 0 {
		weapon = fmt.Sprintf("%s +%d", weapon, run.weaponBonus)
	}
	lines = append(
		lines,
		mutedStyle.Render(strings.Repeat("─", mapWidth)),
		textStyle.Render("PLAYER  ")+hpBar(run.playerHP, run.playerMaxHP, 17),
		textStyle.Render(fmt.Sprintf("COMBO %02d   %s", run.combo, weapon)),
	)
	return panelStyle.Width(mapWidth + 2).Render(strings.Join(lines, "\n"))
}

func (m Model) explorePanel(width int) string {
	run := m.run
	lines := []string{
		accentStyle.Render("SECTOR // EXPLORATION"),
		textStyle.Render("方向键 / WASD 移动，碰到怪物会进入战斗"),
		textStyle.Render("HP ") + hpBar(run.playerHP, run.playerMaxHP, 16) + textStyle.Render(fmt.Sprintf(" %d/%d", run.playerHP, run.playerMaxHP)),
		textStyle.Render("WEAPON   ") + textStyle.Render(weaponLabel(run)),
		textStyle.Render("PASSIVE  ") + mutedStyle.Render(passiveSummary(run)),
		textStyle.Render(fmt.Sprintf("[H] 生命药水 x%d   [F] 专注药水 x%d", run.inventory["health_potion"], run.inventory["focus_potion"])),
		textStyle.Render(fmt.Sprintf("[X] 炸弹 x%d       [R] 护佑卷轴 x%d", run.inventory["bomb"], run.inventory["ward_scroll"])),
		"",
		sectionStyle.Render("MONSTERS"),
		monsterLegend(run),
		mutedStyle.Render("? 道具   @ 玩家   · 通路"),
		"",
		mutedStyle.Render(strings.Repeat("─", max(20, width-2))),
		sectionStyle.Render("SYSTEM LOG"),
	}
	lines = append(lines, logLines(run, 5)...)
	return panelStyle.Width(width + 2).Render(strings.Join(lines, "\n"))
}

func (m Model) battlePanel(width int) string {
	run := m.run
	if run.combat == nil {
		return "combat state unavailable"
	}
	enemy := run.combat.enemy
	lines := []string{
		dangerStyle.Render("ENEMY // " + enemy.label()),
		textStyle.Render(fmt.Sprintf("HP %s %d/%d", hpBar(enemy.hp, enemy.maxHP, 20), enemy.hp, enemy.maxHP)),
		textStyle.Render(enemyStatus(run.combat)),
		"",
	}

	if run.awaitingSkill {
		lines = append(lines, m.skillLines()...)
	} else {
		lines = append(lines, m.questionLines()...)
	}
	lines = append(
		lines,
		"",
		textStyle.Render(inventoryLine(run)),
		mutedStyle.Render(strings.Repeat("─", max(20, width-2))),
		sectionStyle.Render("SYSTEM LOG"),
	)
	lines = append(lines, logLines(run, 3)...)
	return panelStyle.Width(width + 2).Render(strings.Join(lines, "\n"))
}

func (m Model) questionLines() []string {
	run := m.run
	lines := []string{
		accentStyle.Render("TARGET ACQUIRED"),
		titleStyle.Render(run.question.word.Term),
		textStyle.Render("选择正确的中文释义"),
		"",
	}
	labels := []string{"A", "B", "C", "D"}
	for index, option := range run.question.options {
		lines = append(lines, optionStyle.Render(fmt.Sprintf("[%s] %s", labels[index], option)))
	}
	return lines
}

func (m Model) skillLines() []string {
	lines := []string{
		warnStyle.Render("LEVEL UP // 三选一技能"),
		textStyle.Render("输入 1-3 安装一个模块"),
		"",
	}
	for index, skill := range m.run.skillChoices {
		lines = append(
			lines,
			accentStyle.Render(fmt.Sprintf("[%d] %s", index+1, skill.Name))+"  "+mutedStyle.Render(skill.Description),
		)
	}
	lines = append(lines, "", "")
	return lines
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
	lines := []string{
		outcomeStyle.Render("╔══════════════════════════════════════════════╗"),
		outcomeStyle.Render("║   RUN COMPLETE // " + string(run.outcome) + "              ║"),
		outcomeStyle.Render("╚══════════════════════════════════════════════╝"),
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

func enemyStatus(combat *combatState) string {
	enemy := combat.enemy
	var status string
	switch enemy.kind {
	case slimeKind:
		status = "NO COUNTER"
	case skeletonKind:
		status = fmt.Sprintf("COUNTER %d", enemy.counterDamage())
	case gargoyleKind:
		status = fmt.Sprintf("WARD %d/%d · COUNTER %d", enemy.guard, enemy.spec().guard, enemy.counterDamage())
	case bossKind:
		status = fmt.Sprintf("PHASE %d/3 · WARD %d · COUNTER %d", enemy.phase, enemy.guard, enemy.counterDamage())
	}
	if combat.focusNext {
		status += " · FOCUS x2"
	}
	return status
}

func inventoryLine(run *runState) string {
	return fmt.Sprintf("H x%d  F x%d  X x%d  R x%d   SHIELD %s",
		run.inventory["health_potion"],
		run.inventory["focus_potion"],
		run.inventory["bomb"],
		run.inventory["ward_scroll"],
		shieldLabel(run.shield),
	)
}

func weaponLabel(run *runState) string {
	if run.equippedWeapon == "" || run.equippedWeapon == "无" {
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

func logLines(run *runState, count int) []string {
	start := max(0, len(run.logs)-count)
	lines := make([]string, 0, count)
	for _, entry := range run.logs[start:] {
		lines = append(lines, mutedStyle.Render("> ")+textStyle.Render(entry))
	}
	for len(lines) < count {
		lines = append(lines, mutedStyle.Render(">"))
	}
	return lines
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
