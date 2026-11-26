package game

type monsterKind int

const (
	slimeKind monsterKind = iota
	skeletonKind
	gargoyleKind
	bossKind
)

type monsterSpec struct {
	name    string
	chinese string
	glyph   string
	hp      int
	counter int
	guard   int
}

var monsterSpecs = map[monsterKind]monsterSpec{
	slimeKind: {
		name:    "SLIME",
		chinese: "史莱姆",
		glyph:   "s",
		hp:      55,
		counter: 0,
		guard:   0,
	},
	skeletonKind: {
		name:    "SKELETON",
		chinese: "骷髅兵",
		glyph:   "k",
		hp:      95,
		counter: 10,
		guard:   0,
	},
	gargoyleKind: {
		name:    "GARGOYLE",
		chinese: "石像鬼",
		glyph:   "g",
		hp:      150,
		counter: 12,
		guard:   2,
	},
	bossKind: {
		name:    "THE OBLIVION PARSER",
		chinese: "词渊领主",
		glyph:   "B",
		hp:      360,
		counter: 14,
		guard:   0,
	},
}

type monster struct {
	kind  monsterKind
	pos   point
	hp    int
	maxHP int
	guard int
	phase int
}

func newMonster(kind monsterKind, pos point) monster {
	spec := monsterSpecs[kind]
	return monster{
		kind:  kind,
		pos:   pos,
		hp:    spec.hp,
		maxHP: spec.hp,
		guard: spec.guard,
		phase: 1,
	}
}

func (m monster) spec() monsterSpec {
	return monsterSpecs[m.kind]
}

func (m monster) label() string {
	spec := m.spec()
	return spec.name + " // " + spec.chinese
}

func (m monster) counterDamage() int {
	damage := m.spec().counter
	if m.kind == bossKind && m.phase > 1 {
		damage += (m.phase - 1) * 5
	}
	return damage
}

func (m monster) phaseForHP() int {
	if m.kind != bossKind || m.maxHP <= 0 {
		return 1
	}
	ratio := float64(m.hp) / float64(m.maxHP)
	switch {
	case ratio <= 1.0/3.0:
		return 3
	case ratio <= 2.0/3.0:
		return 2
	default:
		return 1
	}
}
