package game

import (
	"context"
	"fmt"
	"math/rand"
	"strings"
	"time"

	tea "github.com/charmbracelet/bubbletea"
)

const (
	// RunDuration is the fixed length of one raid.
	RunDuration = 5 * time.Minute
	skillEvery  = 3

	monsterMoveEvery = 5
	startingLoot     = 3
	maxCombatDamage  = 999
)

type screen int

const (
	screenSelect screen = iota
	screenMode
	screenPlaying
	screenStudy
	screenSettlement
)

type runOutcome string

const (
	outcomeVictory runOutcome = "VICTORY"
	outcomeTimeout runOutcome = "TIME EXPIRED"
	outcomeDefeat  runOutcome = "DEFEAT"
)

type question struct {
	word    Word
	options [4]string
	correct int
}

type combatState struct {
	enemy      monster
	enemyIndex int
	focusNext  bool
}

type runState struct {
	book     Wordbook
	dungeon  dungeonMap
	player   point
	monsters []monster
	items    []groundItem
	sector   string

	playerHP         int
	playerMaxHP      int
	baseDamage       int
	weaponBonus      int
	comboBonus       int
	lifesteal        int
	critChance       int
	memoryBonus      int
	counterReduction int

	shield     bool
	guardCharm bool
	memoryRing bool

	equippedWeapon       string
	equippedWeaponDamage int
	inventory            map[string]int
	combo                int
	maxCombo             int
	correct              int
	wrong                int
	kills                int
	correctSinceSkill    int
	awaitingSkill        bool
	skillChoices         []Skill
	question             question
	lastTerm             string
	wrongTerms           map[string]bool
	loot                 []string

	inCombat        bool
	combat          *combatState
	monsterTicks    int
	floaters        []damageFloater
	hitFlash        int
	hurtFlash       int
	paused          bool
	clockHeldAt     time.Time
	combatLocked    bool
	lastResult      string
	lastCorrect     bool
	lastNotes       []string
	pendingKill     bool
	pendingDefeat   bool
	advanceOnUnlock bool
	lastLoot        string
	bagOpen         bool
	bagSelection    int
	startedAt       time.Time
	endsAt          time.Time
	remaining       time.Duration
	elapsed         time.Duration
	outcome         runOutcome
	logs            []string
}

// Skill is a reward selected after every three correct answers.
type Skill struct {
	Name        string
	Description string
	apply       func(*runState)
}

var skillCatalog = []Skill{
	{
		Name:        "CRIT CALIBRATOR",
		Description: "基础攻击 +6",
		apply: func(r *runState) {
			r.baseDamage += 6
		},
	},
	{
		Name:        "COMBO CORE",
		Description: "连击伤害加成 +2",
		apply: func(r *runState) {
			r.comboBonus += 2
		},
	},
	{
		Name:        "NANO REPAIR",
		Description: "立即恢复 30 HP",
		apply: func(r *runState) {
			r.playerHP = min(r.playerMaxHP, r.playerHP+30)
		},
	},
	{
		Name:        "SHIELD MATRIX",
		Description: "抵消下一次答错伤害",
		apply: func(r *runState) {
			r.shield = true
		},
	},
	{
		Name:        "OVERCLOCK",
		Description: "最大生命 +15，并恢复 15 HP",
		apply: func(r *runState) {
			r.playerMaxHP += 15
			r.playerHP = min(r.playerMaxHP, r.playerHP+15)
		},
	},
	{
		Name:        "LEECH DAEMON",
		Description: "每次答对恢复 3 HP",
		apply: func(r *runState) {
			r.lifesteal += 3
		},
	},
	{
		Name:        "PRECISION ARRAY",
		Description: "暴击率 +15%",
		apply: func(r *runState) {
			r.critChance = min(80, r.critChance+15)
		},
	},
	{
		Name:        "MEMORY CACHE",
		Description: "答错过的单词再次答对时，伤害 +10",
		apply: func(r *runState) {
			r.memoryBonus += 10
		},
	},
	{
		Name:        "WEAPON OVERDRIVE",
		Description: "武器伤害 +4",
		apply: func(r *runState) {
			r.weaponBonus += 4
		},
	},
	{
		Name:        "FIREWALL",
		Description: "受到的答错反击伤害 -4",
		apply: func(r *runState) {
			r.counterReduction = min(20, r.counterReduction+4)
		},
	},
	{
		Name:        "BATTLE TRANCE",
		Description: "下一道答对造成双倍伤害",
		apply: func(r *runState) {
			if r.combat != nil {
				r.combat.focusNext = true
			}
		},
	},
	{
		Name:        "IRON HEART",
		Description: "最大生命 +10，并立即恢复 10 HP",
		apply: func(r *runState) {
			r.playerMaxHP += 10
			r.playerHP = min(r.playerMaxHP, r.playerHP+10)
		},
	},
	{
		Name:        "VAMPIRE PROTOCOL",
		Description: "每次答对恢复 2 HP，暴击率 +5%",
		apply: func(r *runState) {
			r.lifesteal += 2
			r.critChance = min(80, r.critChance+5)
		},
	},
}

type tickMsg time.Time
type frameMsg time.Time

// Model is the Bubble Tea model for the word raid game.
type Model struct {
	screen   screen
	width    int
	height   int
	books    []Wordbook
	selected int
	mode     int
	rng      *rand.Rand
	now      func() time.Time
	run      *runState
	study    *studyState

	reviewTerms map[string]map[string]bool
}

// NewModel creates a fresh game model. The seed controls map, monsters, and words.
func NewModel(seed int64) Model {
	return Model{
		screen:      screenSelect,
		width:       112,
		height:      34,
		books:       availableWordbooks(),
		rng:         rand.New(rand.NewSource(seed)),
		now:         time.Now,
		reviewTerms: map[string]map[string]bool{},
	}
}

// Run starts the full-screen Bubble Tea game.
func Run(ctx context.Context) error {
	program := tea.NewProgram(
		NewModel(time.Now().UnixNano()),
		tea.WithAltScreen(),
		tea.WithContext(ctx),
	)
	_, err := program.Run()
	return err
}

// Init implements tea.Model.
func (m Model) Init() tea.Cmd {
	return nil
}

// Update implements tea.Model.
func (m Model) Update(message tea.Msg) (tea.Model, tea.Cmd) {
	switch message := message.(type) {
	case tea.WindowSizeMsg:
		m.width = message.Width
		m.height = message.Height
		return m, nil
	case tickMsg:
		return m.updateTick()
	case frameMsg:
		return m.updateFrame()
	case tea.KeyMsg:
		return m.updateKey(message)
	default:
		return m, nil
	}
}

func (m Model) updateKey(message tea.KeyMsg) (tea.Model, tea.Cmd) {
	key := strings.ToLower(message.String())
	if key == "ctrl+c" {
		return m, tea.Quit
	}

	switch m.screen {
	case screenSelect:
		switch key {
		case "up", "k":
			m.selected = (m.selected - 1 + len(m.books)) % len(m.books)
		case "down", "j":
			m.selected = (m.selected + 1) % len(m.books)
		case "enter", " ":
			m.mode = 0
			m.screen = screenMode
		case "q":
			return m, tea.Quit
		}
	case screenMode:
		switch key {
		case "up", "k":
			m.mode = (m.mode - 1 + len(gameModes)) % len(gameModes)
		case "down", "j":
			m.mode = (m.mode + 1) % len(gameModes)
		case "enter", " ":
			return m.startMode()
		case "esc", "backspace":
			m.screen = screenSelect
		case "q":
			return m, tea.Quit
		}
	case screenPlaying:
		if key == "q" {
			return m, tea.Quit
		}
		return m.updatePlayingKey(key)
	case screenStudy:
		return m.updateStudyKey(key)
	case screenSettlement:
		switch key {
		case "enter", "r", " ":
			m.screen = screenSelect
			m.run = nil
		case "q":
			return m, tea.Quit
		}
	}
	return m, nil
}

func (m Model) updatePlayingKey(key string) (tea.Model, tea.Cmd) {
	if m.run == nil {
		return m, nil
	}
	if key == "p" {
		m.togglePause()
		return m, nil
	}
	if m.run.paused {
		return m, nil
	}
	if m.run.awaitingSkill {
		index := -1
		switch key {
		case "1", "a":
			index = 0
		case "2", "b":
			index = 1
		case "3", "c":
			index = 2
		}
		if index >= 0 && index < len(m.run.skillChoices) {
			choice := m.run.skillChoices[index]
			choice.apply(m.run)
			m.run.awaitingSkill = false
			m.run.skillChoices = nil
			m.run.correctSinceSkill = 0
			m.run.addLog("SKILL ONLINE // " + choice.Name)
		}
		return m, nil
	}
	if m.run.bagOpen {
		return m.updateBagKey(key)
	}

	if m.run.inCombat {
		return m.updateCombatKey(key)
	}

	switch key {
	case "up", "k", "w":
		return m.movePlayer(0, -1)
	case "down", "j", "s":
		return m.movePlayer(0, 1)
	case "left", "h", "a":
		return m.movePlayer(-1, 0)
	case "right", "l", "d":
		return m.movePlayer(1, 0)
	case "e", "b", "tab":
		m.openBag()
	}
	return m, nil
}

func (m Model) updateCombatKey(key string) (tea.Model, tea.Cmd) {
	if m.run.combatLocked {
		if key == "enter" || key == " " {
			return m.resolveCombatLock()
		}
		return m, nil
	}
	switch key {
	case "a", "1":
		return m.answer(0)
	case "b", "2":
		return m.answer(1)
	case "c", "3":
		return m.answer(2)
	case "d", "4":
		return m.answer(3)
	case "e", "tab":
		m.openBag()
	default:
		return m, nil
	}
	return m, nil
}

func (m *Model) togglePause() {
	run := m.run
	if run == nil {
		return
	}
	now := m.now()
	if run.paused {
		run.paused = false
		run.resumeClock(now)
		return
	}
	run.paused = true
	run.holdClock(now)
}

func (r *runState) holdClock(now time.Time) {
	if r.clockHeldAt.IsZero() {
		r.clockHeldAt = now
	}
}

func (r *runState) resumeClock(now time.Time) {
	if r.paused || r.combatLocked || r.clockHeldAt.IsZero() {
		return
	}
	held := now.Sub(r.clockHeldAt)
	if held > 0 {
		r.startedAt = r.startedAt.Add(held)
		r.endsAt = r.endsAt.Add(held)
	}
	r.clockHeldAt = time.Time{}
}

func (m Model) startRun() (tea.Model, tea.Cmd) {
	book := m.books[m.selected]
	dungeon, err := generateDungeon(m.rng)
	if err != nil {
		m.screen = screenSettlement
		m.run = &runState{outcome: outcomeDefeat, logs: []string{"MAP GENERATION FAILED", err.Error()}}
		return m, nil
	}

	now := m.now()
	run := &runState{
		book:           book,
		dungeon:        dungeon,
		player:         dungeon.start,
		sector:         fmt.Sprintf("SECTOR // %X", 0x1000+m.rng.Intn(0xEFFF)),
		playerMaxHP:    100,
		playerHP:       100,
		baseDamage:     24,
		comboBonus:     3,
		inventory:      map[string]int{},
		wrongTerms:     map[string]bool{},
		startedAt:      now,
		endsAt:         now.Add(RunDuration),
		remaining:      RunDuration,
		equippedWeapon: "无",
		logs:           []string{"RAID LINK ESTABLISHED", "ARCHIVE: " + book.Name},
	}
	m.spawnMonsters(run)
	m.spawnItems(run)
	m.run = run
	m.screen = screenPlaying
	return m, tea.Batch(tickCmd(), frameCmd())
}

func (m Model) spawnMonsters(run *runState) {
	occupied := map[point]bool{run.player: true}
	boss := newMonster(bossKind, run.dungeon.boss)
	run.monsters = append(run.monsters, boss)
	occupied[boss.pos] = true

	kinds := []monsterKind{
		slimeKind, slimeKind,
		skeletonKind,
		gargoyleKind,
	}
	for _, kind := range kinds {
		pos, ok := m.freeCell(run, occupied, 5)
		if !ok {
			continue
		}
		run.monsters = append(run.monsters, newMonster(kind, pos))
		occupied[pos] = true
	}
}

func (m Model) spawnItems(run *runState) {
	occupied := map[point]bool{run.player: true}
	for _, monster := range run.monsters {
		occupied[monster.pos] = true
	}
	for i := 0; i < startingLoot; i++ {
		pos, ok := m.freeCell(run, occupied, 8)
		if !ok {
			break
		}
		run.items = append(run.items, groundItem{pos: pos, item: randomGroundItem(m.rng)})
		occupied[pos] = true
	}
}

func (m Model) freeCell(run *runState, occupied map[point]bool, minDistance int) (point, bool) {
	cells := run.dungeon.floorCells()
	for _, index := range m.rng.Perm(len(cells)) {
		cell := cells[index]
		if occupied[cell] {
			continue
		}
		if manhattan(cell, run.player) < minDistance {
			continue
		}
		return cell, true
	}
	return point{}, false
}

func tickCmd() tea.Cmd {
	return tea.Tick(time.Second, func(now time.Time) tea.Msg {
		return tickMsg(now)
	})
}

func frameCmd() tea.Cmd {
	return tea.Tick(80*time.Millisecond, func(now time.Time) tea.Msg {
		return frameMsg(now)
	})
}

func (m Model) updateFrame() (tea.Model, tea.Cmd) {
	if m.screen != screenPlaying || m.run == nil {
		return m, nil
	}
	if m.run.paused {
		return m, frameCmd()
	}
	m.run.hitFlash = max(0, m.run.hitFlash-1)
	m.run.hurtFlash = max(0, m.run.hurtFlash-1)
	return m, frameCmd()
}

func (m Model) updateTick() (tea.Model, tea.Cmd) {
	if m.screen != screenPlaying || m.run == nil {
		return m, nil
	}
	if m.run.paused || m.run.combatLocked {
		return m, tickCmd()
	}
	m.run.remaining = m.run.endsAt.Sub(m.now())
	if m.run.remaining <= 0 {
		m.run.remaining = 0
		m.settle(outcomeTimeout)
		return m, nil
	}

	if !m.run.inCombat {
		m.run.monsterTicks++
		if m.run.monsterTicks >= monsterMoveEvery {
			m.run.monsterTicks = 0
			if index := m.moveMonsters(); index >= 0 {
				m = m.beginCombat(index)
			}
		}
	}
	return m, tickCmd()
}

func (m *Model) moveMonsters() int {
	run := m.run
	occupied := map[point]bool{run.player: true}
	for _, monster := range run.monsters {
		occupied[monster.pos] = true
	}

	for index := range run.monsters {
		monster := &run.monsters[index]
		delete(occupied, monster.pos)
		next := m.monsterStep(*monster, occupied)
		monster.pos = next
		occupied[next] = true
		if next == run.player {
			return index
		}
	}
	return -1
}

func (m *Model) monsterStep(monster monster, occupied map[point]bool) point {
	directions := []point{{x: 0, y: -1}, {x: 1, y: 0}, {x: 0, y: 1}, {x: -1, y: 0}}
	m.rng.Shuffle(len(directions), func(i, j int) {
		directions[i], directions[j] = directions[j], directions[i]
	})

	candidates := make([]point, 0, 4)
	for _, direction := range directions {
		next := point{x: monster.pos.x + direction.x, y: monster.pos.y + direction.y}
		if !m.run.dungeon.walkable(next) {
			continue
		}
		if occupied[next] && next != m.run.player {
			continue
		}
		candidates = append(candidates, next)
	}
	if len(candidates) == 0 {
		return monster.pos
	}

	if manhattan(monster.pos, m.run.player) <= 6 && m.rng.Intn(100) < 40 {
		best := candidates[0]
		for _, candidate := range candidates[1:] {
			if manhattan(candidate, m.run.player) < manhattan(best, m.run.player) {
				best = candidate
			}
		}
		return best
	}
	return candidates[m.rng.Intn(len(candidates))]
}

func (m Model) movePlayer(dx, dy int) (tea.Model, tea.Cmd) {
	run := m.run
	next := point{x: run.player.x + dx, y: run.player.y + dy}
	if !run.dungeon.walkable(next) {
		return m, nil
	}
	if index := run.monsterAt(next); index >= 0 {
		return m.beginCombat(index), nil
	}

	run.player = next
	if index := run.itemAt(next); index >= 0 {
		item := run.items[index].item
		run.items = append(run.items[:index], run.items[index+1:]...)
		result := run.addItem(item)
		run.addLog(fmt.Sprintf("PICKUP // %s [%s] · %s", item.name, rarityName(item.rarity), result))
	}
	return m, nil
}

func (m Model) beginCombat(index int) Model {
	run := m.run
	if index < 0 || index >= len(run.monsters) {
		return m
	}
	run.inCombat = true
	run.combat = &combatState{enemy: run.monsters[index], enemyIndex: index}
	run.shield = run.shield || run.guardCharm
	run.bagOpen = false
	run.combatLocked = false
	run.pendingKill = false
	run.pendingDefeat = false
	run.advanceOnUnlock = false
	run.lastResult = ""
	run.lastNotes = nil
	run.lastLoot = ""
	run.floaters = nil
	run.hitFlash = 0
	run.hurtFlash = 0
	run.addLog("ENCOUNTER // " + run.monsters[index].label())
	if run.shield {
		run.addLog("GUARD CHARM ONLINE // 1 次答错保护")
	}
	m.nextQuestion()
	return m
}

func (m Model) answer(index int) (tea.Model, tea.Cmd) {
	run := m.run
	if run == nil || run.combat == nil || run.combatLocked || run.paused || run.awaitingSkill || run.bagOpen {
		return m, nil
	}
	if index < 0 || index >= len(run.question.options) {
		return m, nil
	}

	current := run.question
	if index == current.correct {
		return m.correctAnswer(current)
	}
	return m.wrongAnswer(current)
}

func (m Model) correctAnswer(current question) (tea.Model, tea.Cmd) {
	run := m.run
	combat := run.combat
	enemy := &combat.enemy

	run.correct++
	run.correctSinceSkill++
	run.combo++
	run.maxCombo = max(run.maxCombo, run.combo)

	damage := run.baseDamage + run.weaponBonus + run.combo*run.comboBonus
	notes := make([]string, 0, 3)
	impact := damageFloater{value: damage, label: "HIT", kind: floaterNormal}
	if run.memoryBonus > 0 && run.wrongTerms[current.word.Term] {
		damage += run.memoryBonus
		delete(run.wrongTerms, current.word.Term)
		notes = append(notes, fmt.Sprintf("MEMORY +%d", run.memoryBonus))
		impact.kind = floaterMemory
		impact.label = "MEMORY"
	}
	if combat.focusNext {
		damage *= 2
		combat.focusNext = false
		notes = append(notes, "FOCUS x2")
		impact.kind = floaterFocus
		impact.label = "FOCUS"
	}
	if run.critChance > 0 && m.rng.Intn(100) < run.critChance {
		damage *= 2
		notes = append(notes, "CRIT x2")
		impact.kind = floaterCrit
		impact.label = "CRIT"
	}
	if damage > maxCombatDamage {
		notes = append(notes, fmt.Sprintf("CAP %d", maxCombatDamage))
	}
	damage = min(damage, maxCombatDamage)
	impact.value = damage

	if enemy.guard > 0 {
		enemy.guard--
		run.hitFlash = 3
		if enemy.guard > 0 {
			run.addLog(fmt.Sprintf("WARD HIT // %s · 还需连续答对 %d 次", current.word.Term, enemy.guard))
			run.addFloater(0, fmt.Sprintf("WARD %d", enemy.guard), floaterStatus)
		} else {
			run.addLog("WARD BROKEN // 防御已解除")
			run.addFloater(0, "WARD BROKEN", floaterStatus)
		}
		run.lastNotes = notes
		run.advanceOnUnlock = true
		m.lockCombatResult("WARD HELD", true)
		return m, nil
	}

	killed := m.damageEnemy(damage)
	run.addFloater(impact.value, impact.label, impact.kind)
	run.hitFlash = 4
	if run.lifesteal > 0 {
		run.playerHP = min(run.playerMaxHP, run.playerHP+run.lifesteal)
	}
	suffix := ""
	if len(notes) > 0 {
		suffix = " [" + strings.Join(notes, " · ") + "]"
	}
	run.addLog(fmt.Sprintf("HIT %s -%d HP COMBO x%d%s", current.word.Term, damage, run.combo, suffix))
	run.lastNotes = notes
	run.advanceOnUnlock = true

	result := "PERFECT ANSWER"
	if killed {
		m.registerKill()
		result = "TARGET DESTROYED"
	}
	m.lockCombatResult(result, true)
	return m, nil
}

func (m *Model) lockCombatResult(result string, correct bool) {
	run := m.run
	if run == nil {
		return
	}
	run.combatLocked = true
	run.lastResult = result
	run.lastCorrect = correct
	run.holdClock(m.now())
}

func (m *Model) clearCombatResult() {
	run := m.run
	if run == nil {
		return
	}
	run.combatLocked = false
	run.pendingKill = false
	run.pendingDefeat = false
	run.advanceOnUnlock = false
	run.lastResult = ""
	run.lastNotes = nil
	run.lastLoot = ""
	run.floaters = nil
	run.hitFlash = 0
	run.hurtFlash = 0
	run.resumeClock(m.now())
}

func (m Model) resolveCombatLock() (tea.Model, tea.Cmd) {
	run := m.run
	if run == nil {
		return m, nil
	}
	if run.pendingDefeat {
		m.clearCombatResult()
		m.settle(outcomeDefeat)
		return m, nil
	}

	advance := run.advanceOnUnlock
	pendingKill := run.pendingKill
	m.clearCombatResult()
	if pendingKill {
		return m.finishCombat()
	}
	if !advance || run.combat == nil {
		return m, nil
	}

	m.nextQuestion()
	if run.correctSinceSkill >= skillEvery {
		m.offerSkills()
	}
	return m, nil
}

func (m Model) wrongAnswer(current question) (tea.Model, tea.Cmd) {
	run := m.run
	combat := run.combat
	enemy := &combat.enemy

	run.wrong++
	run.combo = 0
	run.correctSinceSkill = 0
	run.wrongTerms[current.word.Term] = true

	if enemy.guard > 0 {
		enemy.guard = enemy.spec().guard
		if enemy.kind == bossKind && enemy.phase > 1 {
			enemy.guard = 2
		}
		run.addLog("WARD RESET // 连续答对记录已清空")
	}

	baseCounter := enemy.counterDamage()
	damage := max(1, baseCounter-run.counterReduction)
	result := "MISS"
	switch {
	case baseCounter == 0:
		run.addLog(fmt.Sprintf("MISS %s = %s · 史莱姆没有反击", current.word.Term, current.word.Meaning))
		run.addFloater(0, "NO COUNTER", floaterStatus)
		result = "NO COUNTER"
	case run.shield:
		run.shield = false
		run.addLog(fmt.Sprintf("GUARD ABSORBED // MISS %s", current.word.Term))
		run.addFloater(0, "ABSORBED", floaterStatus)
		result = "GUARD ABSORBED"
	default:
		run.playerHP = max(0, run.playerHP-damage)
		run.addLog(fmt.Sprintf("COUNTER -%d HP // MISS %s = %s", damage, current.word.Term, current.word.Meaning))
		run.addFloater(damage, "COUNTER", floaterCounter)
		run.hurtFlash = 4
		result = "COUNTER HIT"
	}

	run.advanceOnUnlock = true
	if run.playerHP == 0 {
		run.pendingDefeat = true
		result = "SIGNAL LOST"
	}
	m.lockCombatResult(result, false)
	return m, nil
}

func (m *Model) damageEnemy(damage int) bool {
	combat := m.run.combat
	enemy := &combat.enemy
	enemy.hp = max(0, enemy.hp-damage)
	if enemy.hp == 0 {
		return true
	}
	if enemy.kind == bossKind {
		phase := enemy.phaseForHP()
		if phase > enemy.phase {
			enemy.phase = phase
			enemy.guard = 2
			m.run.addLog(fmt.Sprintf("BOSS PHASE %d // 反击伤害提升，护盾需要连续答对 2 次", phase))
		}
	}
	return false
}

func (m *Model) registerKill() {
	run := m.run
	if run == nil || run.combat == nil || run.pendingKill {
		return
	}
	enemy := run.combat.enemy
	run.kills++
	run.pendingKill = true

	drop := randomItem(m.rng, enemy.kind)
	result := run.addItem(drop)
	lootLine := fmt.Sprintf("%s [%s] · %s", drop.name, rarityName(drop.rarity), result)
	run.loot = append(run.loot, lootLine)
	run.lastLoot = lootLine
	run.addLog("LOOT // " + lootLine)
	if enemy.kind == bossKind {
		run.addLog("BOSS PURGED // SECTOR CLEARED")
	}
}

func (m Model) finishCombat() (tea.Model, tea.Cmd) {
	run := m.run
	if run.combat == nil {
		return m, nil
	}
	enemy := run.combat.enemy

	if enemy.kind == bossKind {
		m.settle(outcomeVictory)
		return m, nil
	}

	if run.combat.enemyIndex >= 0 && run.combat.enemyIndex < len(run.monsters) {
		run.monsters = append(run.monsters[:run.combat.enemyIndex], run.monsters[run.combat.enemyIndex+1:]...)
	}
	run.addLog(fmt.Sprintf("DEFEATED // %s · 返回地图", enemy.label()))
	run.inCombat = false
	run.combat = nil
	run.bagOpen = false
	run.question = question{}
	run.monsterTicks = 0
	if run.correctSinceSkill >= skillEvery {
		m.offerSkills()
	}
	return m, nil
}

func (m Model) useConsumable(id string) (tea.Model, tea.Cmd) {
	run := m.run
	if run == nil || run.inventory[id] <= 0 {
		if run != nil {
			run.addLog("NO ITEM // 背包中没有该道具")
		}
		return m, nil
	}
	item, ok := itemByID(id)
	if !ok {
		return m, nil
	}
	if run.combat == nil && (item.focus || item.directDamage > 0) {
		run.addLog("COMBAT ONLY // 该道具只能在战斗中使用")
		return m, nil
	}

	switch {
	case item.heal > 0:
		if run.playerHP >= run.playerMaxHP {
			run.addLog("HP FULL // 生命药水未消耗")
			return m, nil
		}
		run.inventory[id]--
		run.playerHP = min(run.playerMaxHP, run.playerHP+item.heal)
		run.addFloater(item.heal, "HEAL", floaterHeal)
		run.bagOpen = false
		run.addLog(fmt.Sprintf("POTION // 恢复 %d HP", item.heal))
	case item.focus:
		if run.combat.focusNext {
			run.addLog("FOCUS ACTIVE // 专注药水未消耗")
			return m, nil
		}
		run.inventory[id]--
		run.combat.focusNext = true
		run.addFloater(0, "FOCUS x2", floaterStatus)
		run.bagOpen = false
		run.addLog("POTION // 下一道答对造成双倍伤害")
	case item.directDamage > 0:
		run.inventory[id]--
		damage := min(item.directDamage, maxCombatDamage)
		killed := m.damageEnemy(damage)
		run.addFloater(damage, "BOMB", floaterBomb)
		run.hitFlash = 5
		run.bagOpen = false
		run.addLog(fmt.Sprintf("BOMB // 对敌人造成 %d 点伤害", damage))
		result := "BOMB STRIKE"
		if killed {
			m.registerKill()
			result = "TARGET DESTROYED"
		}
		run.advanceOnUnlock = false
		m.lockCombatResult(result, true)
	case item.shield:
		if run.shield {
			run.addLog("GUARD ACTIVE // 护佑卷轴未消耗")
			return m, nil
		}
		run.inventory[id]--
		run.shield = true
		run.addFloater(0, "WARD ON", floaterStatus)
		run.bagOpen = false
		run.addLog("SCROLL // 下一次答错伤害将被抵消")
	}
	return m, nil
}

type bagEntry struct {
	item  itemSpec
	count int
}

func (r *runState) bagEntries() []bagEntry {
	entries := make([]bagEntry, 0, len(itemCatalog))
	for _, item := range itemCatalog {
		if item.category != categoryConsumable || r.inventory[item.id] <= 0 {
			continue
		}
		entries = append(entries, bagEntry{item: item, count: r.inventory[item.id]})
	}
	return entries
}

func (m *Model) openBag() {
	if m.run == nil {
		return
	}
	m.run.bagOpen = true
	m.run.bagSelection = min(m.run.bagSelection, max(0, len(m.run.bagEntries())-1))
}

func (m Model) updateBagKey(key string) (tea.Model, tea.Cmd) {
	run := m.run
	entries := run.bagEntries()
	switch key {
	case "esc", "e", "tab":
		run.bagOpen = false
	case "up", "k":
		if len(entries) > 0 {
			run.bagSelection = (run.bagSelection - 1 + len(entries)) % len(entries)
		}
	case "down", "j":
		if len(entries) > 0 {
			run.bagSelection = (run.bagSelection + 1) % len(entries)
		}
	case "enter", " ":
		if len(entries) == 0 {
			run.addLog("BAG EMPTY // 没有可使用的消耗品")
			return m, nil
		}
		run.bagSelection = min(run.bagSelection, len(entries)-1)
		return m.useConsumable(entries[run.bagSelection].item.id)
	}
	return m, nil
}

func (m *Model) nextQuestion() {
	run := m.run
	if len(run.book.Words) == 0 {
		return
	}

	word, picked := m.pickQuestionWord()
	if !picked {
		return
	}
	word.Term = strings.ToLower(word.Term)
	run.lastTerm = word.Term

	distractors := make([]string, 0, 3)
	seen := map[string]bool{word.Meaning: true}
	for attempts := 0; len(distractors) < 3 && attempts < 200; attempts++ {
		candidate := run.book.Words[m.rng.Intn(len(run.book.Words))].Meaning
		if !seen[candidate] {
			seen[candidate] = true
			distractors = append(distractors, candidate)
		}
	}
	if len(distractors) < 3 {
		return
	}

	options := make([]string, 0, 4)
	correct := m.rng.Intn(4)
	for i := 0; i < 4; i++ {
		if i == correct {
			options = append(options, word.Meaning)
		} else {
			options = append(options, distractors[0])
			distractors = distractors[1:]
		}
	}
	run.question = question{
		word:    word,
		options: [4]string{options[0], options[1], options[2], options[3]},
		correct: correct,
	}
}

func (m *Model) pickQuestionWord() (Word, bool) {
	run := m.run
	words := run.book.Words
	if len(words) == 0 {
		return Word{}, false
	}

	if len(run.wrongTerms) > 0 && m.rng.Intn(100) < 35 {
		review := make([]Word, 0, len(run.wrongTerms))
		for _, word := range words {
			if run.wrongTerms[strings.ToLower(word.Term)] {
				review = append(review, word)
			}
		}
		if len(review) > 0 {
			return review[m.rng.Intn(len(review))], true
		}
	}

	word := words[m.rng.Intn(len(words))]
	for attempts := 0; attempts < 8 && strings.EqualFold(word.Term, run.lastTerm) && len(words) > 1; attempts++ {
		word = words[m.rng.Intn(len(words))]
	}
	return word, true
}

func (m *Model) offerSkills() {
	permutation := m.rng.Perm(len(skillCatalog))
	choices := make([]Skill, 0, 3)
	for _, index := range permutation[:3] {
		choices = append(choices, skillCatalog[index])
	}
	m.run.skillChoices = choices
	m.run.awaitingSkill = true
}

func (m *Model) settle(outcome runOutcome) {
	if m.run == nil {
		return
	}
	now := m.now()
	m.run.outcome = outcome
	m.run.elapsed = now.Sub(m.run.startedAt)
	if outcome == outcomeTimeout {
		m.run.remaining = 0
	} else if m.run.remaining <= 0 {
		m.run.remaining = max(0, m.run.endsAt.Sub(now))
	}
	m.screen = screenSettlement
}

func (r *runState) addLog(line string) {
	r.logs = append(r.logs, line)
	if len(r.logs) > 8 {
		r.logs = r.logs[len(r.logs)-8:]
	}
}

func (r *runState) monsterAt(pos point) int {
	for index, monster := range r.monsters {
		if monster.pos == pos {
			return index
		}
	}
	return -1
}

func (r *runState) itemAt(pos point) int {
	for index, item := range r.items {
		if item.pos == pos {
			return index
		}
	}
	return -1
}

func manhattan(a, b point) int {
	return abs(a.x-b.x) + abs(a.y-b.y)
}

func abs(value int) int {
	if value < 0 {
		return -value
	}
	return value
}
