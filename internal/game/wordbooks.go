package game

// Word is one vocabulary entry used by the game.
type Word struct {
	Term    string
	Meaning string
}

// Wordbook is a selectable vocabulary archive.
type Wordbook struct {
	ID          string
	Name        string
	Description string
	Words       []Word
}

var wordbooks = []Wordbook{
	{
		ID:          "cet4",
		Name:        "CET-4 CORE",
		Description: "大学英语四级核心词汇",
		Words: []Word{
			{Term: "vulnerability", Meaning: "漏洞；弱点"},
			{Term: "allocate", Meaning: "分配；拨给"},
			{Term: "consume", Meaning: "消耗；消费"},
			{Term: "approach", Meaning: "接近；方法"},
			{Term: "assume", Meaning: "假定；承担"},
			{Term: "benefit", Meaning: "益处；使受益"},
			{Term: "complex", Meaning: "复杂的；综合体"},
			{Term: "concentrate", Meaning: "集中；专心"},
			{Term: "consequence", Meaning: "结果；后果"},
			{Term: "contribute", Meaning: "贡献；促成"},
			{Term: "decline", Meaning: "下降；拒绝"},
			{Term: "demonstrate", Meaning: "证明；演示"},
			{Term: "distinguish", Meaning: "区分；辨别"},
			{Term: "efficient", Meaning: "高效的；有效率的"},
			{Term: "emerge", Meaning: "出现；浮现"},
			{Term: "emphasize", Meaning: "强调；着重"},
			{Term: "estimate", Meaning: "估计；估价"},
			{Term: "evaluate", Meaning: "评估；评价"},
			{Term: "expand", Meaning: "扩大；扩展"},
			{Term: "expose", Meaning: "暴露；揭露"},
			{Term: "flexible", Meaning: "灵活的；可弯曲的"},
			{Term: "generate", Meaning: "产生；生成"},
			{Term: "identify", Meaning: "识别；确认"},
			{Term: "implement", Meaning: "实施；执行"},
			{Term: "indicate", Meaning: "表明；指示"},
			{Term: "maintain", Meaning: "维持；保养"},
			{Term: "modify", Meaning: "修改；调整"},
			{Term: "obvious", Meaning: "明显的；显而易见的"},
			{Term: "participate", Meaning: "参与；参加"},
			{Term: "potential", Meaning: "潜在的；潜力"},
			{Term: "previous", Meaning: "先前的；以前的"},
			{Term: "reduce", Meaning: "减少；降低"},
			{Term: "reliable", Meaning: "可靠的；可信赖的"},
			{Term: "require", Meaning: "要求；需要"},
			{Term: "respond", Meaning: "回应；作出反应"},
			{Term: "significant", Meaning: "重要的；显著的"},
			{Term: "strategy", Meaning: "策略；战略"},
			{Term: "sufficient", Meaning: "足够的；充分的"},
			{Term: "transform", Meaning: "转变；改造"},
			{Term: "valid", Meaning: "有效的；合理的"},
		},
	},
	{
		ID:          "cet6",
		Name:        "CET-6 ADVANCED",
		Description: "大学英语六级进阶词汇",
		Words: []Word{
			{Term: "ambiguous", Meaning: "模棱两可的；含糊的"},
			{Term: "coherent", Meaning: "连贯的；一致的"},
			{Term: "comprehensive", Meaning: "全面的；综合的"},
			{Term: "constrain", Meaning: "限制；约束"},
			{Term: "deteriorate", Meaning: "恶化；变坏"},
			{Term: "discrepancy", Meaning: "差异；不一致"},
			{Term: "elaborate", Meaning: "详尽阐述；精心制作的"},
			{Term: "empirical", Meaning: "经验的；实证的"},
			{Term: "facilitate", Meaning: "促进；使便利"},
			{Term: "fluctuate", Meaning: "波动；起伏"},
			{Term: "hierarchy", Meaning: "等级制度；层级"},
			{Term: "imminent", Meaning: "即将发生的；逼近的"},
			{Term: "incentive", Meaning: "激励；诱因"},
			{Term: "inevitable", Meaning: "不可避免的；必然的"},
			{Term: "integrate", Meaning: "整合；使融入"},
			{Term: "intrinsic", Meaning: "内在的；固有的"},
			{Term: "legitimate", Meaning: "合法的；正当的"},
			{Term: "negligible", Meaning: "微不足道的；可忽略的"},
			{Term: "obsolete", Meaning: "过时的；淘汰的"},
			{Term: "paradox", Meaning: "悖论；自相矛盾的事物"},
			{Term: "plausible", Meaning: "看似合理的；可信的"},
			{Term: "preliminary", Meaning: "初步的；预备的"},
			{Term: "prevalence", Meaning: "流行；普遍存在"},
			{Term: "reinforce", Meaning: "加强；巩固"},
			{Term: "reluctant", Meaning: "不情愿的；勉强的"},
			{Term: "resilient", Meaning: "有韧性的；能恢复的"},
			{Term: "scrutiny", Meaning: "审查；仔细检查"},
			{Term: "subsequently", Meaning: "随后；后来"},
			{Term: "susceptible", Meaning: "易受影响的；敏感的"},
			{Term: "undermine", Meaning: "削弱；破坏"},
		},
	},
	{
		ID:          "toefl",
		Name:        "TOEFL ACADEMIC",
		Description: "托福学术场景高频词汇",
		Words: []Word{
			{Term: "abundant", Meaning: "丰富的；充裕的"},
			{Term: "adjacent", Meaning: "邻近的；相邻的"},
			{Term: "advocate", Meaning: "提倡；拥护者"},
			{Term: "alter", Meaning: "改变；修改"},
			{Term: "anomaly", Meaning: "异常；反常现象"},
			{Term: "arbitrary", Meaning: "任意的；武断的"},
			{Term: "attribute", Meaning: "归因于；属性"},
			{Term: "catalyst", Meaning: "催化剂；促进因素"},
			{Term: "cease", Meaning: "停止；终止"},
			{Term: "compelling", Meaning: "令人信服的；引人注目的"},
			{Term: "constitute", Meaning: "构成；组成"},
			{Term: "contemporary", Meaning: "当代的；同时代的"},
			{Term: "contradict", Meaning: "反驳；与……矛盾"},
			{Term: "cumulative", Meaning: "累积的；渐增的"},
			{Term: "derive", Meaning: "源自；获得"},
			{Term: "diminish", Meaning: "减少；削弱"},
			{Term: "diverse", Meaning: "多样的；不同的"},
			{Term: "dominant", Meaning: "占主导地位的；显性的"},
			{Term: "ecosystem", Meaning: "生态系统"},
			{Term: "enhance", Meaning: "增强；提高"},
			{Term: "explicit", Meaning: "明确的；直白的"},
			{Term: "inherent", Meaning: "固有的；内在的"},
			{Term: "inhibit", Meaning: "抑制；阻止"},
			{Term: "migrate", Meaning: "迁徙；迁移"},
			{Term: "mitigate", Meaning: "减轻；缓和"},
			{Term: "phenomenon", Meaning: "现象；非凡的人或事"},
			{Term: "precipitate", Meaning: "促成；使沉淀"},
			{Term: "profound", Meaning: "深刻的；深远的"},
			{Term: "sustain", Meaning: "维持；支撑"},
			{Term: "synthesize", Meaning: "综合；合成"},
		},
	},
}

func availableWordbooks() []Wordbook {
	books := make([]Wordbook, len(wordbooks))
	copy(books, wordbooks)
	return books
}
