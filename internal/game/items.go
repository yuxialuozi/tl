package game

import (
	"fmt"
	"math/rand"
)

type itemCategory int

const (
	categoryConsumable itemCategory = iota
	categoryWeapon
	categoryAccessory
	categoryPassive
)

type itemRarity int

const (
	rarityCommon itemRarity = iota
	rarityRare
	rarityEpic
)

type itemSpec struct {
	id              string
	name            string
	english         string
	description     string
	category        itemCategory
	rarity          itemRarity
	heal            int
	directDamage    int
	focus           bool
	shield          bool
	weaponDamage    int
	maxHP           int
	lifesteal       int
	comboBonus      int
	critChance      int
	memoryBonus     int
	wrongProtection bool
}

type groundItem struct {
	pos  point
	item itemSpec
}

type lootEntry struct {
	id     string
	weight int
}

var itemCatalog = []itemSpec{
	{
		id:          "health_potion",
		name:        "生命药水",
		english:     "Health Potion",
		description: "恢复 30 点生命值",
		category:    categoryConsumable,
		rarity:      rarityCommon,
		heal:        30,
	},
	{
		id:          "focus_potion",
		name:        "专注药水",
		english:     "Focus Potion",
		description: "下一道答对造成双倍伤害",
		category:    categoryConsumable,
		rarity:      rarityRare,
		focus:       true,
	},
	{
		id:           "bomb",
		name:         "炸弹",
		english:      "Bomb",
		description:  "立即对敌人造成 35 点伤害",
		category:     categoryConsumable,
		rarity:       rarityCommon,
		directDamage: 35,
	},
	{
		id:          "ward_scroll",
		name:        "护佑卷轴",
		english:     "Scroll of Warding",
		description: "抵消下一次答错伤害",
		category:    categoryConsumable,
		rarity:      rarityRare,
		shield:      true,
	},
	{
		id:           "wood_sword",
		name:         "木剑",
		english:      "Wood Sword",
		description:  "基础伤害 +2",
		category:     categoryWeapon,
		rarity:       rarityCommon,
		weaponDamage: 2,
	},
	{
		id:           "iron_sword",
		name:         "铁剑",
		english:      "Iron Sword",
		description:  "基础伤害 +4",
		category:     categoryWeapon,
		rarity:       rarityRare,
		weaponDamage: 4,
	},
	{
		id:           "sharp_sword",
		name:         "锋利之剑",
		english:      "Sharp Sword",
		description:  "所有正确答案的基础伤害 +5",
		category:     categoryWeapon,
		rarity:       rarityRare,
		weaponDamage: 5,
	},
	{
		id:           "flame_sword",
		name:         "火焰剑",
		english:      "Flame Sword",
		description:  "基础伤害 +8",
		category:     categoryWeapon,
		rarity:       rarityEpic,
		weaponDamage: 8,
	},
	{
		id:          "memory_ring",
		name:        "记忆戒指",
		english:     "Ring of Memory",
		description: "答错过的单词再次答对时，额外造成 20 点伤害",
		category:    categoryAccessory,
		rarity:      rarityRare,
		memoryBonus: 20,
	},
	{
		id:          "vitality_amulet",
		name:        "生命护符",
		english:     "Amulet of Vitality",
		description: "最大生命 +25，并恢复 25 HP",
		category:    categoryAccessory,
		rarity:      rarityRare,
		maxHP:       25,
	},
	{
		id:          "momentum_necklace",
		name:        "连击项链",
		english:     "Necklace of Momentum",
		description: "每层连击伤害 +1",
		category:    categoryAccessory,
		rarity:      rarityRare,
		comboBonus:  1,
	},
	{
		id:          "crit_rune",
		name:        "暴击符文",
		english:     "Crit Rune",
		description: "20% 概率造成双倍伤害",
		category:    categoryPassive,
		rarity:      rarityRare,
		critChance:  20,
	},
	{
		id:          "vampire_fang",
		name:        "吸血獠牙",
		english:     "Vampire Fang",
		description: "每次答对恢复 3 HP",
		category:    categoryPassive,
		rarity:      rarityRare,
		lifesteal:   3,
	},
	{
		id:          "combo_core",
		name:        "连击核心",
		english:     "Combo Core",
		description: "连击伤害加成 +2",
		category:    categoryPassive,
		rarity:      rarityRare,
		comboBonus:  2,
	},
	{
		id:              "guard_charm",
		name:            "答错护符",
		english:         "Guard Charm",
		description:     "每场战斗抵消一次答错伤害",
		category:        categoryPassive,
		rarity:          rarityEpic,
		wrongProtection: true,
	},
}

func itemByID(id string) (itemSpec, bool) {
	for _, item := range itemCatalog {
		if item.id == id {
			return item, true
		}
	}
	return itemSpec{}, false
}

var monsterLoot = map[monsterKind]map[itemRarity][]string{
	slimeKind: {
		rarityCommon: {"health_potion", "bomb", "wood_sword"},
		rarityRare:   {"combo_core"},
	},
	skeletonKind: {
		rarityCommon: {"health_potion", "ward_scroll"},
		rarityRare:   {"focus_potion", "iron_sword", "memory_ring", "vampire_fang"},
	},
	gargoyleKind: {
		rarityRare: {"sharp_sword", "crit_rune", "momentum_necklace", "vitality_amulet"},
		rarityEpic: {"flame_sword", "guard_charm"},
	},
	bossKind: {
		rarityRare: {"memory_ring", "sharp_sword"},
		rarityEpic: {"flame_sword", "guard_charm"},
	},
}

var monsterRarityWeights = map[monsterKind][3]int{
	slimeKind:    {90, 10, 0},
	skeletonKind: {60, 40, 0},
	gargoyleKind: {0, 75, 25},
	bossKind:     {0, 25, 75},
}

var startingLootTable = []lootEntry{
	{id: "health_potion", weight: 25},
	{id: "focus_potion", weight: 15},
	{id: "ward_scroll", weight: 15},
	{id: "iron_sword", weight: 20},
	{id: "memory_ring", weight: 15},
	{id: "vitality_amulet", weight: 10},
}

func randomItem(rng *rand.Rand, kind monsterKind) itemSpec {
	weights := monsterRarityWeights[kind]
	type rollGroup struct {
		rarity itemRarity
		weight int
	}
	groups := make([]rollGroup, 0, 3)
	total := 0
	for rarity := rarityCommon; rarity <= rarityEpic; rarity++ {
		if weights[rarity] <= 0 || len(monsterLoot[kind][rarity]) == 0 {
			continue
		}
		groups = append(groups, rollGroup{rarity: rarity, weight: weights[rarity]})
		total += weights[rarity]
	}
	if total > 0 {
		roll := rng.Intn(total)
		rarity := rarityCommon
		for _, group := range groups {
			roll -= group.weight
			if roll < 0 {
				rarity = group.rarity
				break
			}
		}
		pool := monsterLoot[kind][rarity]
		if len(pool) > 0 {
			if item, ok := itemByID(pool[rng.Intn(len(pool))]); ok {
				return item
			}
		}
	}
	return itemCatalog[0]
}

func randomGroundItem(rng *rand.Rand) itemSpec {
	total := 0
	for _, entry := range startingLootTable {
		total += entry.weight
	}
	roll := rng.Intn(total)
	for _, entry := range startingLootTable {
		roll -= entry.weight
		if roll < 0 {
			if item, ok := itemByID(entry.id); ok {
				return item
			}
		}
	}
	return itemCatalog[0]
}

func rarityName(rarity itemRarity) string {
	switch rarity {
	case rarityRare:
		return "RARE"
	case rarityEpic:
		return "EPIC"
	default:
		return "COMMON"
	}
}

func (r *runState) addItem(item itemSpec) string {
	r.inventory[item.id]++
	switch item.category {
	case categoryWeapon:
		if item.weaponDamage > r.equippedWeaponDamage {
			skillBonus := max(0, r.weaponBonus-r.equippedWeaponDamage)
			r.equippedWeaponDamage = item.weaponDamage
			r.weaponBonus = item.weaponDamage + skillBonus
			r.equippedWeapon = item.name
			return fmt.Sprintf("已装备 %s，基础伤害 +%d", item.name, r.weaponBonus)
		}
		return fmt.Sprintf("%s 已存入背包", item.name)
	case categoryAccessory, categoryPassive:
		return r.applyGear(item)
	default:
		return fmt.Sprintf("%s 已存入背包", item.name)
	}
}

func (r *runState) applyGear(item itemSpec) string {
	switch {
	case item.memoryBonus > 0:
		r.memoryRing = true
		r.memoryBonus = item.memoryBonus
		return fmt.Sprintf("已装备 %s，重温错词 +%d", item.name, item.memoryBonus)
	case item.maxHP > 0:
		r.playerMaxHP += item.maxHP
		r.playerHP = min(r.playerMaxHP, r.playerHP+item.maxHP)
		return fmt.Sprintf("已装备 %s，最大生命 +%d", item.name, item.maxHP)
	case item.wrongProtection:
		r.guardCharm = true
		r.shield = true
		return fmt.Sprintf("已装备 %s，每场战斗获得 1 次答错保护", item.name)
	case item.critChance > 0:
		r.critChance += item.critChance
		return fmt.Sprintf("已获得 %s，暴击率 +%d%%", item.name, item.critChance)
	case item.lifesteal > 0:
		r.lifesteal += item.lifesteal
		return fmt.Sprintf("已获得 %s，答对回复 %d HP", item.name, item.lifesteal)
	case item.comboBonus > 0:
		r.comboBonus += item.comboBonus
		return fmt.Sprintf("已获得 %s，连击伤害 +%d", item.name, item.comboBonus)
	default:
		return fmt.Sprintf("%s 已存入背包", item.name)
	}
}
