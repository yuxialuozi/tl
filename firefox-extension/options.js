const DEFAULT_SETTINGS = globalThis.TL_DEFAULT_SETTINGS;
const form = document.getElementById("settings-form");
const baseUrlInput = document.getElementById("base-url");
const apiKeyInput = document.getElementById("api-key");
const modelInput = document.getElementById("model");
const timeoutInput = document.getElementById("timeout");
const enableThinkingInput = document.getElementById("enable-thinking");
const status = document.getElementById("status");

function setStatus(message, kind = "") {
  status.textContent = message;
  status.className = kind;
}

function normalizeBaseUrl(value) {
  const parsed = new URL(value);
  if (!["http:", "https:"].includes(parsed.protocol)) {
    throw new Error("Base URL 必须以 http:// 或 https:// 开头");
  }
  return parsed.toString().replace(/\/+$/, "");
}

async function ensureHostPermission(baseUrl) {
  const parsed = new URL(baseUrl);
  const originPattern = `${parsed.protocol}//${parsed.host}/*`;
  const permissions = { origins: [originPattern] };
  if (await browser.permissions.contains(permissions)) {
    return;
  }
  const granted = await browser.permissions.request(permissions);
  if (!granted) {
    throw new Error(`未授权访问 ${parsed.host}`);
  }
}

async function loadSettings() {
  const stored = await browser.storage.local.get(DEFAULT_SETTINGS);
  baseUrlInput.value = stored.baseUrl || DEFAULT_SETTINGS.baseUrl;
  apiKeyInput.value = stored.apiKey || "";
  modelInput.value = stored.model || DEFAULT_SETTINGS.model;
  timeoutInput.value = stored.timeout || DEFAULT_SETTINGS.timeout;
  enableThinkingInput.checked = Boolean(stored.enableThinking);
}

form.addEventListener("submit", async (event) => {
  event.preventDefault();
  setStatus("保存中…");

  try {
    const baseUrl = normalizeBaseUrl(baseUrlInput.value.trim());
    const apiKey = apiKeyInput.value.trim();
    const model = modelInput.value.trim();
    const timeout = Number(timeoutInput.value);

    if (!apiKey) {
      throw new Error("请填写 API Key");
    }
    if (!model) {
      throw new Error("请填写模型名称");
    }
    if (!Number.isFinite(timeout) || timeout <= 0) {
      throw new Error("Timeout 必须是大于 0 的数字");
    }

    await ensureHostPermission(baseUrl);
    await browser.storage.local.set({
      baseUrl,
      apiKey,
      model,
      timeout,
      enableThinking: enableThinkingInput.checked,
    });
    setStatus("已保存", "success");
  } catch (error) {
    setStatus(error?.message || String(error), "error");
  }
});

loadSettings().catch((error) => {
  setStatus(`读取设置失败：${error?.message || error}`, "error");
});
