(() => {
  if (document.getElementById("tl-selection-root")) {
    return;
  }

  const host = document.createElement("div");
  host.id = "tl-selection-root";
  host.style.position = "fixed";
  host.style.inset = "0";
  host.style.width = "0";
  host.style.height = "0";
  host.style.zIndex = "2147483647";
  host.style.pointerEvents = "none";

  const shadow = host.attachShadow({ mode: "closed" });
  const style = document.createElement("style");
  style.textContent = `
    :host {
      all: initial;
    }
    * {
      box-sizing: border-box;
    }
    .tl-icon,
    .tl-panel {
      position: fixed;
      z-index: 2147483647;
      font-family: -apple-system, BlinkMacSystemFont, "Segoe UI", sans-serif;
      pointer-events: auto;
    }
    .tl-icon {
      display: none;
      width: 32px;
      height: 32px;
      padding: 0;
      border: 1px solid rgba(255, 255, 255, 0.35);
      border-radius: 50%;
      background: #1976d2;
      color: #fff;
      font-size: 15px;
      font-weight: 700;
      line-height: 30px;
      text-align: center;
      cursor: pointer;
      box-shadow: 0 4px 14px rgba(0, 0, 0, 0.22);
      user-select: none;
    }
    .tl-icon:hover {
      background: #1565c0;
      transform: translateY(-1px);
    }
    .tl-panel {
      display: none;
      width: min(380px, calc(100vw - 16px));
      max-height: min(440px, calc(100vh - 16px));
      overflow: hidden;
      border: 1px solid rgba(0, 0, 0, 0.12);
      border-radius: 12px;
      background: #fff;
      color: #1f2328;
      box-shadow: 0 12px 36px rgba(0, 0, 0, 0.22);
    }
    .tl-header {
      display: flex;
      align-items: center;
      justify-content: space-between;
      min-height: 38px;
      padding: 0 8px 0 14px;
      border-bottom: 1px solid #eceff1;
      color: #57606a;
      font-size: 12px;
      font-weight: 600;
      letter-spacing: 0.02em;
    }
    .tl-close {
      width: 28px;
      height: 28px;
      padding: 0;
      border: 0;
      border-radius: 6px;
      background: transparent;
      color: #57606a;
      font-size: 18px;
      line-height: 28px;
      cursor: pointer;
    }
    .tl-close:hover {
      background: #f3f4f6;
    }
    .tl-body {
      max-height: min(390px, calc(100vh - 70px));
      overflow: auto;
      padding: 14px 16px 16px;
      color: #1f2328;
      font-size: 14px;
      line-height: 1.65;
      overflow-wrap: anywhere;
    }
    .tl-body.tl-error {
      color: #b42318;
    }
    .tl-word-header {
      margin-bottom: 16px;
      color: #1f2328;
      font-size: 16px;
      font-weight: 700;
    }
    .tl-section + .tl-section {
      margin-top: 15px;
    }
    .tl-heading {
      margin-bottom: 4px;
      color: #1976d2;
      font-size: 12px;
      font-weight: 700;
      letter-spacing: 0.03em;
    }
    .tl-text {
      white-space: pre-wrap;
    }
    .tl-examples {
      display: grid;
      gap: 4px;
    }
    .tl-example::before {
      content: "- ";
    }
    .tl-loading {
      display: flex;
      align-items: center;
      gap: 9px;
      color: #57606a;
    }
    .tl-spinner {
      width: 16px;
      height: 16px;
      flex: 0 0 auto;
      border: 2px solid #d0d7de;
      border-top-color: #1976d2;
      border-radius: 50%;
      animation: tl-spin 0.8s linear infinite;
    }
    .tl-open-options {
      margin-top: 10px;
      padding: 5px 10px;
      border: 1px solid #b42318;
      border-radius: 6px;
      background: transparent;
      color: #b42318;
      font: inherit;
      font-size: 12px;
      cursor: pointer;
    }
    @keyframes tl-spin {
      to { transform: rotate(360deg); }
    }
    @media (prefers-color-scheme: dark) {
      .tl-panel {
        border-color: #30363d;
        background: #161b22;
        color: #e6edf3;
      }
      .tl-header {
        border-bottom-color: #30363d;
        color: #8b949e;
      }
      .tl-close {
        color: #8b949e;
      }
      .tl-close:hover {
        background: #21262d;
      }
      .tl-body {
        color: #e6edf3;
      }
      .tl-body.tl-error {
        color: #ff7b72;
      }
      .tl-word-header {
        color: #e6edf3;
      }
      .tl-heading {
        color: #58a6ff;
      }
      .tl-loading {
        color: #8b949e;
      }
      .tl-spinner {
        border-color: #30363d;
        border-top-color: #58a6ff;
      }
      .tl-open-options {
        border-color: #ff7b72;
        color: #ff7b72;
      }
    }
  `;

  const icon = document.createElement("button");
  icon.className = "tl-icon";
  icon.type = "button";
  icon.textContent = "译";
  icon.title = "翻译选中内容";
  icon.setAttribute("aria-label", "翻译选中内容");

  const panel = document.createElement("section");
  panel.className = "tl-panel";
  panel.setAttribute("role", "dialog");
  panel.setAttribute("aria-live", "polite");

  const header = document.createElement("div");
  header.className = "tl-header";
  const title = document.createElement("span");
  title.textContent = "tl";
  const close = document.createElement("button");
  close.className = "tl-close";
  close.type = "button";
  close.textContent = "×";
  close.title = "关闭";
  close.setAttribute("aria-label", "关闭");
  header.append(title, close);

  const body = document.createElement("div");
  body.className = "tl-body";
  panel.append(header, body);
  shadow.append(style, icon, panel);
  document.documentElement.appendChild(host);

  let selectionText = "";
  let anchorRect = null;

  function eventFromWidget(event) {
    try {
      return event.composedPath?.().includes(host) || event.target === host;
    } catch {
      return event.target === host;
    }
  }

  function hideIcon() {
    icon.style.display = "none";
  }

  function hidePanel() {
    panel.style.display = "none";
  }

  function hideAll() {
    hideIcon();
    hidePanel();
  }

  function placeNear(rect, width, height) {
    const margin = 8;
    let left = rect.right + 6;
    let top = rect.bottom + margin;

    if (left + width > window.innerWidth - margin) {
      left = window.innerWidth - width - margin;
    }
    if (left < margin) {
      left = margin;
    }
    if (top + height > window.innerHeight - margin) {
      top = Math.max(margin, rect.top - height - margin);
    }
    return { left: Math.max(margin, left), top };
  }

  function showIcon(rect, text) {
    selectionText = text;
    anchorRect = rect;
    const position = placeNear(rect, 32, 32);
    icon.style.left = `${position.left}px`;
    icon.style.top = `${position.top}px`;
    icon.style.display = "block";
  }

  function showPanel(rect, content, isError = false) {
    hideIcon();
    const panelWidth = Math.min(380, window.innerWidth - 16);
    const position = placeNear(rect, panelWidth, Math.min(440, window.innerHeight - 16));
    panel.style.left = `${position.left}px`;
    panel.style.top = `${position.top}px`;
    panel.style.display = "block";
    body.className = isError ? "tl-body tl-error" : "tl-body";
    body.replaceChildren(content);
  }

  function section(headingText, content) {
    const wrapper = document.createElement("div");
    wrapper.className = "tl-section";
    const heading = document.createElement("div");
    heading.className = "tl-heading";
    heading.textContent = headingText;
    const text = document.createElement("div");
    text.className = "tl-text";
    text.textContent = content;
    wrapper.append(heading, text);
    return wrapper;
  }

  function renderVocabulary(vocabulary) {
    const wrapper = document.createElement("div");
    const wordHeader = document.createElement("div");
    wordHeader.className = "tl-word-header";
    const parts = [vocabulary.word, vocabulary.partOfSpeech];
    if (vocabulary.category) {
      parts.push(vocabulary.category);
    }
    wordHeader.textContent = parts.join(" · ");
    wrapper.append(wordHeader, section("Definition", vocabulary.definition), section("中文", vocabulary.chinese));

    const examples = document.createElement("div");
    examples.className = "tl-section";
    const heading = document.createElement("div");
    heading.className = "tl-heading";
    heading.textContent = "Examples";
    const list = document.createElement("div");
    list.className = "tl-examples";
    for (const example of vocabulary.examples) {
      const item = document.createElement("div");
      item.className = "tl-example";
      item.textContent = example;
      list.appendChild(item);
    }
    examples.append(heading, list);
    wrapper.appendChild(examples);
    return wrapper;
  }

  function renderTranslation(result) {
    const wrapper = document.createElement("div");
    wrapper.append(
      section("Original", result.original),
      section("Translation", result.translation),
    );
    return wrapper;
  }

  function showLoading(rect) {
    title.textContent = "tl · 处理中";
    const loading = document.createElement("div");
    loading.className = "tl-loading";
    const spinner = document.createElement("span");
    spinner.className = "tl-spinner";
    const label = document.createElement("span");
    label.textContent = "处理中…";
    loading.append(spinner, label);
    showPanel(rect, loading);
  }

  function showResult(rect, result) {
    if (result.mode === "vocabulary" && result.vocabulary) {
      title.textContent = "English · 词典";
      showPanel(rect, renderVocabulary(result.vocabulary));
      return;
    }
    const languageLabels = { English: "English", Chinese: "中文" };
    const source = languageLabels[result.sourceLanguage] || result.sourceLanguage;
    const target = languageLabels[result.targetLanguage] || result.targetLanguage;
    title.textContent = `${source} → ${target}`;
    showPanel(rect, renderTranslation(result));
  }

  function showError(rect, message, needsSettings) {
    title.textContent = "tl · 出错";
    const content = document.createElement("div");
    content.textContent = message;
    if (needsSettings) {
      const settingsButton = document.createElement("button");
      settingsButton.className = "tl-open-options";
      settingsButton.type = "button";
      settingsButton.textContent = "打开设置";
      settingsButton.addEventListener("click", () => {
        browser.runtime.sendMessage({ type: "tl-open-options" });
      });
      content.appendChild(document.createElement("br"));
      content.appendChild(settingsButton);
    }
    showPanel(rect, content, true);
  }

  function selectionInfo() {
    const selection = window.getSelection();
    if (!selection || selection.rangeCount === 0 || selection.isCollapsed) {
      return null;
    }
    const text = selection.toString().trim();
    if (!text) {
      return null;
    }
    const range = selection.getRangeAt(0);
    const rects = Array.from(range.getClientRects()).filter((rect) => rect.width > 0 && rect.height > 0);
    const rect = rects.length > 0 ? rects[rects.length - 1] : range.getBoundingClientRect();
    if (!rect || (rect.width === 0 && rect.height === 0)) {
      return null;
    }
    return { text, rect };
  }

  function handleSelection(event) {
    if (eventFromWidget(event)) {
      return;
    }
    window.setTimeout(() => {
      const info = selectionInfo();
      if (!info) {
        hideIcon();
        return;
      }
      hidePanel();
      showIcon(info.rect, info.text);
    }, 10);
  }

  document.addEventListener("mouseup", handleSelection, true);
  document.addEventListener("keyup", (event) => {
    if (event.shiftKey || ["Shift", "ArrowLeft", "ArrowRight", "ArrowUp", "ArrowDown", "Home", "End"].includes(event.key)) {
      handleSelection(event);
    }
  }, true);
  document.addEventListener("selectionchange", () => {
    const selection = window.getSelection();
    if (!selection || selection.isCollapsed) {
      hideIcon();
    }
  }, true);
  document.addEventListener("mousedown", (event) => {
    if (!eventFromWidget(event)) {
      hideAll();
    }
  }, true);
  window.addEventListener("scroll", (event) => {
    if (!eventFromWidget(event)) {
      hideIcon();
    }
  }, true);
  window.addEventListener("resize", hideIcon, true);
  document.addEventListener("keydown", (event) => {
    if (event.key === "Escape") {
      hideAll();
    }
  }, true);

  icon.addEventListener("click", async (event) => {
    event.preventDefault();
    event.stopPropagation();
    const text = selectionText;
    const rect = anchorRect;
    if (!text || !rect) {
      return;
    }
    showLoading(rect);
    try {
      const result = await browser.runtime.sendMessage({ type: "tl-translate", text });
      if (result?.ok) {
        showResult(rect, result);
      } else {
        showError(rect, result?.error || "翻译失败，请稍后重试。", Boolean(result?.needsSettings));
      }
    } catch (error) {
      showError(rect, `扩展通信失败：${error?.message || error}`, false);
    }
  });

  close.addEventListener("click", (event) => {
    event.preventDefault();
    event.stopPropagation();
    hideAll();
  });
})();
