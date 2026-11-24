const DEFAULT_SETTINGS = globalThis.TL_DEFAULT_SETTINGS;

const TRANSLATION_SYSTEM_PROMPT = `You are a precise translator.
Return only one valid JSON object with this exact shape:
{"translation":"translated text"}

Rules:
- Translate according to the requested source and target languages.
- Preserve meaning, tone, names, numbers, and paragraph breaks.
- Do not explain grammar, define words, or add advice.
- Do not use Markdown, commentary, or extra keys.`;

const VOCABULARY_SYSTEM_PROMPT = `You are a concise English learner's dictionary.
Return only one valid JSON object with this exact shape:
{"word":"keyring","part_of_speech":"noun","category":"Computing","definition":"A system that securely stores passwords and other secret information.","chinese":"密钥环，用于安全存储密码、密钥等敏感信息。","examples":["The application stores the password in the keyring.","Linux uses a keyring to manage credentials."]}

Rules:
- definition must be simple, natural English that an intermediate learner can understand.
- chinese must be a concise Simplified Chinese meaning, not an explanation.
- examples must contain exactly two natural English sentences.
- category is an optional short subject label such as Computing or Biology.
- Do not use Markdown, commentary, or extra keys.`;

const ENGLISH_WORD_PATTERN = /^[A-Za-z]+(?:['’-][A-Za-z]+)*$/;

function isSingleEnglishWord(text) {
  return ENGLISH_WORD_PATTERN.test(text.trim());
}

function detectLanguage(text) {
  if (/[\u3400-\u9fff\uf900-\ufaff]/.test(text)) {
    return "Chinese";
  }
  if (/[A-Za-z]/.test(text)) {
    return "English";
  }
  return "";
}

function buildTranslationPrompt(text, source, target) {
  return `Source language: ${source}\nTarget language: ${target}\n\nText:\n${text}`;
}

function buildVocabularyPrompt(word) {
  return `Explain this English word: ${word}`;
}

function normalizeSettings(stored) {
  const timeout = Number(stored.timeout);
  return {
    baseUrl: String(stored.baseUrl || DEFAULT_SETTINGS.baseUrl).trim().replace(/\/+$/, ""),
    apiKey: String(stored.apiKey || "").trim(),
    model: String(stored.model || DEFAULT_SETTINGS.model).trim(),
    timeout: Number.isFinite(timeout) && timeout > 0 ? timeout : DEFAULT_SETTINGS.timeout,
    enableThinking: Boolean(stored.enableThinking),
  };
}

async function getSettings() {
  const stored = await browser.storage.local.get(DEFAULT_SETTINGS);
  return normalizeSettings(stored);
}

function extractJSONObject(content) {
  let trimmed = String(content || "").trim();
  if (trimmed.startsWith("```")) {
    const newline = trimmed.indexOf("\n");
    if (newline >= 0) {
      trimmed = trimmed.slice(newline + 1);
    }
    trimmed = trimmed.replace(/```\s*$/, "").trim();
  }

  const start = trimmed.indexOf("{");
  const end = trimmed.lastIndexOf("}");
  if (start < 0 || end < start) {
    throw new Error("模型返回内容不是有效的 JSON");
  }
  return trimmed.slice(start, end + 1);
}

function redact(text, apiKey) {
  const value = String(text || "");
  return apiKey ? value.split(apiKey).join("<redacted>") : value;
}

function errorMessage(status, body) {
  let message = "";
  try {
    const decoded = JSON.parse(body);
    message = decoded?.error?.message || decoded?.msg || "";
  } catch {
    // Fall back to the raw response below.
  }
  if (!message) {
    message = String(body || "").trim().slice(0, 300);
  }
  return message || `HTTP ${status}`;
}

async function requestCompletion(settings, systemPrompt, userPrompt) {
  const controller = new AbortController();
  const timer = setTimeout(() => controller.abort(), settings.timeout * 1000);
  let response;
  let body;

  try {
    response = await fetch(`${settings.baseUrl}/chat/completions`, {
      method: "POST",
      headers: {
        "Authorization": `Bearer ${settings.apiKey}`,
        "Content-Type": "application/json",
        "Accept": "application/json",
      },
      body: JSON.stringify({
        model: settings.model,
        messages: [
          { role: "system", content: systemPrompt },
          { role: "user", content: userPrompt },
        ],
        enable_thinking: settings.enableThinking,
      }),
      signal: controller.signal,
    });
    body = await response.text();
  } catch (error) {
    if (error?.name === "AbortError" || controller.signal.aborted) {
      return { ok: false, error: `请求超时（${settings.timeout} 秒），请稍后重试。` };
    }
    return { ok: false, error: redact(`请求失败：${error?.message || error}`, settings.apiKey) };
  } finally {
    clearTimeout(timer);
  }

  if (!response.ok) {
    return {
      ok: false,
      error: redact(`翻译服务返回错误（HTTP ${response.status}）：${errorMessage(response.status, body)}`, settings.apiKey),
    };
  }

  let decoded;
  try {
    decoded = JSON.parse(body);
  } catch {
    return { ok: false, error: "翻译服务返回了无法解析的响应。" };
  }

  const content = decoded?.choices?.[0]?.message?.content;
  if (!content) {
    return { ok: false, error: "翻译服务没有返回内容。" };
  }
  return { ok: true, content };
}

function parseVocabulary(content, input) {
  const decoded = JSON.parse(extractJSONObject(content));
  const word = String(decoded.word || input).trim();
  const partOfSpeech = String(decoded.part_of_speech || "").trim();
  const category = String(decoded.category || "").trim();
  const definition = String(decoded.definition || "").trim();
  const chinese = String(decoded.chinese || "").trim();
  const examples = Array.isArray(decoded.examples)
    ? decoded.examples.map((example) => String(example || "").trim()).filter(Boolean).slice(0, 2)
    : [];

  if (!partOfSpeech || !definition || !chinese) {
    throw new Error("模型返回的词典内容不完整");
  }
  if (examples.length < 2) {
    throw new Error("模型返回的例句不足两个");
  }

  return { word, partOfSpeech, category, definition, chinese, examples };
}

function parseTranslation(content) {
  const translation = String(JSON.parse(extractJSONObject(content))?.translation || "").trim();
  if (!translation) {
    throw new Error("模型返回的译文为空");
  }
  return translation;
}

async function processText(text) {
  const input = String(text || "").trim();
  if (!input) {
    return { ok: false, error: "选中的内容为空。" };
  }

  const settings = await getSettings();
  if (!settings.apiKey) {
    return { ok: false, error: "尚未配置 API Key，请先打开扩展设置。", needsSettings: true };
  }
  if (!settings.model) {
    return { ok: false, error: "尚未配置模型，请先打开扩展设置。", needsSettings: true };
  }

  if (isSingleEnglishWord(input)) {
    const completion = await requestCompletion(settings, VOCABULARY_SYSTEM_PROMPT, buildVocabularyPrompt(input));
    if (!completion.ok) {
      return completion;
    }
    try {
      return { ok: true, mode: "vocabulary", vocabulary: parseVocabulary(completion.content, input) };
    } catch (error) {
      return { ok: false, error: redact(error?.message || "模型返回格式无法解析。", settings.apiKey) };
    }
  }

  const source = detectLanguage(input);
  if (!source) {
    return { ok: false, error: "选中的内容不包含英文或中文。" };
  }
  const target = source === "Chinese" ? "English" : "Chinese";
  const completion = await requestCompletion(
    settings,
    TRANSLATION_SYSTEM_PROMPT,
    buildTranslationPrompt(input, source, target),
  );
  if (!completion.ok) {
    return completion;
  }

  try {
    return {
      ok: true,
      mode: "translation",
      original: input,
      translation: parseTranslation(completion.content),
      sourceLanguage: source,
      targetLanguage: target,
    };
  } catch (error) {
    return { ok: false, error: redact(error?.message || "模型返回格式无法解析。", settings.apiKey) };
  }
}

browser.runtime.onMessage.addListener((message) => {
  if (message?.type === "tl-translate") {
    return processText(message.text);
  }
  if (message?.type === "tl-open-options") {
    return browser.runtime.openOptionsPage();
  }
  return undefined;
});

browser.runtime.onInstalled.addListener((details) => {
  if (details.reason === "install") {
    browser.runtime.openOptionsPage();
  }
});
