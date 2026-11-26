package game

import "github.com/charmbracelet/lipgloss"

var (
	neonGreen  = lipgloss.AdaptiveColor{Light: "#0b6b4f", Dark: "#45ff9a"}
	dimGreen   = lipgloss.AdaptiveColor{Light: "#5f8f7b", Dark: "#1f7a4d"}
	cyberCyan  = lipgloss.AdaptiveColor{Light: "#086788", Dark: "#58d9ff"}
	bossRed    = lipgloss.AdaptiveColor{Light: "#c92a2a", Dark: "#ff6b74"}
	amber      = lipgloss.AdaptiveColor{Light: "#8a5a00", Dark: "#ffc857"}
	muted      = lipgloss.AdaptiveColor{Light: "#52685f", Dark: "#8aa79b"}
	foreground = lipgloss.AdaptiveColor{Light: "#173b31", Dark: "#b9f8d6"}
	wallColor  = lipgloss.AdaptiveColor{Light: "#9ab5a9", Dark: "#35544a"}
	slimeColor = lipgloss.AdaptiveColor{Light: "#147a3b", Dark: "#66ff99"}
	bonesColor = lipgloss.AdaptiveColor{Light: "#7a5600", Dark: "#ffd166"}
	stoneColor = lipgloss.AdaptiveColor{Light: "#6d3a9e", Dark: "#d2a8ff"}
	epicColor  = lipgloss.AdaptiveColor{Light: "#6d28d9", Dark: "#c792ea"}
	panelStyle = lipgloss.NewStyle().
			Border(lipgloss.NormalBorder()).
			BorderForeground(neonGreen).
			Padding(0, 1)
	titleStyle    = lipgloss.NewStyle().Foreground(neonGreen).Bold(true)
	accentStyle   = lipgloss.NewStyle().Foreground(cyberCyan).Bold(true)
	warnStyle     = lipgloss.NewStyle().Foreground(amber).Bold(true)
	dangerStyle   = lipgloss.NewStyle().Foreground(bossRed).Bold(true)
	textStyle     = lipgloss.NewStyle().Foreground(foreground)
	mutedStyle    = lipgloss.NewStyle().Foreground(muted)
	pathStyle     = lipgloss.NewStyle().Foreground(dimGreen)
	walkedStyle   = lipgloss.NewStyle().Foreground(neonGreen)
	playerStyle   = lipgloss.NewStyle().Foreground(cyberCyan).Bold(true)
	bossStyle     = lipgloss.NewStyle().Foreground(bossRed).Bold(true)
	slimeStyle    = lipgloss.NewStyle().Foreground(slimeColor).Bold(true)
	skeletonStyle = lipgloss.NewStyle().Foreground(bonesColor).Bold(true)
	gargoyleStyle = lipgloss.NewStyle().Foreground(stoneColor).Bold(true)
	lootStyle     = lipgloss.NewStyle().Foreground(epicColor).Bold(true)
	wallStyle     = lipgloss.NewStyle().Foreground(wallColor)
	optionStyle   = lipgloss.NewStyle().Foreground(foreground)
	sectionStyle  = lipgloss.NewStyle().Foreground(cyberCyan).Bold(true)
)
