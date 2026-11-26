package game

import (
	"fmt"
	"strings"

	"github.com/charmbracelet/lipgloss"
)

type floaterKind int

const (
	floaterNormal floaterKind = iota
	floaterCrit
	floaterFocus
	floaterMemory
	floaterBomb
	floaterHeal
	floaterCounter
	floaterStatus
)

type damageFloater struct {
	value int
	label string
	kind  floaterKind
}

var blockDigits = map[rune][5]string{
	'0': {"███", "█ █", "█ █", "█ █", "███"},
	'1': {" █ ", "██ ", " █ ", " █ ", "███"},
	'2': {"███", "  █", "███", "█  ", "███"},
	'3': {"███", "  █", "███", "  █", "███"},
	'4': {"█ █", "█ █", "███", "  █", "  █"},
	'5': {"███", "█  ", "███", "  █", "███"},
	'6': {"███", "█  ", "███", "█ █", "███"},
	'7': {"███", "  █", "  █", "  █", "  █"},
	'8': {"███", "█ █", "███", "█ █", "███"},
	'9': {"███", "█ █", "███", "  █", "███"},
}

var enemyArt = map[monsterKind][]string{
	slimeKind: {
		`      .-~~~-.`,
		`     /  o o  \`,
		`    |    ^    |`,
		`     \  ~~~  /`,
		`      '-...-'`,
		`       /   \`,
	},
	skeletonKind: {
		`      .-"""-.`,
		`     /  o o  \`,
		`    |    ^    |`,
		`    |   ---   |`,
		`     \  \_/  /`,
		`      '-...-'`,
	},
	gargoyleKind: {
		`       /\_/\`,
		`      / o o \`,
		`     (   ^   )`,
		`      \ '-' /`,
		`      /|   |\`,
		`     /_|___|_\`,
	},
	bossKind: {
		`      /\___/\`,
		`     /  O O  \`,
		`    |    ^    |`,
		`    |  \___/  |`,
		`     \  |||  /`,
		`      \_|||_/`,
	},
}

func (r *runState) addFloater(value int, label string, kind floaterKind) {
	r.floaters = append(r.floaters, damageFloater{
		value: value,
		label: label,
		kind:  kind,
	})
	if len(r.floaters) > 5 {
		r.floaters = r.floaters[len(r.floaters)-5:]
	}
}

func (r *runState) latestFloater() (damageFloater, bool) {
	if len(r.floaters) == 0 {
		return damageFloater{}, false
	}
	return r.floaters[len(r.floaters)-1], true
}

func monsterArt(kind monsterKind) []string {
	art := enemyArt[kind]
	result := make([]string, len(art))
	copy(result, art)
	return result
}

func floaterStyle(kind floaterKind) lipgloss.Style {
	switch kind {
	case floaterCrit:
		return critStyle
	case floaterFocus:
		return focusStyle
	case floaterMemory:
		return memoryStyle
	case floaterBomb:
		return bombStyle
	case floaterHeal:
		return healStyle
	case floaterCounter:
		return counterStyle
	case floaterStatus:
		return accentStyle
	default:
		return damageStyle
	}
}

func floaterCaption(floater damageFloater) string {
	if floater.kind == floaterStatus {
		return floater.label
	}
	caption := floater.label
	if caption == "" {
		caption = "DAMAGE"
	}
	if floater.kind == floaterHeal {
		return caption + " +"
	}
	return caption + " -"
}

func renderImpact(floater damageFloater) []string {
	style := floaterStyle(floater.kind)
	if floater.kind == floaterStatus || floater.value <= 0 {
		return []string{style.Render(floaterCaption(floater))}
	}
	value := min(floater.value, maxCombatDamage)
	if value < 20 {
		return []string{style.Render(fmt.Sprintf("%s%d", floaterCaption(floater), value))}
	}

	lines := []string{style.Render(floaterCaption(floater))}
	for _, row := range blockNumber(value, 2) {
		lines = append(lines, style.Render(row))
	}
	return lines
}

func blockNumber(value, scale int) []string {
	digits := fmt.Sprintf("%d", value)
	rows := make([]string, 5)
	for index, digit := range digits {
		glyph, ok := blockDigits[digit]
		if !ok {
			continue
		}
		for row := 0; row < 5; row++ {
			if index > 0 {
				rows[row] += strings.Repeat(" ", scale)
			}
			for _, char := range glyph[row] {
				rows[row] += strings.Repeat(string(char), scale)
			}
		}
	}
	return rows
}

func renderFloaterFeed(run *runState) string {
	if run.combatLocked {
		return accentStyle.Render("ENTER TO CONTINUE")
	}
	if len(run.floaters) <= 1 {
		return mutedStyle.Render("AWAITING IMPACT")
	}
	start := max(0, len(run.floaters)-4)
	parts := make([]string, 0, 3)
	for _, floater := range run.floaters[start : len(run.floaters)-1] {
		if floater.kind == floaterStatus {
			parts = append(parts, floaterStyle(floater.kind).Render(floater.label))
			continue
		}
		sign := "-"
		if floater.kind == floaterHeal {
			sign = "+"
		}
		parts = append(parts, floaterStyle(floater.kind).Render(fmt.Sprintf("%s%d", sign, floater.value)))
	}
	if len(parts) == 0 {
		return mutedStyle.Render("AWAITING IMPACT")
	}
	return strings.Join(parts, mutedStyle.Render("  "))
}
