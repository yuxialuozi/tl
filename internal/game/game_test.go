package game

import (
	"math/rand"
	"strings"
	"testing"
	"time"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
)

func TestWordbooksHaveUniqueChoices(t *testing.T) {
	books := availableWordbooks()
	if len(books) < 6 {
		t.Fatalf("wordbooks = %d, want at least 6", len(books))
	}
	for _, book := range books {
		if len(book.Words) < 20 {
			t.Fatalf("%s has only %d words", book.Name, len(book.Words))
		}
		seen := make(map[string]bool)
		for _, word := range book.Words {
			if word.Term == "" || word.Meaning == "" || word.Phonetic == "" || word.PartOfSpeech == "" || word.Category == "" || word.Definition == "" || len(word.Examples) == 0 {
				t.Fatalf("%s contains an incomplete word entry: %+v", book.Name, word)
			}
			if word.Term != strings.ToLower(word.Term) {
				t.Fatalf("%s word %q is not lowercase", book.Name, word.Term)
			}
			if !strings.HasPrefix(word.Phonetic, "/") || !strings.HasSuffix(word.Phonetic, "/") {
				t.Fatalf("%s word %q has invalid phonetic %q", book.Name, word.Term, word.Phonetic)
			}
			for _, example := range word.Examples {
				if strings.TrimSpace(example) == "" {
					t.Fatalf("%s word %q has an empty example", book.Name, word.Term)
				}
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

func confirmCombatResult(t *testing.T, model Model) Model {
	t.Helper()
	updated, _ := model.updatePlayingKey("enter")
	result, ok := updated.(Model)
	if !ok {
		t.Fatal("confirm result returned a different model type")
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
	if len(model.run.monsters) != 5 {
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
	for _, item := range model.run.items {
		if manhattan(item.pos, model.run.player) < 8 {
			t.Fatalf("starting item too close to player: %+v", item.pos)
		}
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
		model = confirmCombatResult(t, model)
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
		model = confirmCombatResult(t, model)
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
	locked := updated.(Model)
	if !locked.run.combatLocked || !locked.run.pendingKill {
		t.Fatal("kill result should stay locked on screen")
	}
	if locked.run.kills != 1 || len(locked.run.loot) != 1 || locked.run.lastLoot == "" {
		t.Fatalf("kills = %d, loot = %d, last = %q", locked.run.kills, len(locked.run.loot), locked.run.lastLoot)
	}
	finished := confirmCombatResult(t, locked)
	if finished.run.inCombat || finished.run.combat != nil {
		t.Fatal("combat did not end after confirmation")
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
	locked := updated.(Model)
	if !locked.run.combatLocked || !locked.run.pendingKill {
		t.Fatal("boss defeat should remain visible before confirmation")
	}
	settled := confirmCombatResult(t, locked)
	if settled.screen != screenSettlement || settled.run.outcome != outcomeVictory {
		t.Fatalf("screen = %d outcome = %s", settled.screen, settled.run.outcome)
	}
}

func TestPlayerDefeatSettlesDefeat(t *testing.T) {
	model := combatModel(t, skeletonKind)
	model.run.playerHP = 1
	wrong := (model.run.question.correct + 1) % len(model.run.question.options)
	updated, _ := model.answer(wrong)
	locked := updated.(Model)
	if !locked.run.combatLocked || !locked.run.pendingDefeat {
		t.Fatal("defeat should remain visible before confirmation")
	}
	settled := confirmCombatResult(t, locked)
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
	model.run.playerHP = 42
	term := model.run.question.word.Term
	view = model.View()
	for _, part := range []string{"FIRST PERSON // ENCOUNTER", "ENEMY HP", "PLAYER HP", "42/100", "TARGET LOCK", "SYSTEM LOG", term, model.run.question.word.Phonetic} {
		if !strings.Contains(view, part) {
			t.Errorf("combat view missing %q", part)
		}
	}
	correct := model.run.question.correct
	updated, _ := model.answer(correct)
	model = updated.(Model)
	view = model.View()
	for _, part := range []string{"Definition", "中文", "Examples", model.run.question.word.Definition, model.run.question.word.Examples[0], "ENTER 继续"} {
		if !strings.Contains(view, part) {
			t.Errorf("combat result view missing %q", part)
		}
	}
	if strings.Contains(view, "SECTOR MAP") {
		t.Error("combat view should replace the map with the encounter scene")
	}
	if strings.Contains(view, strings.ToUpper(term)) {
		t.Errorf("target word rendered in uppercase: %q", term)
	}
}

func TestBookSelectionOpensModeScreen(t *testing.T) {
	model := NewModel(7)
	updated, _ := model.updateKey(tea.KeyMsg{Type: tea.KeyEnter})
	selected := updated.(Model)
	if selected.screen != screenMode {
		t.Fatalf("screen = %d, want mode selection", selected.screen)
	}
	if len(gameModes) < 3 {
		t.Fatalf("modes = %d", len(gameModes))
	}
}

func TestStudyCoversEveryWordAndCollectsMistakes(t *testing.T) {
	model := NewModel(11)
	model.selected = 0
	model = model.startStudy(studyFull)
	if model.screen != screenStudy || model.study == nil {
		t.Fatal("study mode did not start")
	}
	total := len(model.study.words)
	if total != len(model.books[model.selected].Words) {
		t.Fatalf("study words = %d, want %d", total, len(model.books[model.selected].Words))
	}

	for index := 0; index < total; index++ {
		if model.study.stage != studyPrompt {
			t.Fatalf("word %d stage = %d, want prompt", index, model.study.stage)
		}
		updated, _ := model.updateStudyKey("enter")
		model = updated.(Model)
		if model.study.stage != studyReveal {
			t.Fatalf("word %d did not reveal", index)
		}
		answer := "2"
		if index < 2 {
			answer = "1"
		}
		updated, _ = model.updateStudyKey(answer)
		model = updated.(Model)
	}

	if model.study.stage != studyDone {
		t.Fatalf("stage = %d, want done", model.study.stage)
	}
	if model.study.known+model.study.unknown != total {
		t.Fatalf("known + unknown = %d, want %d", model.study.known+model.study.unknown, total)
	}
	if model.reviewCount(model.books[model.selected]) != 2 {
		t.Fatalf("review count = %d, want 2", model.reviewCount(model.books[model.selected]))
	}

	review := model.startStudy(studyReview)
	if len(review.study.words) != 2 {
		t.Fatalf("review words = %d, want 2", len(review.study.words))
	}
}

func TestStudyRevealShowsPhoneticDefinitionAndExample(t *testing.T) {
	model := NewModel(17)
	model = model.startStudy(studyFull)
	updated, _ := model.updateStudyKey("enter")
	model = updated.(Model)
	word := model.study.words[model.study.index]
	view := model.View()
	for _, part := range []string{word.Phonetic, "Definition", word.Definition, "中文", word.Meaning, "Examples", word.Examples[0]} {
		if !strings.Contains(view, part) {
			t.Errorf("study card missing %q", part)
		}
	}
}

func TestBagSelectionUsesChosenConsumable(t *testing.T) {
	model := combatModel(t, skeletonKind)
	model.run.inventory["health_potion"] = 1
	model.run.inventory["bomb"] = 1
	updated, _ := model.updatePlayingKey("e")
	model = updated.(Model)
	if !model.run.bagOpen || len(model.run.bagEntries()) != 2 {
		t.Fatalf("bag open = %v, entries = %d", model.run.bagOpen, len(model.run.bagEntries()))
	}
	model.run.bagSelection = 1
	before := model.run.combat.enemy.hp
	updated, _ = model.updateBagKey("enter")
	model = updated.(Model)
	if model.run.combat.enemy.hp >= before {
		t.Fatal("selected bomb did not damage the enemy")
	}
	if model.run.inventory["bomb"] != 0 || model.run.inventory["health_potion"] != 1 {
		t.Fatalf("wrong item consumed: %v", model.run.inventory)
	}
	if model.run.bagOpen {
		t.Fatal("bag should close after using an item")
	}
}

func TestMonsterLootPoolsMatchDifficulty(t *testing.T) {
	allowed := func(kind monsterKind) map[string]itemRarity {
		result := map[string]itemRarity{}
		for rarity, ids := range monsterLoot[kind] {
			for _, id := range ids {
				result[id] = rarity
			}
		}
		return result
	}
	rng := rand.New(rand.NewSource(99))
	for _, kind := range []monsterKind{slimeKind, skeletonKind, gargoyleKind, bossKind} {
		pool := allowed(kind)
		for draw := 0; draw < 500; draw++ {
			item := randomItem(rng, kind)
			if _, ok := pool[item.id]; !ok {
				t.Fatalf("%v dropped %s outside its pool", kind, item.id)
			}
			if kind == slimeKind && item.rarity > rarityRare {
				t.Fatalf("slime dropped epic %s", item.id)
			}
			if (kind == gargoyleKind || kind == bossKind) && item.rarity < rarityRare {
				t.Fatalf("%v dropped common %s", kind, item.id)
			}
		}
	}
}

func TestDamageFloatersStayUntilConfirmationAndCapAt999(t *testing.T) {
	small := renderImpact(damageFloater{value: 19, label: "HIT"})
	large := renderImpact(damageFloater{value: 80, label: "CRIT", kind: floaterCrit})
	capped := renderImpact(damageFloater{value: 5000, label: "CRIT", kind: floaterCrit})
	if len(small) != 1 {
		t.Fatalf("small impact lines = %d, want 1", len(small))
	}
	if len(large) <= len(small) {
		t.Fatalf("large impact lines = %d, want more than %d", len(large), len(small))
	}
	for _, line := range capped {
		if lipgloss.Width(line) > mapWidth {
			t.Fatalf("capped impact width = %d, want <= %d", lipgloss.Width(line), mapWidth)
		}
	}

	model := combatModel(t, slimeKind)
	model.run.combat.enemy.hp = 5000
	model.run.combat.enemy.maxHP = 5000
	model.run.baseDamage = 5000
	correct := model.run.question.correct
	updated, _ := model.answer(correct)
	model = updated.(Model)
	if len(model.run.floaters) == 0 || model.run.hitFlash == 0 {
		t.Fatal("correct answer did not trigger hit feedback")
	}
	if !model.run.combatLocked {
		t.Fatal("answer result was not locked")
	}
	floater, ok := model.run.latestFloater()
	if !ok || floater.value != maxCombatDamage {
		t.Fatalf("float value = %d, want %d", floater.value, maxCombatDamage)
	}
	if model.run.correct != 1 {
		t.Fatalf("correct = %d", model.run.correct)
	}
	updated, _ = model.answer((correct + 1) % len(model.run.question.options))
	model = updated.(Model)
	if model.run.correct != 1 || model.run.wrong != 0 {
		t.Fatal("answer input should be ignored while the result is locked")
	}
	for i := 0; i < 80; i++ {
		updated, _ = model.updateFrame()
		model = updated.(Model)
	}
	if len(model.run.floaters) == 0 || !model.run.combatLocked {
		t.Fatal("impact disappeared before confirmation")
	}

	model = confirmCombatResult(t, model)
	if model.run.combatLocked || len(model.run.floaters) != 0 {
		t.Fatal("impact did not clear after confirmation")
	}
}

func TestPauseAndResultHoldFreezeClock(t *testing.T) {
	now := time.Date(2026, 11, 26, 12, 0, 0, 0, time.Local)
	model := NewModel(21)
	model.now = func() time.Time { return now }
	started, _ := model.startRun()
	model = started.(Model)

	model.run.hitFlash = 3
	model.run.monsterTicks = 2
	remaining := model.run.remaining
	endsAt := model.run.endsAt
	updated, _ := model.updatePlayingKey("p")
	model = updated.(Model)
	if !model.run.paused {
		t.Fatal("pause did not activate")
	}

	now = now.Add(20 * time.Second)
	updated, _ = model.updateTick()
	model = updated.(Model)
	updated, _ = model.updateFrame()
	model = updated.(Model)
	if model.run.remaining != remaining || model.run.monsterTicks != 2 || model.run.hitFlash != 3 {
		t.Fatal("pause did not freeze the clock, monsters, and animation")
	}
	if !model.run.endsAt.Equal(endsAt) {
		t.Fatal("paused clock changed endsAt before resuming")
	}

	updated, _ = model.updatePlayingKey("p")
	model = updated.(Model)
	if model.run.paused || !model.run.endsAt.Equal(endsAt.Add(20*time.Second)) {
		t.Fatalf("resume clock = %v, want %v", model.run.endsAt, endsAt.Add(20*time.Second))
	}

	model = combatModel(t, slimeKind)
	clockNow := time.Date(2026, 11, 26, 13, 0, 0, 0, time.Local)
	model.now = func() time.Time { return clockNow }
	model.run.endsAt = clockNow.Add(RunDuration)
	model.run.remaining = RunDuration
	correct := model.run.question.correct
	updated, _ = model.answer(correct)
	model = updated.(Model)
	heldRemaining := model.run.remaining
	clockNow = clockNow.Add(15 * time.Second)
	updated, _ = model.updateTick()
	model = updated.(Model)
	if model.run.remaining != heldRemaining {
		t.Fatal("result screen did not hold the run clock")
	}
	model = confirmCombatResult(t, model)
	updated, _ = model.updateTick()
	model = updated.(Model)
	if model.run.remaining != heldRemaining {
		t.Fatalf("remaining after result = %v, want %v", model.run.remaining, heldRemaining)
	}
}

func TestSkillCatalogOffersThreeDistinctChoices(t *testing.T) {
	if len(skillCatalog) < 12 {
		t.Fatalf("skills = %d, want at least 12", len(skillCatalog))
	}
	model := combatModel(t, slimeKind)
	model.offerSkills()
	seen := map[string]bool{}
	for _, skill := range model.run.skillChoices {
		if seen[skill.Name] {
			t.Fatalf("duplicate skill choice %q", skill.Name)
		}
		seen[skill.Name] = true
	}
}

func TestViewsFitMinimumTerminal(t *testing.T) {
	assertFits := func(t *testing.T, name string, model Model) {
		t.Helper()
		view := model.View()
		if width := lipgloss.Width(view); width > 80 {
			t.Errorf("%s width = %d, want <= 80", name, width)
		}
		lines := strings.Split(view, "\n")
		if len(lines) > 24 {
			t.Errorf("%s height = %d, want <= 24", name, len(lines))
		}
		for index, line := range lines {
			if width := lipgloss.Width(line); width > 80 {
				t.Errorf("%s line %d width = %d, want <= 80", name, index+1, width)
			}
		}
	}

	selectModel := NewModel(13)
	selectModel.width = 80
	selectModel.height = 24
	assertFits(t, "select", selectModel)

	selectModel.screen = screenMode
	assertFits(t, "mode", selectModel)

	study := selectModel.startStudy(studyFull)
	assertFits(t, "study prompt", study)
	study.study.stage = studyReveal
	assertFits(t, "study reveal", study)
	study.study.stage = studyDone
	assertFits(t, "study result", study)

	exploration := startedModel(t)
	exploration.width = 80
	exploration.height = 24
	assertFits(t, "exploration", exploration)

	combat := combatModel(t, bossKind)
	combat.width = 80
	combat.height = 24
	combat.run.addFloater(88, "CRIT", floaterCrit)
	assertFits(t, "combat", combat)

	combat.run.inventory["bomb"] = 1
	combat.run.bagOpen = true
	assertFits(t, "bag", combat)

	combat.run.bagOpen = false
	combat.run.combatLocked = true
	combat.run.pendingKill = true
	combat.run.lastResult = "TARGET DESTROYED"
	combat.run.lastCorrect = true
	combat.run.lastLoot = "锋利之剑 [RARE] · 已装备"
	combat.run.floaters = nil
	combat.run.addFloater(maxCombatDamage, "CRIT", floaterCrit)
	assertFits(t, "combat result", combat)

	combat.run.paused = true
	assertFits(t, "paused", combat)
	combat.run.paused = false

	combat.screen = screenSettlement
	combat.run.outcome = outcomeVictory
	assertFits(t, "settlement", combat)
}
