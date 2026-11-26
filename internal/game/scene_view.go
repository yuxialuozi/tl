package game

import (
	"fmt"
	"strings"
)

func (m Model) encounterPanel() string {
	run := m.run
	if run == nil || run.combat == nil {
		return panelStyle.Width(mapWidth + 2).Render(mutedStyle.Render("ENCOUNTER DATA LOST"))
	}
	enemy := run.combat.enemy
	style := monsterStyle(enemy.kind)
	if run.hitFlash > 0 {
		style = style.Bold(true).Underline(true)
	}

	lines := []string{
		titleStyle.Render("FIRST PERSON // ENCOUNTER"),
		dangerStyle.Render(enemy.spec().name),
		dangerStyle.Render("ENEMY HP  ") + hpBar(enemy.hp, enemy.maxHP, 12) + textStyle.Render(fmt.Sprintf(" %d/%d", enemy.hp, enemy.maxHP)),
		textStyle.Render(shortEnemyStatus(run.combat)),
		"",
	}
	shake := ""
	if run.hitFlash > 0 && run.hitFlash%2 == 1 {
		shake = " "
	}
	for _, line := range monsterArt(enemy.kind) {
		lines = append(lines, shake+style.Render(line))
	}
	if run.combo >= 3 {
		lines = append(lines, warnStyle.Render(fmt.Sprintf("COMBO x%d // DAMAGE SURGE", run.combo)))
	}
	if floater, ok := run.latestFloater(); ok {
		lines = append(lines, renderImpact(floater)...)
	} else {
		lines = append(lines, mutedStyle.Render("AWAITING STRIKE"))
	}
	lines = append(lines, renderFloaterFeed(run))

	panel := panelStyle.Width(mapWidth + 2)
	switch {
	case run.hurtFlash > 0:
		panel = panelHurtStyle.Width(mapWidth + 2)
	case run.hitFlash > 0:
		panel = panelHitStyle.Width(mapWidth + 2)
	}
	return panel.Render(strings.Join(lines, "\n"))
}

func (m Model) bagPanel(width int) string {
	run := m.run
	entries := run.bagEntries()
	lines := []string{
		accentStyle.Render("FIELD PACK // CONSUMABLES"),
		playerHealthLine(run, 18),
		mutedStyle.Render("选择一个道具使用，数量为 0 的道具不会显示"),
		"",
	}
	if len(entries) == 0 {
		lines = append(lines, mutedStyle.Render("背包中没有可使用的消耗品。"))
	} else {
		run.bagSelection = min(run.bagSelection, len(entries)-1)
		for index, entry := range entries {
			line := fmt.Sprintf("%s x%d  [%s]", entry.item.name, entry.count, rarityName(entry.item.rarity))
			if index == run.bagSelection {
				lines = append(lines, selectedStyle.Render("▶ "+line))
			} else {
				lines = append(lines, textStyle.Render("  "+line))
			}
			detail := entry.item.english + " · " + entry.item.description
			lines = append(lines, mutedStyle.Render("  "+clipPlain(detail, max(12, width-4))))
		}
	}
	lines = append(
		lines,
		"",
		textStyle.Render("SHIELD  ")+shieldLabel(run.shield),
		mutedStyle.Render(strings.Repeat("─", max(20, width-2))),
		sectionStyle.Render("SYSTEM LOG"),
	)
	lines = append(lines, logLines(run, 3, width)...)
	return panelStyle.Width(width + 2).Render(strings.Join(lines, "\n"))
}

func shortEnemyStatus(combat *combatState) string {
	enemy := combat.enemy
	status := "NO COUNTER"
	switch enemy.kind {
	case skeletonKind:
		status = fmt.Sprintf("COUNTER %d", enemy.counterDamage())
	case gargoyleKind:
		status = fmt.Sprintf("WARD %d/%d · HIT %d", enemy.guard, enemy.spec().guard, enemy.counterDamage())
	case bossKind:
		status = fmt.Sprintf("PHASE %d/3 · WARD %d · HIT %d", enemy.phase, enemy.guard, enemy.counterDamage())
	}
	return status
}
