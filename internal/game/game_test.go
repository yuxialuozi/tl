package game

import (
	"math/rand"
	"strings"
	"testing"
	"time"

	tea "github.com/charmbracelet/bubbletea"
)

func TestWordbooksHaveUniqueChoices(t *testing.T) {
	books := availableWordbooks()
	if len(books) < 3 {
		t.Fatalf("wordbooks = %d, want at least 3", len(books))
	}
	for _, book := range books {
		if len(book.Words) < 4 {
			t.Fatalf("%s has only %d words", book.Name, len(book.Words))
		}
		seen := make(map[string]bool)
		for _, word := range book.Words {
			if word.Term == "" || word.Meaning == "" {
				t.Fatalf("%s contains an empty word entry", book.Name)
			}
			if word.Term != strings.ToLower(word.Term) {
				t.Fatalf("%s word %q is not lowercase", book.Name, word.Term)
			}
			if seen[word.Meaning] {
				t.Fatalf("%s contains duplicate meaning %q", book.Name, word.Meaning)
			}
			seen[word.Meaning] = true
		}
	}
}

func TestDungeonGenerationProducesConnectedRoute(t *testing.T) {
	dungeon, err := generateDungeon(rand.New(rand.NewSource(42)))
	if err != nil {
		t.Fatalf("generateDungeon() error = %v", err)
	}
	if len(dungeon.path) < 8 {
		t.Fatalf("path length = %d", len(dungeon.path))
	}
	if dungeon.path[0] != dungeon.start {
		t.Fatalf("path start = %+v, want %+v", dungeon.path[0], dungeon.start)
	}
	if dungeon.path[len(dungeon.path)-1] != dungeon.boss {
		t.Fatalf("path end = %+v, want %+v", dungeon.path[len(dungeon.path)-1], dungeon.boss)
	}
	for _, p := range dungeon.path {
		if dungeon.grid[p.y][p.x] == '#' {
			t.Fatalf("path enters wall at %+v", p)
		}
	}
	if len(dungeon.floorCells()) < 20 {
		t.Fatalf("floor cells = %d", len(dungeon.floorCells()))
	}
}

func startedModel(t *testing.T) Model {
	t.Helper()
	now := time.Date(2026, 11, 26, 12, 0, 0, 0, time.Local)
	model := NewModel(7)
	model.now = func() time.Time { return now }
	updated, _ := model.startRun()
	started, ok := updated.(Model)
	if !ok {
		t.Fatal("startRun() returned a different model type")
	}
	return started
}

func combatModel(t *testing.T, kind monsterKind) Model {
	t.Helper()
	model := startedModel(t)
	model.run.monsters = []monster{newMonster(kind, model.run.player)}
	return model.beginCombat(0)
}

func updateModel(t *testing.T, model Model, message tea.Msg) Model {
	t.Helper()
	updated, _ := model.Update(message)
	result, ok := updated.(Model)
	if !ok {
		t.Fatal("Update() returned a different model type")
	}
	return result
}

func TestStartRunCreatesExplorationState(t *testing.T) {
	model := startedModel(t)
	if model.screen != screenPlaying || model.run == nil {
		t.Fatalf("screen = %d, run = %+v", model.screen, model.run)
	}
	if model.run.remaining != RunDuration {
		t.Fatalf("remaining = %v", model.run.remaining)
	}
	if model.run.inCombat || model.run.combat != nil {
		t.Fatal("run should start in exploration mode")
	}
	if model.run.player != model.run.dungeon.start {
		t.Fatalf("player = %+v, start = %+v", model.run.player, model.run.dungeon.start)
	}
	if model.run.playerHP != model.run.playerMaxHP {
		t.Fatalf("player HP = %d/%d", model.run.playerHP, model.run.playerMaxHP)
	}
	if len(model.run.monsters) < 7 {
		t.Fatalf("monsters = %d", len(model.run.monsters))
	}
	if len(model.run.items) != startingLoot {
		t.Fatalf("items = %d", len(model.run.items))
	}
	bosses := 0
	for _, monster := range model.run.monsters {
		if monster.kind == bossKind {
			bosses++
		}
		if monster.pos == model.run.player {
			t.Fatalf("monster spawned on player at %+v", monster.pos)
		}
	}
	if bosses != 1 {
		t.Fatalf("boss count = %d", bosses)
	}
}

func TestCombatQuestionIsValidAndLowercase(t *testing.T) {
	model := combatModel(t, slimeKind)
	if !model.run.inCombat || model.run.combat == nil {
		t.Fatal("combat did not start")
	}
	options := model.run.question.options
	seen := make(map[string]bool)
	for _, option := range options {
		if option == "" || seen[option] {
			t.Fatalf("options are not unique: %v", options)
		}
		seen[option] = true
	}
	if options[model.run.question.correct] != model.run.question.word.Meaning {
		t.Fatalf("correct option = %q, word meaning = %q", options[model.run.question.correct], model.run.question.word.Meaning)
	}
	if model.run.question.word.Term != strings.ToLower(model.run.question.word.Term) {
		t.Fatalf("target term = %q", model.run.question.word.Term)
	}
}

func TestCorrectAnswerDamagesMonster(t *testing.T) {
	model := combatModel(t, slimeKind)
	before := model.run.combat.enemy.hp
	correct := model.run.question.correct
	updated, _ := model.answer(correct)
	attacked := updated.(Model)

	if attacked.run.combat.enemy.hp >= before {
		t.Fatalf("enemy HP = %d, want less than %d", attacked.run.combat.enemy.hp, before)
	}
	if attacked.run.correct != 1 || attacked.run.combo != 1 || attacked.run.maxCombo != 1 {
		t.Fatalf("stats = correct:%d combo:%d max:%d", attacked.run.correct, attacked.run.combo, attacked.run.maxCombo)
	}
}

func TestSlimeDoesNotCounterattack(t *testing.T) {
	model := combatModel(t, slimeKind)
	before := model.run.playerHP
	wrong := (model.run.question.correct + 1) % len(model.run.question.options)
	updated, _ := model.answer(wrong)
	attacked := updated.(Model)
	if attacked.run.playerHP != before {
		t.Fatalf("player HP = %d, want %d", attacked.run.playerHP, before)
	}
	if attacked.run.wrong != 1 {
		t.Fatalf("wrong = %d", attacked.run.wrong)
	}
}

func TestSkeletonCounterattacksOnWrongAnswer(t *testing.T) {
	model := combatModel(t, skeletonKind)
	model.run.combo = 4
	before := model.run.playerHP
	wrong := (model.run.question.correct + 1) % len(model.run.question.options)
	updated, _ := model.answer(wrong)
	attacked := updated.(Model)

	if attacked.run.playerHP >= before {
		t.Fatalf("player HP = %d, want less than %d", attacked.run.playerHP, before)
	}
	if attacked.run.combo != 0 || attacked.run.wrong != 1 {
		t.Fatalf("stats = combo:%d wrong:%d", attacked.run.combo, attacked.run.wrong)
	}
}

func TestGargoyleRequiresConsecutiveCorrectAnswers(t *testing.T) {
	model := combatModel(t, gargoyleKind)
	before := model.run.combat.enemy.hp

	for i := 0; i < 2; i++ {
		correct := model.run.question.correct
		updated, _ := model.answer(correct)
		model = updated.(Model)
		if model.run.combat.enemy.hp != before {
			t.Fatalf("enemy HP changed before ward broke: %d", model.run.combat.enemy.hp)
		}
	}
	if model.run.combat.enemy.guard != 0 {
		t.Fatalf("guard = %d, want 0", model.run.combat.enemy.guard)
	}

	correct := model.run.question.correct
	updated, _ := model.answer(correct)
	model = updated.(Model)
	if model.run.combat.enemy.hp >= before {
		t.Fatalf("enemy HP = %d, want less than %d", model.run.combat.enemy.hp, before)
	}
}

func TestGargoyleWardResetsOnWrongAnswer(t *testing.T) {
	model := combatModel(t, gargoyleKind)
	model.run.combat.enemy.guard = 1
	before := model.run.playerHP
	wrong := (model.run.question.correct + 1) % len(model.run.question.options)
	updated, _ := model.answer(wrong)
	attacked := updated.(Model)
	if attacked.run.combat.enemy.guard != 2 {
		t.Fatalf("guard = %d, want 2", attacked.run.combat.enemy.guard)
	}
	if attacked.run.playerHP >= before {
		t.Fatal("gargoyle did not counterattack")
	}
}

func TestBossEntersNewPhaseWithWard(t *testing.T) {
	model := combatModel(t, bossKind)
	model.run.combat.enemy.hp = 245
	if model.run.combat.enemy.phaseForHP() != 1 {
		t.Fatalf("phase before hit = %d", model.run.combat.enemy.phaseForHP())
	}
	correct := model.run.question.correct
	updated, _ := model.answer(correct)
	attacked := updated.(Model)
	enemy := attacked.run.combat.enemy
	if enemy.phase != 2 || enemy.guard != 2 {
		t.Fatalf("phase = %d guard = %d", enemy.phase, enemy.guard)
	}
	if enemy.counterDamage() <= monsterSpecs[bossKind].counter {
		t.Fatalf("boss counter = %d", enemy.counterDamage())
	}
}

func TestThreeCorrectAnswersOfferSkills(t *testing.T) {
	model := combatModel(t, bossKind)
	for i := 0; i < skillEvery; i++ {
		correct := model.run.question.correct
		updated, _ := model.answer(correct)
		model = updated.(Model)
	}
	if !model.run.awaitingSkill {
		t.Fatal("awaitingSkill = false")
	}
	if len(model.run.skillChoices) != 3 {
		t.Fatalf("skill choices = %d", len(model.run.skillChoices))
	}
	selected := model.run.skillChoices[0].Name
	updated, _ := model.updatePlayingKey("1")
	model = updated.(Model)
	if model.run.awaitingSkill {
		t.Fatal("awaitingSkill still true after selection")
	}
	if !strings.Contains(strings.Join(model.run.logs, "\n"), selected) {
		t.Fatalf("skill log missing %q: %v", selected, model.run.logs)
	}
}

func TestShieldAbsorbsWrongAnswer(t *testing.T) {
	model := combatModel(t, skeletonKind)
	model.run.shield = true
	before := model.run.playerHP
	wrong := (model.run.question.correct + 1) % len(model.run.question.options)
	updated, _ := model.answer(wrong)
	attacked := updated.(Model)
	if attacked.run.playerHP != before {
		t.Fatalf("player HP = %d, want %d", attacked.run.playerHP, before)
	}
	if attacked.run.shield {
		t.Fatal("shield still active")
	}
}

func TestItemsCoverRequiredCategories(t *testing.T) {
	categories := map[itemCategory]bool{}
	for _, item := range itemCatalog {
		categories[item.category] = true
	}
	for _, category := range []itemCategory{categoryConsumable, categoryWeapon, categoryAccessory, categoryPassive} {
		if !categories[category] {
			t.Fatalf("missing item category %d", category)
		}
	}

	effects := map[string]bool{}
	for _, item := range itemCatalog {
		switch item.id {
		case "health_potion":
			effects["heal30"] = item.heal == 30
		case "focus_potion":
			effects["focus"] = item.focus
		case "sharp_sword":
			effects["sharp"] = item.weaponDamage == 5
		case "memory_ring":
			effects["memory"] = item.memoryBonus == 20
		}
	}
	for _, effect := range []string{"heal30", "focus", "sharp", "memory"} {
		if !effects[effect] {
			t.Fatalf("missing item effect %s", effect)
		}
	}
}

func TestHealthPotionRestoresThirtyHP(t *testing.T) {
	model := combatModel(t, skeletonKind)
	model.run.playerHP = 50
	model.run.inventory["health_potion"] = 1
	updated, _ := model.useConsumable("health_potion")
	used := updated.(Model)
	if used.run.playerHP != 80 {
		t.Fatalf("player HP = %d, want 80", used.run.playerHP)
	}
	if used.run.inventory["health_potion"] != 0 {
		t.Fatal("potion was not consumed")
	}
}

func TestFocusPotionDoublesNextCorrectHit(t *testing.T) {
	model := combatModel(t, bossKind)
	model.run.combat.enemy.hp = 300
	model.run.inventory["focus_potion"] = 1
	updated, _ := model.useConsumable("focus_potion")
	model = updated.(Model)
	before := model.run.combat.enemy.hp
	correct := model.run.question.correct
	updated, _ = model.answer(correct)
	model = updated.(Model)
	if dealt := before - model.run.combat.enemy.hp; dealt != 54 {
		t.Fatalf("damage = %d, want 54", dealt)
	}
	if model.run.inventory["focus_potion"] != 0 {
		t.Fatal("focus potion was not consumed")
	}
}

func TestSharpSwordAddsFiveBaseDamage(t *testing.T) {
	model := combatModel(t, skeletonKind)
	sword, _ := itemByID("sharp_sword")
	model.run.addItem(sword)
	if model.run.weaponBonus != 5 || model.run.equippedWeapon != "锋利之剑" {
		t.Fatalf("weapon bonus = %d, weapon = %q", model.run.weaponBonus, model.run.equippedWeapon)
	}
	before := model.run.combat.enemy.hp
	correct := model.run.question.correct
	updated, _ := model.answer(correct)
	attacked := updated.(Model)
	if dealt := before - attacked.run.combat.enemy.hp; dealt != 32 {
		t.Fatalf("damage = %d, want 32", dealt)
	}
}

func TestMemoryRingAddsDamageForPreviouslyWrongWord(t *testing.T) {
	model := combatModel(t, bossKind)
	model.run.combat.enemy.hp = 300
	model.run.memoryRing = true
	model.run.memoryBonus = 20
	term := model.run.question.word.Term
	model.run.wrongTerms[term] = true
	before := model.run.combat.enemy.hp
	correct := model.run.question.correct
	updated, _ := model.answer(correct)
	attacked := updated.(Model)
	if dealt := before - attacked.run.combat.enemy.hp; dealt != 47 {
		t.Fatalf("damage = %d, want 47", dealt)
	}
	if attacked.run.wrongTerms[term] {
		t.Fatal("remembered word was not cleared")
	}
}

func TestBombDealsDirectDamage(t *testing.T) {
	model := combatModel(t, skeletonKind)
	model.run.inventory["bomb"] = 1
	before := model.run.combat.enemy.hp
	updated, _ := model.useConsumable("bomb")
	attacked := updated.(Model)
	if dealt := before - attacked.run.combat.enemy.hp; dealt != 35 {
		t.Fatalf("damage = %d, want 35", dealt)
	}
	if attacked.run.inventory["bomb"] != 0 {
		t.Fatal("bomb was not consumed")
	}
}

func TestKillingMonsterDropsLootAndReturnsToMap(t *testing.T) {
	model := combatModel(t, slimeKind)
	model.run.combat.enemy.hp = 1
	correct := model.run.question.correct
	updated, _ := model.answer(correct)
	finished := updated.(Model)
	if finished.run.inCombat || finished.run.combat != nil {
		t.Fatal("combat did not end")
	}
	if finished.run.kills != 1 || len(finished.run.loot) != 1 {
		t.Fatalf("kills = %d, loot = %d", finished.run.kills, len(finished.run.loot))
	}
	if len(finished.run.monsters) != 0 {
		t.Fatalf("monster was not removed: %d", len(finished.run.monsters))
	}
}

func TestBossDefeatSettlesVictory(t *testing.T) {
	model := combatModel(t, bossKind)
	model.run.combat.enemy.hp = 1
	correct := model.run.question.correct
	updated, _ := model.answer(correct)
	settled := updated.(Model)
	if settled.screen != screenSettlement || settled.run.outcome != outcomeVictory {
		t.Fatalf("screen = %d outcome = %s", settled.screen, settled.run.outcome)
	}
}

func TestPlayerDefeatSettlesDefeat(t *testing.T) {
	model := combatModel(t, skeletonKind)
	model.run.playerHP = 1
	wrong := (model.run.question.correct + 1) % len(model.run.question.options)
	updated, _ := model.answer(wrong)
	settled := updated.(Model)
	if settled.screen != screenSettlement || settled.run.outcome != outcomeDefeat {
		t.Fatalf("screen = %d outcome = %s", settled.screen, settled.run.outcome)
	}
}

func TestTimerExpirySettlesRun(t *testing.T) {
	model := startedModel(t)
	model.run.endsAt = model.now().Add(-time.Second)
	updated, _ := model.updateTick()
	settled := updated.(Model)
	if settled.screen != screenSettlement || settled.run.outcome != outcomeTimeout {
		t.Fatalf("screen = %d outcome = %s", settled.screen, settled.run.outcome)
	}
	if settled.run.remaining != 0 {
		t.Fatalf("remaining = %v", settled.run.remaining)
	}
}

func TestPlayerCanMoveVerticallyAndPickUpItems(t *testing.T) {
	model := startedModel(t)
	model.run.monsters = nil
	next := point{x: model.run.player.x, y: model.run.player.y + 1}
	if !model.run.dungeon.walkable(next) {
		next = point{x: model.run.player.x, y: model.run.player.y - 1}
	}
	if !model.run.dungeon.walkable(next) {
		t.Fatal("no walkable vertical cell from start")
	}
	dx, dy := next.x-model.run.player.x, next.y-model.run.player.y
	model.run.items = []groundItem{{pos: next, item: itemCatalog[0]}}
	updated, _ := model.movePlayer(dx, dy)
	moved := updated.(Model)
	if moved.run.player != next {
		t.Fatalf("player = %+v, want %+v", moved.run.player, next)
	}
	if moved.run.inventory["health_potion"] != 1 || len(moved.run.items) != 0 {
		t.Fatalf("pickup failed: inventory=%v items=%d", moved.run.inventory, len(moved.run.items))
	}
}

func TestWalkingIntoMonsterStartsCombat(t *testing.T) {
	model := startedModel(t)
	next := point{x: model.run.player.x + 1, y: model.run.player.y}
	if !model.run.dungeon.walkable(next) {
		t.Fatal("cell to the right is not walkable")
	}
	model.run.monsters = []monster{newMonster(skeletonKind, next)}
	updated, _ := model.movePlayer(1, 0)
	attacked := updated.(Model)
	if !attacked.run.inCombat || attacked.run.combat == nil {
		t.Fatal("combat did not start")
	}
	if attacked.run.combat.enemy.kind != skeletonKind {
		t.Fatalf("enemy = %d", attacked.run.combat.enemy.kind)
	}
}

func TestMonsterMovingOntoPlayerTriggersCombat(t *testing.T) {
	model := NewModel(1)
	grid := [][]byte{
		{'#', '#', '#', '#', '#'},
		{'#', '.', '.', '#', '#'},
		{'#', '#', '#', '#', '#'},
	}
	model.run = &runState{
		dungeon:  dungeonMap{grid: grid},
		player:   point{x: 1, y: 1},
		monsters: []monster{newMonster(slimeKind, point{x: 2, y: 1})},
	}
	if index := model.moveMonsters(); index != 0 {
		t.Fatalf("collision index = %d, want 0", index)
	}
	if model.run.monsters[0].pos != model.run.player {
		t.Fatalf("monster = %+v, player = %+v", model.run.monsters[0].pos, model.run.player)
	}
}

func TestViewsContainExplorationAndCombatPanels(t *testing.T) {
	model := startedModel(t)
	view := model.View()
	for _, part := range []string{"SECTOR MAP", "SECTOR // EXPLORATION", "MONSTERS", "SYSTEM LOG"} {
		if !strings.Contains(view, part) {
			t.Errorf("exploration view missing %q", part)
		}
	}

	model = combatModel(t, slimeKind)
	term := model.run.question.word.Term
	view = model.View()
	for _, part := range []string{"ENEMY //", "TARGET ACQUIRED", "SYSTEM LOG", term} {
		if !strings.Contains(view, part) {
			t.Errorf("combat view missing %q", part)
		}
	}
	if strings.Contains(view, strings.ToUpper(term)) {
		t.Errorf("target word rendered in uppercase: %q", term)
	}
}
