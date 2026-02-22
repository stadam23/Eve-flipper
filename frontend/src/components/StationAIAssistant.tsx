import { useCallback, useEffect, useMemo, useRef, useState, type PointerEvent as ReactPointerEvent } from "react";
import type {
  ScanParams,
  StationAIChatContext,
  StationAIHistoryMessage,
  StationAIScanSnapshot,
  StationAIUsage,
  StationCommandRow,
  StationTrade,
} from "@/lib/types";
import { stationAIChatStream } from "@/lib/api";
import { useI18n } from "@/lib/i18n";
import { useGlobalToast } from "./Toast";

const AI_CONFIG_STORAGE_KEY = "station.ai.config.v1";
const AI_OPEN_STORAGE_KEY = "station.ai.open.v1";
const AI_SIDEBAR_COLLAPSED_STORAGE_KEY = "station.ai.sidebar.collapsed.v1";
const AI_PROGRESS_COLLAPSED_STORAGE_KEY = "station.ai.progress.collapsed.v1";
const AI_CHAT_SESSIONS_STORAGE_KEY = "station.ai.sessions.v1";
const AI_ACTIVE_SESSION_STORAGE_KEY = "station.ai.active_session.v1";
const AI_MAX_SESSIONS = 40;
const AI_MAX_MESSAGES_PER_SESSION = 400;
const AI_ASSISTANT_NAME = "Ivy AI";
const AI_RELEASE_STAGE = "ALPHA";
const AI_MAX_TOKENS_LIMIT = 1_000_000;
const AI_WINDOW_MARGIN_PX = 12;
const AI_WINDOW_MIN_WIDTH_PX = 520;
const AI_WINDOW_MIN_HEIGHT_PX = 420;

type AIWindowRect = {
  left: number;
  top: number;
  width: number;
  height: number;
};

type AIResizeAxis = "e" | "s" | "se";

type AIWindowInteraction =
  | {
      kind: "move";
      pointerID: number;
      startX: number;
      startY: number;
      startRect: AIWindowRect;
    }
  | {
      kind: "resize";
      axis: AIResizeAxis;
      pointerID: number;
      startX: number;
      startY: number;
      startRect: AIWindowRect;
    };

type ChatRole = "user" | "assistant";

type ChatMessage = {
  id: number;
  role: ChatRole;
  text: string;
  createdAt: number;
};

type ChatSession = {
  id: string;
  title: string;
  createdAt: number;
  updatedAt: number;
  systemName: string;
  stationScope: string;
  model: string;
  messages: ChatMessage[];
};

function buildHistoryForRequest(messages: ChatMessage[]): StationAIHistoryMessage[] {
  if (!Array.isArray(messages) || messages.length === 0) {
    return [];
  }
  return messages
    .filter((m) => (m.role === "user" || m.role === "assistant") && m.text.trim().length > 0)
    .slice(-16)
    .map((m) => ({
      role: m.role,
      content: m.text.trim(),
    }));
}

type StationAIConfig = {
  provider: "openrouter";
  apiKey: string;
  model: string;
  useCustomModel: boolean;
  customModel: string;
  temperature: number;
  maxTokens: number;
  enableWikiContext: boolean;
  enableWebResearch: boolean;
  wikiRepo: string;
};

interface Props {
  params: ScanParams;
  rows: StationTrade[];
  totalRows: number;
  commandRowsByKey: Record<string, StationCommandRow>;
  regionID: number;
  selectedStationLabel: string;
  scanSnapshot: StationAIScanSnapshot;
  disabled?: boolean;
}

const OPENROUTER_MODELS = [
  "openai/gpt-4o-mini",
  "anthropic/claude-3.5-sonnet",
  "google/gemini-2.0-flash-exp",
  "meta-llama/llama-3.3-70b-instruct",
];

const DEFAULT_CONFIG: StationAIConfig = {
  provider: "openrouter",
  apiKey: "",
  model: OPENROUTER_MODELS[0],
  useCustomModel: false,
  customModel: "",
  temperature: 0.2,
  maxTokens: 900,
  enableWikiContext: true,
  enableWebResearch: false,
  wikiRepo: "https://github.com/ilyaux/Eve-flipper/wiki",
};

const IVY_MASCOT_SPRITE = [
  ".....oo.....",
  "....obbo....",
  "...obggbo...",
  "..obggggbo..",
  "..obgeegbo..",
  "..obggggbo..",
  "..obgaagbo..",
  "...obggbo...",
  "....obbo....",
  "...obbbbo...",
  "..o......o..",
  ".o........o.",
] as const;

type IvyPixelTone = "o" | "b" | "g" | "e" | "a";

type IvyPixel = {
  x: number;
  y: number;
  tone: IvyPixelTone;
  eye: boolean;
  mouth: boolean;
};

const IVY_MASCOT_PIXELS: IvyPixel[] = (() => {
  const out: IvyPixel[] = [];
  for (let y = 0; y < IVY_MASCOT_SPRITE.length; y++) {
    const row = IVY_MASCOT_SPRITE[y];
    for (let x = 0; x < row.length; x++) {
      const tone = row[x];
      if (tone === ".") {
        continue;
      }
      if (tone === "o" || tone === "b" || tone === "g" || tone === "e" || tone === "a") {
        const mouth = tone === "a" && y === 6;
        out.push({ x, y, tone, eye: tone === "e", mouth });
      }
    }
  }
  return out;
})();

function IvyAIMascot({ thinking }: { thinking: boolean }) {
  return (
    <div
      aria-hidden="true"
      className={`ai-ivy-mascot${thinking ? " ai-ivy-mascot--thinking" : ""}`}
    >
      <div className="ai-ivy-mascot-grid">
        {IVY_MASCOT_PIXELS.map((pixel) => (
          <span
            key={`${pixel.x}-${pixel.y}`}
            className={`ai-ivy-pixel ai-ivy-pixel--${pixel.tone}${
              pixel.eye ? " ai-ivy-pixel--eye" : ""
            }${pixel.mouth ? " ai-ivy-pixel--mouth" : ""}`}
            style={{ gridColumnStart: pixel.x + 1, gridRowStart: pixel.y + 1 }}
          />
        ))}
      </div>
    </div>
  );
}

function stationRowKey(row: StationTrade): string {
  return `${row.TypeID}-${row.StationID}`;
}

function loadAIConfig(): StationAIConfig {
  if (typeof window === "undefined") return DEFAULT_CONFIG;
  try {
    const raw = window.localStorage.getItem(AI_CONFIG_STORAGE_KEY);
    if (!raw) return DEFAULT_CONFIG;
    const parsed = JSON.parse(raw) as Partial<StationAIConfig>;
    return {
      ...DEFAULT_CONFIG,
      ...parsed,
      provider: "openrouter",
      temperature: Number.isFinite(parsed.temperature)
        ? Math.max(0, Math.min(2, Number(parsed.temperature)))
        : DEFAULT_CONFIG.temperature,
      maxTokens: Number.isFinite(parsed.maxTokens)
        ? Math.max(200, Math.min(AI_MAX_TOKENS_LIMIT, Number(parsed.maxTokens)))
        : DEFAULT_CONFIG.maxTokens,
      enableWikiContext:
        typeof parsed.enableWikiContext === "boolean"
          ? parsed.enableWikiContext
          : DEFAULT_CONFIG.enableWikiContext,
      enableWebResearch:
        typeof parsed.enableWebResearch === "boolean"
          ? parsed.enableWebResearch
          : DEFAULT_CONFIG.enableWebResearch,
      wikiRepo:
        typeof parsed.wikiRepo === "string" && parsed.wikiRepo.trim()
          ? parsed.wikiRepo.trim()
          : DEFAULT_CONFIG.wikiRepo,
    };
  } catch {
    return DEFAULT_CONFIG;
  }
}

function saveAIConfig(cfg: StationAIConfig) {
  if (typeof window === "undefined") return;
  try {
    window.localStorage.setItem(AI_CONFIG_STORAGE_KEY, JSON.stringify(cfg));
  } catch {
    // Keep runtime stable if storage quota is exceeded.
  }
}

function loadOpenState(): boolean {
  if (typeof window === "undefined") return false;
  return window.localStorage.getItem(AI_OPEN_STORAGE_KEY) === "1";
}

function saveOpenState(open: boolean) {
  if (typeof window === "undefined") return;
  try {
    window.localStorage.setItem(AI_OPEN_STORAGE_KEY, open ? "1" : "0");
  } catch {
    // Keep runtime stable if storage is unavailable.
  }
}

function loadSidebarCollapsedState(): boolean {
  if (typeof window === "undefined") return false;
  return window.localStorage.getItem(AI_SIDEBAR_COLLAPSED_STORAGE_KEY) === "1";
}

function saveSidebarCollapsedState(collapsed: boolean) {
  if (typeof window === "undefined") return;
  try {
    window.localStorage.setItem(AI_SIDEBAR_COLLAPSED_STORAGE_KEY, collapsed ? "1" : "0");
  } catch {
    // Keep runtime stable if storage is unavailable.
  }
}

function loadProgressCollapsedState(): boolean {
  if (typeof window === "undefined") return false;
  return window.localStorage.getItem(AI_PROGRESS_COLLAPSED_STORAGE_KEY) === "1";
}

function saveProgressCollapsedState(collapsed: boolean) {
  if (typeof window === "undefined") return;
  try {
    window.localStorage.setItem(AI_PROGRESS_COLLAPSED_STORAGE_KEY, collapsed ? "1" : "0");
  } catch {
    // Keep runtime stable if storage is unavailable.
  }
}

function makeSessionID(): string {
  return `s_${Date.now()}_${Math.random().toString(36).slice(2, 8)}`;
}

function deriveSessionTitle(text: string, fallback: string): string {
  const trimmed = text.trim();
  if (!trimmed) return fallback;
  const firstLine = trimmed.split(/\r?\n/, 1)[0].trim();
  if (!firstLine) return fallback;
  return firstLine.length > 68 ? `${firstLine.slice(0, 68)}...` : firstLine;
}

function createEmptySession(
  systemName: string,
  stationScope: string,
  model: string,
  title = "New chat",
): ChatSession {
  const now = Date.now();
  return {
    id: makeSessionID(),
    title,
    createdAt: now,
    updatedAt: now,
    systemName,
    stationScope,
    model,
    messages: [],
  };
}

function sanitizeMessage(raw: unknown): ChatMessage | null {
  if (!raw || typeof raw !== "object") return null;
  const obj = raw as Partial<ChatMessage>;
  if (obj.role !== "user" && obj.role !== "assistant") return null;
  if (typeof obj.text !== "string") return null;
  const createdAt = Number.isFinite(obj.createdAt)
    ? Number(obj.createdAt)
    : Date.now();
  const id = Number.isFinite(obj.id) ? Number(obj.id) : createdAt;
  return {
    id,
    role: obj.role,
    text: obj.text,
    createdAt,
  };
}

function sanitizeSession(raw: unknown): ChatSession | null {
  if (!raw || typeof raw !== "object") return null;
  const obj = raw as Partial<ChatSession>;
  if (typeof obj.id !== "string" || !obj.id.trim()) return null;
  const messagesRaw = Array.isArray(obj.messages) ? obj.messages : [];
  const messages = messagesRaw
    .map((m) => sanitizeMessage(m))
    .filter((m): m is ChatMessage => m !== null)
    .slice(-AI_MAX_MESSAGES_PER_SESSION);
  const createdAt = Number.isFinite(obj.createdAt)
    ? Number(obj.createdAt)
    : Date.now();
  const updatedAt = Number.isFinite(obj.updatedAt)
    ? Number(obj.updatedAt)
    : createdAt;
  return {
    id: obj.id,
    title: typeof obj.title === "string" && obj.title.trim() ? obj.title : "New chat",
    createdAt,
    updatedAt,
    systemName: typeof obj.systemName === "string" ? obj.systemName : "",
    stationScope: typeof obj.stationScope === "string" ? obj.stationScope : "",
    model: typeof obj.model === "string" ? obj.model : "",
    messages,
  };
}

function loadChatSessions(): ChatSession[] {
  if (typeof window === "undefined") return [];
  try {
    const raw = window.localStorage.getItem(AI_CHAT_SESSIONS_STORAGE_KEY);
    if (!raw) return [];
    const parsed = JSON.parse(raw) as unknown;
    if (!Array.isArray(parsed)) return [];
    return parsed
      .map((s) => sanitizeSession(s))
      .filter((s): s is ChatSession => s !== null)
      .sort((a, b) => b.updatedAt - a.updatedAt)
      .slice(0, AI_MAX_SESSIONS);
  } catch {
    return [];
  }
}

function saveChatSessions(sessions: ChatSession[]) {
  if (typeof window === "undefined") return;
  const trimmed = sessions.slice(0, AI_MAX_SESSIONS);
  try {
    window.localStorage.setItem(AI_CHAT_SESSIONS_STORAGE_KEY, JSON.stringify(trimmed));
    return;
  } catch {
    // Fallback for quota pressure: keep only a shorter tail of messages.
  }

  try {
    const compact = trimmed.map((session) => ({
      ...session,
      messages: session.messages.slice(-120),
    }));
    window.localStorage.setItem(AI_CHAT_SESSIONS_STORAGE_KEY, JSON.stringify(compact));
  } catch {
    // Ignore storage errors; chat continues in memory.
  }
}

function loadActiveSessionID(): string {
  if (typeof window === "undefined") return "";
  return window.localStorage.getItem(AI_ACTIVE_SESSION_STORAGE_KEY) ?? "";
}

function saveActiveSessionID(id: string) {
  if (typeof window === "undefined") return;
  try {
    if (!id) {
      window.localStorage.removeItem(AI_ACTIVE_SESSION_STORAGE_KEY);
      return;
    }
    window.localStorage.setItem(AI_ACTIVE_SESSION_STORAGE_KEY, id);
  } catch {
    // Ignore storage errors; active session remains in component state.
  }
}

function escapeHtml(text: string): string {
  return text
    .replace(/&/g, "&amp;")
    .replace(/</g, "&lt;")
    .replace(/>/g, "&gt;")
    .replace(/\"/g, "&quot;")
    .replace(/'/g, "&#39;");
}

function sanitizeSafeURL(rawURL: string): string | null {
  try {
    const parsed = new URL(rawURL);
    if (parsed.protocol !== "http:" && parsed.protocol !== "https:") {
      return null;
    }
    return parsed.toString();
  } catch {
    return null;
  }
}

function renderInlineMarkdown(text: string): string {
  let out = escapeHtml(text);
  const inlineCode: string[] = [];

  out = out.replace(/`([^`\n]+)`/g, (_m, code: string) => {
    const token = `@@INLINE_CODE_${inlineCode.length}@@`;
    inlineCode.push(
      `<code class="rounded bg-eve-dark/80 border border-eve-border/50 px-1 py-0.5 font-mono text-[11px]">${code}</code>`,
    );
    return token;
  });

  out = out.replace(/\*\*([^*]+)\*\*/g, "<strong>$1</strong>");
  out = out.replace(/\*([^*]+)\*/g, "<em>$1</em>");

  out = out.replace(/\[([^\]]+)\]\(([^)\s]+)\)/g, (_m, label: string, url: string) => {
    const decodedURL = url.replace(/&amp;/g, "&");
    const safeURL = sanitizeSafeURL(decodedURL);
    if (!safeURL) {
      return label;
    }
    return `<a href="${escapeHtml(safeURL)}" target="_blank" rel="noopener noreferrer nofollow" class="text-eve-accent underline decoration-eve-accent/40">${label}</a>`;
  });

  out = out.replace(/@@INLINE_CODE_(\d+)@@/g, (_m, idx: string) => {
    const i = Number(idx);
    return inlineCode[i] ?? "";
  });

  return out;
}

function parseMarkdownTableRow(line: string): string[] {
  const trimmed = line.trim();
  const withoutOuterPipes = trimmed.replace(/^\|/, "").replace(/\|$/, "");
  return withoutOuterPipes.split("|").map((cell) => cell.trim());
}

function isMarkdownTableSeparator(line: string): boolean {
  const cells = parseMarkdownTableRow(line);
  if (cells.length === 0) {
    return false;
  }
  return cells.every((cell) => /^:?-{3,}:?$/.test(cell.replace(/\s+/g, "")));
}

function renderMarkdownTable(headers: string[], rows: string[][]): string {
  const cols = headers.length;
  if (cols === 0) {
    return "";
  }
  const safeHeaders = headers.map((h) => renderInlineMarkdown(h));
  const safeRows = rows
    .filter((row) => row.some((cell) => cell.trim() !== ""))
    .map((row) => {
      const normalized = row.slice(0, cols);
      while (normalized.length < cols) {
        normalized.push("");
      }
      return normalized.map((cell) => renderInlineMarkdown(cell));
    });

  const thead = `<thead><tr>${safeHeaders
    .map(
      (header) =>
        `<th class="px-2 py-1 text-left font-semibold text-eve-accent border-b border-eve-border/60 bg-eve-dark/45">${header}</th>`,
    )
    .join("")}</tr></thead>`;
  const tbody = `<tbody>${safeRows
    .map(
      (row) =>
        `<tr>${row
          .map(
            (cell) =>
              `<td class="px-2 py-1 align-top border-b border-eve-border/40 text-eve-text">${cell}</td>`,
          )
          .join("")}</tr>`,
    )
    .join("")}</tbody>`;

  return `<div class="my-2 overflow-x-auto rounded-sm border border-eve-border/60 bg-eve-dark/35"><table class="min-w-full text-[11px] leading-5 border-collapse">${thead}${tbody}</table></div>`;
}

function renderMarkdownToHtml(markdown: string): string {
  const normalized = markdown.replace(/\r\n/g, "\n");
  const codeBlocks: string[] = [];
  const text = normalized.replace(
    /```([a-zA-Z0-9_-]+)?\n?([\s\S]*?)```/g,
    (_m, lang: string | undefined, code: string) => {
      const token = `@@CODE_BLOCK_${codeBlocks.length}@@`;
      const langLabel = lang ? escapeHtml(lang) : "";
      const codeEscaped = escapeHtml(code.replace(/\n$/, ""));
      const langHeader = langLabel
        ? `<div class="px-2 py-1 border-b border-eve-border/40 text-[10px] uppercase tracking-wide text-eve-dim">${langLabel}</div>`
        : "";
      codeBlocks.push(
        `<div class="rounded-sm border border-eve-border/60 bg-eve-dark/80 my-2 overflow-hidden">${langHeader}<pre class="p-2 overflow-x-auto text-[11px] leading-5 text-eve-text"><code>${codeEscaped}</code></pre></div>`,
      );
      return token;
    },
  );

  const lines = text.split("\n");
  const out: string[] = [];
  let inUL = false;
  let inOL = false;

  const closeLists = () => {
    if (inUL) {
      out.push("</ul>");
      inUL = false;
    }
    if (inOL) {
      out.push("</ol>");
      inOL = false;
    }
  };

  let idx = 0;
  while (idx < lines.length) {
    const rawLine = lines[idx];
    const line = rawLine.trim();
    if (!line) {
      closeLists();
      idx += 1;
      continue;
    }

    const codeTokenMatch = line.match(/^@@CODE_BLOCK_(\d+)@@$/);
    if (codeTokenMatch) {
      closeLists();
      const tokenIndex = Number(codeTokenMatch[1]);
      out.push(codeBlocks[tokenIndex] ?? "");
      idx += 1;
      continue;
    }

    // Markdown pipe-table support.
    if (line.includes("|")) {
      let separatorIdx = idx + 1;
      while (separatorIdx < lines.length && lines[separatorIdx].trim() === "") {
        separatorIdx += 1;
      }
      const separatorLine = separatorIdx < lines.length ? lines[separatorIdx].trim() : "";
      if (isMarkdownTableSeparator(separatorLine)) {
        closeLists();
        const headers = parseMarkdownTableRow(line);
        const rows: string[][] = [];
        idx = separatorIdx + 1;
        while (idx < lines.length) {
          const rowLine = lines[idx].trim();
          if (!rowLine) {
            break;
          }
          if (!rowLine.includes("|")) {
            break;
          }
          rows.push(parseMarkdownTableRow(rowLine));
          idx += 1;
        }
        out.push(renderMarkdownTable(headers, rows));
        continue;
      }
    }

    const heading = line.match(/^(#{1,3})\s+(.+)$/);
    if (heading) {
      closeLists();
      const level = heading[1].length;
      const tag = level === 1 ? "h1" : level === 2 ? "h2" : "h3";
      const cls =
        level === 1
          ? "text-sm font-semibold text-eve-text mt-2 mb-1"
          : "text-xs font-semibold text-eve-text mt-2 mb-1";
      out.push(`<${tag} class="${cls}">${renderInlineMarkdown(heading[2])}</${tag}>`);
      idx += 1;
      continue;
    }

    const ul = line.match(/^[-*]\s+(.+)$/);
    if (ul) {
      if (inOL) {
        out.push("</ol>");
        inOL = false;
      }
      if (!inUL) {
        out.push('<ul class="list-disc pl-4 my-1 space-y-0.5">');
        inUL = true;
      }
      out.push(`<li>${renderInlineMarkdown(ul[1])}</li>`);
      idx += 1;
      continue;
    }

    const ol = line.match(/^\d+\.\s+(.+)$/);
    if (ol) {
      if (inUL) {
        out.push("</ul>");
        inUL = false;
      }
      if (!inOL) {
        out.push('<ol class="list-decimal pl-4 my-1 space-y-0.5">');
        inOL = true;
      }
      out.push(`<li>${renderInlineMarkdown(ol[1])}</li>`);
      idx += 1;
      continue;
    }

    closeLists();
    out.push(`<p class="my-1">${renderInlineMarkdown(line)}</p>`);
    idx += 1;
  }

  closeLists();
  return out.join("");
}

function tryParseStandaloneJSON(text: string): unknown | null {
  const trimmed = text.trim();
  if (!trimmed) return null;
  const isJSONObject = trimmed.startsWith("{") && trimmed.endsWith("}");
  const isJSONArray = trimmed.startsWith("[") && trimmed.endsWith("]");
  if (!isJSONObject && !isJSONArray) {
    return null;
  }
  try {
    return JSON.parse(trimmed);
  } catch {
    return null;
  }
}

function renderStandaloneJSONToHtml(value: unknown): string {
  const pretty = JSON.stringify(value, null, 2);
  const escaped = escapeHtml(pretty);
  const highlighted = escaped.replace(
    /(&quot;(?:\\.|[^"\\])*&quot;)(?=\s*:)|(&quot;(?:\\.|[^"\\])*&quot;)|\b(true|false|null)\b|(-?\d+(?:\.\d+)?(?:[eE][+\-]?\d+)?)/g,
    (_m, keyToken: string, stringToken: string, boolNullToken: string, numberToken: string) => {
      if (keyToken) return `<span class="text-eve-accent">${keyToken}</span>`;
      if (stringToken) return `<span class="text-sky-300">${stringToken}</span>`;
      if (boolNullToken) return `<span class="text-orange-300">${boolNullToken}</span>`;
      if (numberToken) return `<span class="text-emerald-300">${numberToken}</span>`;
      return _m;
    },
  );
  return (
    `<div class="my-2 overflow-hidden rounded-sm border border-eve-border/60 bg-eve-dark/80">` +
    `<div class="px-2 py-1 border-b border-eve-border/40 text-[10px] uppercase tracking-wide text-eve-dim">JSON</div>` +
    `<pre class="p-2 overflow-x-auto text-[11px] leading-5 text-eve-text"><code>${highlighted}</code></pre>` +
    `</div>`
  );
}

function MarkdownMessage({ text }: { text: string }) {
  const html = useMemo(() => {
    const parsedJSON = tryParseStandaloneJSON(text);
    if (parsedJSON !== null) {
      return renderStandaloneJSONToHtml(parsedJSON);
    }
    return renderMarkdownToHtml(text);
  }, [text]);
  return (
    <div
      className="text-xs leading-5 text-eve-text select-text"
      dangerouslySetInnerHTML={{ __html: html }}
    />
  );
}

function formatSessionTimestamp(ts: number, locale: "ru" | "en"): string {
  const d = new Date(ts);
  if (Number.isNaN(d.getTime())) return "";
  return new Intl.DateTimeFormat(locale === "ru" ? "ru-RU" : "en-US", {
    month: "2-digit",
    day: "2-digit",
    hour: "2-digit",
    minute: "2-digit",
  }).format(d);
}

function isTechnicalAIWarning(text: string): boolean {
  const lower = text.trim().toLowerCase();
  if (!lower) return false;
  const patterns = [
    "planner read failed",
    "planner provider error",
    "planner unavailable",
    "planner invalid response",
    "planner did not return json",
    "planner json parse failed",
    "fallback intent routing",
    "preflight",
    "server validation requested retry",
    "retry failed",
    "retry rejected",
  ];
  return patterns.some((p) => lower.includes(p));
}

function firstUserFacingWarning(warnings: unknown): string {
  if (!Array.isArray(warnings)) return "";
  for (const raw of warnings) {
    if (typeof raw !== "string") continue;
    const msg = raw.trim();
    if (!msg) continue;
    if (isTechnicalAIWarning(msg)) continue;
    return msg;
  }
  return "";
}

function defaultAIWindowRect(): AIWindowRect {
  if (typeof window === "undefined") {
    return { left: 16, top: 16, width: 860, height: 700 };
  }
  const vw = Math.max(320, window.innerWidth);
  const vh = Math.max(320, window.innerHeight);
  const width = Math.min(860, Math.floor(vw * 0.96));
  const height = Math.min(700, Math.floor(vh * 0.82));
  return {
    left: Math.max(AI_WINDOW_MARGIN_PX, vw - width - 16),
    top: Math.max(AI_WINDOW_MARGIN_PX, vh - height - 16),
    width,
    height,
  };
}

function clampAIWindowRect(rect: AIWindowRect, vw: number, vh: number): AIWindowRect {
  const maxWidth = Math.max(320, vw - AI_WINDOW_MARGIN_PX * 2);
  const maxHeight = Math.max(320, vh - AI_WINDOW_MARGIN_PX * 2);
  const minWidth = Math.min(AI_WINDOW_MIN_WIDTH_PX, maxWidth);
  const minHeight = Math.min(AI_WINDOW_MIN_HEIGHT_PX, maxHeight);

  const width = Math.max(minWidth, Math.min(maxWidth, rect.width));
  const height = Math.max(minHeight, Math.min(maxHeight, rect.height));

  const maxLeft = Math.max(AI_WINDOW_MARGIN_PX, vw - AI_WINDOW_MARGIN_PX - width);
  const maxTop = Math.max(AI_WINDOW_MARGIN_PX, vh - AI_WINDOW_MARGIN_PX - height);
  const left = Math.max(AI_WINDOW_MARGIN_PX, Math.min(maxLeft, rect.left));
  const top = Math.max(AI_WINDOW_MARGIN_PX, Math.min(maxTop, rect.top));

  return { left, top, width, height };
}

export function StationAIAssistant({
  params,
  rows,
  totalRows,
  commandRowsByKey,
  regionID,
  selectedStationLabel,
  scanSnapshot,
  disabled = false,
}: Props) {
  const { t, locale } = useI18n();
  const { addToast } = useGlobalToast();
  const [open, setOpen] = useState<boolean>(loadOpenState);
  const [sessionsCollapsed, setSessionsCollapsed] = useState<boolean>(
    loadSidebarCollapsedState,
  );
  const [configOpen, setConfigOpen] = useState(false);
  const [promptToolsOpen, setPromptToolsOpen] = useState(false);
  const [thinking, setThinking] = useState(false);
  const [input, setInput] = useState("");
  const [cfg, setCfg] = useState<StationAIConfig>(loadAIConfig);
  const [nextPromptWiki, setNextPromptWiki] = useState<boolean>(
    DEFAULT_CONFIG.enableWikiContext,
  );
  const [nextPromptWeb, setNextPromptWeb] = useState<boolean>(
    DEFAULT_CONFIG.enableWebResearch,
  );
  const [sessions, setSessions] = useState<ChatSession[]>(loadChatSessions);
  const [activeSessionID, setActiveSessionID] = useState<string>(loadActiveSessionID);
  const [progressPct, setProgressPct] = useState(0);
  const [progressText, setProgressText] = useState("");
  const [progressCollapsed, setProgressCollapsed] = useState<boolean>(
    loadProgressCollapsedState,
  );
  const [promptTokensEst, setPromptTokensEst] = useState(0);
  const [completionTokensEst, setCompletionTokensEst] = useState(0);
  const [totalTokensEst, setTotalTokensEst] = useState(0);
  const [usage, setUsage] = useState<StationAIUsage | null>(null);
  const [elapsedSec, setElapsedSec] = useState(0);
  const [windowRect, setWindowRect] = useState<AIWindowRect>(() => {
    const base = defaultAIWindowRect();
    if (typeof window === "undefined") {
      return base;
    }
    return clampAIWindowRect(base, window.innerWidth, window.innerHeight);
  });
  const endRef = useRef<HTMLDivElement | null>(null);
  const abortRef = useRef<AbortController | null>(null);
  const elapsedTimerRef = useRef<number | null>(null);
  const saveSessionsTimerRef = useRef<number | null>(null);
  const sendInFlightRef = useRef(false);
  const messageSeqRef = useRef<number>(Date.now() * 1000);
  const promptToolsRef = useRef<HTMLDivElement | null>(null);
  const interactionRef = useRef<AIWindowInteraction | null>(null);

  const effectiveModel = useMemo(
    () =>
      cfg.useCustomModel
        ? cfg.customModel.trim() || cfg.model
        : cfg.model.trim(),
    [cfg.customModel, cfg.model, cfg.useCustomModel],
  );

  const createSession = useCallback(
    (seedTitle?: string) => {
      const title = seedTitle?.trim() || t("aiSessionUntitled");
      const created = createEmptySession(
        params.system_name || "",
        selectedStationLabel,
        effectiveModel || cfg.model,
        title,
      );
      setSessions((prev) => [created, ...prev].slice(0, AI_MAX_SESSIONS));
      setActiveSessionID(created.id);
      return created.id;
    },
    [cfg.model, effectiveModel, params.system_name, selectedStationLabel, t],
  );

  const activeSession = useMemo(
    () => sessions.find((s) => s.id === activeSessionID) ?? sessions[0] ?? null,
    [activeSessionID, sessions],
  );

  const messages = activeSession?.messages ?? [];

  useEffect(() => {
    saveOpenState(open);
  }, [open]);

  useEffect(() => {
    saveSidebarCollapsedState(sessionsCollapsed);
  }, [sessionsCollapsed]);

  useEffect(() => {
    saveProgressCollapsedState(progressCollapsed);
  }, [progressCollapsed]);

  useEffect(() => {
    saveAIConfig(cfg);
  }, [cfg]);

  useEffect(() => {
    setNextPromptWiki(cfg.enableWikiContext);
  }, [cfg.enableWikiContext]);

  useEffect(() => {
    setNextPromptWeb(cfg.enableWebResearch);
  }, [cfg.enableWebResearch]);

  useEffect(() => {
    if (typeof window === "undefined") return;
    if (saveSessionsTimerRef.current !== null) {
      window.clearTimeout(saveSessionsTimerRef.current);
      saveSessionsTimerRef.current = null;
    }
    saveSessionsTimerRef.current = window.setTimeout(() => {
      saveChatSessions(sessions);
      saveSessionsTimerRef.current = null;
    }, 250);
    return () => {
      if (saveSessionsTimerRef.current !== null) {
        window.clearTimeout(saveSessionsTimerRef.current);
        saveSessionsTimerRef.current = null;
      }
    };
  }, [sessions]);

  useEffect(() => {
    saveActiveSessionID(activeSessionID);
  }, [activeSessionID]);

  useEffect(() => {
    if (sessions.length === 0) {
      createSession(t("aiSessionUntitled"));
      return;
    }
    if (!activeSessionID || !sessions.some((s) => s.id === activeSessionID)) {
      setActiveSessionID(sessions[0].id);
    }
  }, [activeSessionID, createSession, sessions, t]);

  useEffect(() => {
    if (!open) return;
    endRef.current?.scrollIntoView({ behavior: "smooth", block: "end" });
  }, [messages, open, thinking]);

  useEffect(() => {
    return () => {
      abortRef.current?.abort();
      if (elapsedTimerRef.current !== null) {
        window.clearInterval(elapsedTimerRef.current);
      }
      if (saveSessionsTimerRef.current !== null) {
        window.clearTimeout(saveSessionsTimerRef.current);
      }
    };
  }, []);

  useEffect(() => {
    if (!promptToolsOpen) return;
    const onPointerDown = (event: MouseEvent) => {
      const node = promptToolsRef.current;
      if (!node) return;
      if (!node.contains(event.target as Node)) {
        setPromptToolsOpen(false);
      }
    };
    window.addEventListener("mousedown", onPointerDown);
    return () => window.removeEventListener("mousedown", onPointerDown);
  }, [promptToolsOpen]);

  const stopWindowInteraction = useCallback(() => {
    interactionRef.current = null;
    if (typeof document !== "undefined") {
      document.body.style.userSelect = "";
      document.body.style.cursor = "";
    }
  }, []);

  const startWindowMove = useCallback(
    (event: ReactPointerEvent<HTMLElement>) => {
      if (!open || event.button !== 0 || event.pointerType === "touch") return;
      const target = event.target as HTMLElement | null;
      if (target?.closest("button, input, textarea, select, a, [role='button']")) {
        return;
      }
      event.preventDefault();
      interactionRef.current = {
        kind: "move",
        pointerID: event.pointerId,
        startX: event.clientX,
        startY: event.clientY,
        startRect: windowRect,
      };
      if (typeof document !== "undefined") {
        document.body.style.userSelect = "none";
        document.body.style.cursor = "grabbing";
      }
    },
    [open, windowRect],
  );

  const startWindowResize = useCallback(
    (axis: AIResizeAxis) => (event: ReactPointerEvent<HTMLDivElement>) => {
      if (!open || event.button !== 0 || event.pointerType === "touch") return;
      event.preventDefault();
      interactionRef.current = {
        kind: "resize",
        axis,
        pointerID: event.pointerId,
        startX: event.clientX,
        startY: event.clientY,
        startRect: windowRect,
      };
      if (typeof document !== "undefined") {
        document.body.style.userSelect = "none";
        document.body.style.cursor =
          axis === "e" ? "ew-resize" : axis === "s" ? "ns-resize" : "nwse-resize";
      }
    },
    [open, windowRect],
  );

  useEffect(() => {
    const onPointerMove = (event: PointerEvent) => {
      const active = interactionRef.current;
      if (!active || active.pointerID !== event.pointerId) return;
      event.preventDefault();

      const dx = event.clientX - active.startX;
      const dy = event.clientY - active.startY;
      const vw = Math.max(320, window.innerWidth);
      const vh = Math.max(320, window.innerHeight);

      if (active.kind === "move") {
        const candidate: AIWindowRect = {
          ...active.startRect,
          left: active.startRect.left + dx,
          top: active.startRect.top + dy,
        };
        setWindowRect(clampAIWindowRect(candidate, vw, vh));
        return;
      }

      const maxWidth = Math.max(320, vw - AI_WINDOW_MARGIN_PX * 2);
      const maxHeight = Math.max(320, vh - AI_WINDOW_MARGIN_PX * 2);
      const minWidth = Math.min(AI_WINDOW_MIN_WIDTH_PX, maxWidth);
      const minHeight = Math.min(AI_WINDOW_MIN_HEIGHT_PX, maxHeight);
      const maxWidthFromLeft = Math.max(
        minWidth,
        vw - AI_WINDOW_MARGIN_PX - active.startRect.left,
      );
      const maxHeightFromTop = Math.max(
        minHeight,
        vh - AI_WINDOW_MARGIN_PX - active.startRect.top,
      );
      let width = active.startRect.width;
      let height = active.startRect.height;
      if (active.axis === "e" || active.axis === "se") {
        width = Math.max(minWidth, Math.min(maxWidthFromLeft, active.startRect.width + dx));
      }
      if (active.axis === "s" || active.axis === "se") {
        height = Math.max(minHeight, Math.min(maxHeightFromTop, active.startRect.height + dy));
      }
      const candidate: AIWindowRect = { ...active.startRect, width, height };
      setWindowRect(clampAIWindowRect(candidate, vw, vh));
    };

    const onPointerUp = (event: PointerEvent) => {
      const active = interactionRef.current;
      if (!active || active.pointerID !== event.pointerId) return;
      stopWindowInteraction();
    };

    window.addEventListener("pointermove", onPointerMove, { passive: false });
    window.addEventListener("pointerup", onPointerUp);
    window.addEventListener("pointercancel", onPointerUp);
    return () => {
      window.removeEventListener("pointermove", onPointerMove);
      window.removeEventListener("pointerup", onPointerUp);
      window.removeEventListener("pointercancel", onPointerUp);
    };
  }, [stopWindowInteraction]);

  useEffect(() => {
    const onResize = () => {
      const vw = Math.max(320, window.innerWidth);
      const vh = Math.max(320, window.innerHeight);
      setWindowRect((prev) => clampAIWindowRect(prev, vw, vh));
    };
    window.addEventListener("resize", onResize);
    return () => window.removeEventListener("resize", onResize);
  }, []);

  useEffect(() => {
    if (!open) {
      stopWindowInteraction();
    }
  }, [open, stopWindowInteraction]);

  useEffect(() => {
    return () => {
      stopWindowInteraction();
    };
  }, [stopWindowInteraction]);

  const quickPrompts = useMemo(
    () => [
      t("aiQuickTopActions"),
      t("aiQuickExplainSelection"),
      t("aiQuickTightFilters"),
      t("aiQuickRiskCheck"),
    ],
    [t],
  );

  useEffect(() => {
    let maxID = messageSeqRef.current;
    for (const session of sessions) {
      for (const msg of session.messages) {
        if (Number.isFinite(msg.id) && msg.id > maxID) {
          maxID = msg.id;
        }
      }
    }
    messageSeqRef.current = maxID;
  }, [sessions]);

  const nextMessageID = useCallback(() => {
    messageSeqRef.current += 1;
    return messageSeqRef.current;
  }, []);

  const buildContext = (): StationAIChatContext => {
    const topRows = rows.slice(0, 100);
    const hasCommandMap = Object.keys(commandRowsByKey).length > 0;
    let highRiskRows = 0;
    let extremeRows = 0;
    let avgCTS = 0;
    let avgMargin = 0;
    let avgDailyProfit = 0;
    let avgDailyVolume = 0;
    let actionableRows = 0;

    const contextRows = topRows.map((row) => {
      const cmd = commandRowsByKey[stationRowKey(row)];
      const action = cmd?.recommended_action ?? (hasCommandMap ? "hold" : "scan_candidate");
      if (action !== "hold") actionableRows++;
      if (row.IsHighRiskFlag) highRiskRows++;
      if (row.IsExtremePriceFlag) extremeRows++;
      avgCTS += row.CTS || 0;
      avgMargin += row.MarginPercent || 0;
      avgDailyProfit += row.DailyProfit || row.RealizableDailyProfit || 0;
      avgDailyVolume += row.DailyVolume || 0;
      return {
        type_id: row.TypeID,
        type_name: row.TypeName,
        station_name: row.StationName,
        cts: row.CTS || 0,
        margin_percent: row.MarginPercent || 0,
        daily_profit: row.DailyProfit || row.RealizableDailyProfit || 0,
        daily_volume: row.DailyVolume || 0,
        s2b_bfs_ratio: row.S2BBfSRatio || 0,
        action,
        reason: cmd?.action_reason || "",
        confidence: row.ConfidenceLabel || "",
        high_risk: !!row.IsHighRiskFlag,
        extreme_price: !!row.IsExtremePriceFlag,
      };
    });

    const denom = contextRows.length > 0 ? contextRows.length : 1;
    return {
      tab_id: "station_trading",
      tab_title: "Station Trading",
      system_name: scanSnapshot.system_name || params.system_name || "",
      station_scope: selectedStationLabel,
      region_id: scanSnapshot.region_id || regionID,
      station_id: scanSnapshot.station_id,
      radius: scanSnapshot.radius,
      min_margin: scanSnapshot.min_margin,
      min_daily_volume: scanSnapshot.min_daily_volume,
      min_item_profit: scanSnapshot.min_item_profit,
      scan_snapshot: scanSnapshot,
      summary: {
        total_rows: Math.max(totalRows, rows.length),
        visible_rows: rows.length,
        high_risk_rows: highRiskRows,
        extreme_rows: extremeRows,
        avg_cts: avgCTS / denom,
        avg_margin: avgMargin / denom,
        avg_daily_profit: avgDailyProfit / denom,
        avg_daily_volume: avgDailyVolume / denom,
        actionable_rows: actionableRows,
      },
      rows: contextRows,
    };
  };

  const resetProgress = useCallback(() => {
    setProgressPct(0);
    setProgressText("");
    setPromptTokensEst(0);
    setCompletionTokensEst(0);
    setTotalTokensEst(0);
    setUsage(null);
    setElapsedSec(0);
  }, []);

  const startElapsedTimer = useCallback(() => {
    if (elapsedTimerRef.current !== null) {
      window.clearInterval(elapsedTimerRef.current);
    }
    const startedAt = Date.now();
    setElapsedSec(0);
    elapsedTimerRef.current = window.setInterval(() => {
      setElapsedSec(Math.max(0, Math.floor((Date.now() - startedAt) / 1000)));
    }, 1000);
  }, []);

  const stopElapsedTimer = useCallback(() => {
    if (elapsedTimerRef.current !== null) {
      window.clearInterval(elapsedTimerRef.current);
      elapsedTimerRef.current = null;
    }
  }, []);

  const updateSessionByID = useCallback(
    (sessionID: string, updater: (session: ChatSession) => ChatSession) => {
      setSessions((prev) => {
        const existing = prev.find((s) => s.id === sessionID);
        if (!existing) return prev;
        const updatedRaw = updater(existing);
        const updated: ChatSession = {
          ...updatedRaw,
          updatedAt: Date.now(),
          messages: updatedRaw.messages.slice(-AI_MAX_MESSAGES_PER_SESSION),
        };
        const rest = prev.filter((s) => s.id !== sessionID);
        return [updated, ...rest].slice(0, AI_MAX_SESSIONS);
      });
    },
    [],
  );

  const stopStreaming = useCallback(() => {
    if (!thinking) return;
    abortRef.current?.abort();
    abortRef.current = null;
    stopElapsedTimer();
    setThinking(false);
    setProgressText(t("aiProgressCanceled"));
    setProgressPct((prev) => (prev > 0 ? prev : 0));
  }, [stopElapsedTimer, t, thinking]);

  const deleteSession = useCallback(
    (id: string) => {
      if (thinking && id === activeSessionID) {
        stopStreaming();
      }
      setSessions((prev) => {
        const next = prev.filter((s) => s.id !== id);
        if (activeSessionID === id) {
          setActiveSessionID(next[0]?.id ?? "");
        }
        return next;
      });
      resetProgress();
    },
    [activeSessionID, resetProgress, stopStreaming, thinking],
  );

  const openSession = useCallback(
    (id: string) => {
      if (thinking) {
        stopStreaming();
      }
      setActiveSessionID(id);
      resetProgress();
    },
    [resetProgress, stopStreaming, thinking],
  );

  const startNewChat = useCallback(() => {
    if (thinking) {
      stopStreaming();
    }
    createSession(t("aiSessionUntitled"));
    resetProgress();
  }, [createSession, resetProgress, stopStreaming, t, thinking]);

  const toggleWikiTool = useCallback(() => {
    setNextPromptWiki((prev) => {
      const next = !prev;
      setCfg((cfgPrev) => ({
        ...cfgPrev,
        enableWikiContext: next,
      }));
      return next;
    });
  }, []);

  const toggleWebTool = useCallback(() => {
    setNextPromptWeb((prev) => {
      const next = !prev;
      setCfg((cfgPrev) => ({
        ...cfgPrev,
        enableWebResearch: next,
      }));
      return next;
    });
  }, []);

  const sendMessage = async (text: string) => {
    const content = text.trim();
    if (!content || thinking || disabled || sendInFlightRef.current) return;
    if (!cfg.apiKey.trim()) {
      setConfigOpen(true);
      addToast(t("aiErrorNoKey"), "error", 2800);
      return;
    }
    if (!effectiveModel) {
      addToast(t("aiErrorNoModel"), "error", 2800);
      return;
    }
    sendInFlightRef.current = true;
    const requestWikiContext = nextPromptWiki;
    const requestWebResearch = nextPromptWeb;
    setPromptToolsOpen(false);

    let sessionID = activeSession?.id || activeSessionID;
    if (!sessionID) {
      sessionID = createSession(t("aiSessionUntitled"));
    }
    const requestHistory = buildHistoryForRequest(activeSession?.messages ?? []);

    resetProgress();
    startElapsedTimer();
    setProgressText(t("aiProgressPreparing"));
    setProgressPct(6);
    setInput("");
    setThinking(true);

    const now = Date.now();
    const userMessageID = nextMessageID();
    const assistantMessageID = nextMessageID();
    const userMessage: ChatMessage = {
      id: userMessageID,
      role: "user",
      text: content,
      createdAt: now,
    };
    const assistantMessage: ChatMessage = {
      id: assistantMessageID,
      role: "assistant",
      text: "",
      createdAt: now + 1,
    };

    updateSessionByID(sessionID, (session) => ({
      ...session,
      title:
        session.messages.length === 0
          ? deriveSessionTitle(content, t("aiSessionUntitled"))
          : session.title,
      systemName: params.system_name || "",
      stationScope: selectedStationLabel,
      model: effectiveModel,
      messages: [...session.messages, userMessage, assistantMessage],
    }));
    setActiveSessionID(sessionID);

    const controller = new AbortController();
    abortRef.current = controller;

    try {
      const res = await stationAIChatStream(
        {
          provider: "openrouter",
          api_key: cfg.apiKey.trim(),
          model: effectiveModel,
          temperature: cfg.temperature,
          max_tokens: cfg.maxTokens,
          assistant_name: AI_ASSISTANT_NAME,
          locale,
          user_message: content,
          enable_wiki_context: requestWikiContext,
          enable_web_research: requestWebResearch,
          wiki_repo: cfg.wikiRepo.trim() || "https://github.com/ilyaux/Eve-flipper/wiki",
          history: requestHistory,
          context: buildContext(),
        },
        {
          onProgress: (msg) => {
            const pct = msg.progress_pct;
            if (typeof pct === "number") {
              setProgressPct((prev) => Math.max(prev, Math.min(100, pct)));
            }
            if (msg.message) {
              setProgressText(msg.message);
            }
            if (typeof msg.prompt_tokens_est === "number") {
              setPromptTokensEst(msg.prompt_tokens_est);
            }
            if (typeof msg.completion_tokens_est === "number") {
              setCompletionTokensEst(msg.completion_tokens_est);
            }
            if (typeof msg.total_tokens_est === "number") {
              setTotalTokensEst(msg.total_tokens_est);
            }
          },
          onDelta: (msg) => {
            const pct = msg.progress_pct;
            if (typeof pct === "number") {
              setProgressPct((prev) => Math.max(prev, Math.min(100, pct)));
            } else {
              setProgressPct((prev) => (prev < 90 ? prev + 1 : prev));
            }
            setProgressText(t("aiProgressStreaming"));
            if (typeof msg.completion_tokens_est === "number") {
              setCompletionTokensEst(msg.completion_tokens_est);
            }
            if (typeof msg.total_tokens_est === "number") {
              setTotalTokensEst(msg.total_tokens_est);
            }
            if (msg.delta) {
              updateSessionByID(sessionID, (session) => ({
                ...session,
                messages: session.messages.map((m) =>
                  m.id === assistantMessage.id
                    ? { ...m, text: `${m.text}${msg.delta}` }
                    : m,
                ),
              }));
            }
          },
          onUsage: (msg) => {
            setUsage({
              prompt_tokens: msg.prompt_tokens,
              completion_tokens: msg.completion_tokens,
              total_tokens: msg.total_tokens,
            });
            const pct = msg.progress_pct;
            if (typeof pct === "number") {
              setProgressPct((prev) => Math.max(prev, Math.min(100, pct)));
            }
          },
          onResult: (msg) => {
            if (msg.answer) {
              updateSessionByID(sessionID, (session) => ({
                ...session,
                messages: session.messages.map((m) =>
                  m.id === assistantMessage.id ? { ...m, text: msg.answer } : m,
                ),
              }));
            }
            if (msg.usage) {
              setUsage(msg.usage);
            }
            if (typeof msg.progress_pct === "number") {
              setProgressPct(Math.max(0, Math.min(100, msg.progress_pct)));
            }
            if (msg.progress_text) {
              setProgressText(msg.progress_text);
            }
          },
        },
        controller.signal,
      );

      const answer = (res.answer || "").trim();
      if (!answer) {
        throw new Error(t("aiErrorEmptyAnswer"));
      }
      const userWarning = firstUserFacingWarning(res.warnings);
      if (userWarning) {
        addToast(userWarning, "warning", 2800);
      }
      setProgressPct(100);
      setProgressText(t("aiProgressDone"));
    } catch (err: unknown) {
      if (err instanceof Error && err.name === "AbortError") {
        setProgressText(t("aiProgressCanceled"));
        return;
      }
      const msg = err instanceof Error ? err.message : t("aiErrorGeneric");
      addToast(msg, "error", 3200);
      updateSessionByID(sessionID, (session) => ({
        ...session,
        messages: session.messages.map((m) =>
          m.id === assistantMessage.id && !m.text.trim()
            ? { ...m, text: `${t("aiErrorPrefix")} ${msg}` }
            : m,
        ),
      }));
      setProgressText(`${t("aiErrorPrefix")} ${msg}`);
    } finally {
      abortRef.current = null;
      stopElapsedTimer();
      setThinking(false);
      sendInFlightRef.current = false;
    }
  };

  if (!open) {
    return (
      <button
        type="button"
        onClick={() => setOpen(true)}
        className="fixed right-4 bottom-4 z-[55] px-3 py-1.5 rounded-sm border border-eve-accent/60 bg-eve-dark/95 text-eve-accent hover:bg-eve-accent/10 hover:border-eve-accent transition-colors shadow-eve-glow text-xs uppercase tracking-wide leading-tight"
        title={t("aiOpen")}
      >
        <span className="block">{AI_ASSISTANT_NAME}</span>
        <span className="block text-[9px] text-eve-dim tracking-widest">{AI_RELEASE_STAGE}</span>
      </button>
    );
  }

  return (
    <>
      <section
        className="ai-chat-window fixed z-[55] rounded-sm border flex flex-col overflow-hidden"
        style={{
          left: `${windowRect.left}px`,
          top: `${windowRect.top}px`,
          width: `${windowRect.width}px`,
          height: `${windowRect.height}px`,
        }}
      >
        <header
          className="ai-chat-header shrink-0 px-3 py-2 border-b flex items-center gap-2 select-none cursor-move"
          onPointerDown={startWindowMove}
        >
          <div className="min-w-0 flex items-center gap-2">
            <IvyAIMascot thinking={thinking} />
            <div className="min-w-0">
              <h4 className="text-xs uppercase tracking-wider text-eve-accent font-semibold truncate">
                {AI_ASSISTANT_NAME}
              </h4>
              <p className="text-[10px] text-eve-dim truncate">{t("aiStationCopilot")}</p>
            </div>
          </div>
          <div className="flex-1" />
          <button
            type="button"
            onClick={() => setConfigOpen(true)}
            className="px-2 py-0.5 rounded-sm border border-eve-border/60 text-eve-dim hover:text-eve-accent hover:border-eve-accent/50 transition-colors text-xs"
            title={t("aiConfig")}
          >
            ⚙
          </button>
          {thinking && (
            <button
              type="button"
              onClick={stopStreaming}
              className="px-2 py-0.5 rounded-sm border border-amber-500/60 text-amber-300 hover:border-amber-400 hover:text-amber-200 transition-colors text-xs"
              title={t("aiStop")}
            >
              {t("aiStop")}
            </button>
          )}
          <button
            type="button"
            onClick={() => setOpen(false)}
            className="px-2 py-0.5 rounded-sm border border-eve-border/60 text-eve-dim hover:text-eve-accent hover:border-eve-accent/50 transition-colors text-xs"
            title={t("close")}
          >
            ×
          </button>
        </header>

        <div className="flex-1 min-h-0 flex flex-col sm:flex-row">
          <aside
            className={`ai-chat-sidebar shrink-0 border-b sm:border-b-0 sm:border-r flex flex-col min-h-[60px] sm:min-h-0 transition-[width] duration-200 ${
              sessionsCollapsed ? "sm:w-[56px]" : "sm:w-[248px]"
            }`}
          >
            <div className="px-2 py-2 border-b border-eve-border/40 flex items-center gap-1.5">
              {!sessionsCollapsed && (
                <span className="text-[11px] uppercase tracking-wider text-eve-dim flex-1 truncate">
                  {t("aiChats")}
                </span>
              )}
              <button
                type="button"
                onClick={() => setSessionsCollapsed((prev) => !prev)}
                className="h-6 px-1.5 rounded-sm border border-eve-border/60 text-eve-dim hover:text-eve-accent hover:border-eve-accent/50 transition-colors text-[10px]"
                title={sessionsCollapsed ? t("aiExpandChats") : t("aiCollapseChats")}
              >
                {sessionsCollapsed ? "»" : "«"}
              </button>
              <button
                type="button"
                onClick={startNewChat}
                disabled={disabled || thinking}
                className={`h-6 rounded-sm border border-eve-accent/60 text-eve-accent hover:bg-eve-accent/10 transition-colors text-[10px] disabled:opacity-40 disabled:cursor-not-allowed ${
                  sessionsCollapsed ? "w-6 px-0" : "px-2"
                }`}
                title={t("aiNewChat")}
              >
                {sessionsCollapsed ? "+" : `+ ${t("aiNewChat")}`}
              </button>
            </div>

            {sessionsCollapsed ? (
              <button
                type="button"
                onClick={() => setSessionsCollapsed(false)}
                className="m-2 px-1.5 py-2 rounded-sm border border-eve-border/50 bg-eve-dark/45 text-[10px] text-eve-dim hover:text-eve-accent hover:border-eve-accent/50 transition-colors"
                title={t("aiExpandChats")}
              >
                {sessions.length}
              </button>
            ) : (
              <div className="flex-1 min-h-0 overflow-y-auto eve-scrollbar p-2 space-y-1.5">
                {sessions.length === 0 && (
                  <div className="text-[11px] text-eve-dim px-2 py-1">{t("aiNoChats")}</div>
                )}
                {sessions.map((session) => {
                  const isActive = session.id === activeSession?.id;
                  return (
                    <div
                      key={session.id}
                      className={`rounded-sm border transition-colors ${
                        isActive
                          ? "border-eve-accent/70 bg-eve-accent/14"
                          : "border-eve-border/50 bg-eve-dark/28"
                      }`}
                    >
                      <div className="flex items-start gap-1.5">
                        <button
                          type="button"
                          onClick={() => openSession(session.id)}
                          disabled={disabled || thinking}
                          className="flex-1 min-w-0 text-left px-2 py-1.5 disabled:opacity-50"
                        >
                          <p className="text-[11px] text-eve-text truncate">
                            {session.title || t("aiSessionUntitled")}
                          </p>
                          <p className="text-[10px] text-eve-dim truncate">
                            {formatSessionTimestamp(session.updatedAt, locale)} • {session.messages.length}{" "}
                            {t("aiMessages")}
                          </p>
                        </button>
                        <button
                          type="button"
                          onClick={() => deleteSession(session.id)}
                          disabled={disabled || thinking}
                          className="mt-1.5 mr-1.5 px-1.5 py-0.5 rounded-sm border border-eve-border/60 text-eve-dim hover:text-red-300 hover:border-red-500/50 transition-colors text-[10px] disabled:opacity-40 disabled:cursor-not-allowed"
                          title={t("aiDeleteChat")}
                        >
                          ×
                        </button>
                      </div>
                    </div>
                  );
                })}
              </div>
            )}
          </aside>

          <div className="ai-chat-main flex-1 min-w-0 flex flex-col">
            <div className="ai-chat-context-strip shrink-0 px-3 py-1.5 border-b text-[10px] text-eve-dim flex flex-wrap items-center gap-1.5">
              <span>{params.system_name || "-"}</span>
              <span>•</span>
              <span>{selectedStationLabel}</span>
              <span>•</span>
              <span>{effectiveModel || "-"}</span>
              <span>•</span>
              <span>{nextPromptWiki ? t("aiWikiOn") : t("aiWikiOff")}</span>
              <span>•</span>
              <span>{nextPromptWeb ? t("aiWebOn") : t("aiWebOff")}</span>
            </div>

            <div className="ai-chat-progress-panel shrink-0 px-3 py-2 border-b space-y-1.5">
              <div className="flex items-center justify-between gap-2 text-[10px] text-eve-dim">
                <span className="truncate">{progressText || t("aiProgressIdle")}</span>
                <div className="flex items-center gap-2">
                  <span className="font-mono tabular-nums text-eve-text">{progressPct}%</span>
                  <button
                    type="button"
                    onClick={() => setProgressCollapsed((prev) => !prev)}
                    className="px-1.5 py-0.5 rounded-sm border border-eve-border/60 text-eve-dim hover:text-eve-accent hover:border-eve-accent/50 transition-colors text-[10px]"
                    title={progressCollapsed ? t("aiExpandProgress") : t("aiCollapseProgress")}
                    aria-label={progressCollapsed ? t("aiExpandProgress") : t("aiCollapseProgress")}
                  >
                    {progressCollapsed ? "▸" : "▾"}
                  </button>
                </div>
              </div>
              {!progressCollapsed && (
                <>
                  <div className="h-1.5 rounded-sm border border-eve-border/50 bg-eve-dark overflow-hidden">
                    <div
                      className="h-full bg-eve-accent transition-[width] duration-200 ease-out"
                      style={{ width: `${Math.max(0, Math.min(100, progressPct))}%` }}
                    />
                  </div>
                  <div className="grid grid-cols-2 gap-x-2 gap-y-1 text-[10px] text-eve-dim">
                    <span>
                      {t("aiTokensPromptEst")}: <span className="font-mono text-eve-text">{promptTokensEst}</span>
                    </span>
                    <span>
                      {t("aiTokensCompletionEst")}: <span className="font-mono text-eve-text">{completionTokensEst}</span>
                    </span>
                    <span>
                      {t("aiTokensTotalEst")}: <span className="font-mono text-eve-text">{totalTokensEst || promptTokensEst + completionTokensEst}</span>
                    </span>
                    <span>
                      {t("aiElapsed")}: <span className="font-mono text-eve-text">{elapsedSec}s</span>
                    </span>
                    {usage && (
                      <span className="col-span-2 text-eve-accent">
                        {t("aiTokensTotal")}: <span className="font-mono text-eve-accent">{usage.total_tokens}</span> ({t("aiTokensPrompt")} {usage.prompt_tokens} / {t("aiTokensCompletion")} {usage.completion_tokens})
                      </span>
                    )}
                  </div>
                </>
              )}
            </div>

            <div className="ai-chat-thread flex-1 min-h-0 p-3 overflow-y-auto eve-scrollbar space-y-2 select-text">
              {messages.length === 0 && (
                <div className="space-y-2">
                  <p className="text-xs text-eve-dim">{t("aiGreeting")}</p>
                  <div className="grid grid-cols-1 gap-1.5">
                    {quickPrompts.map((prompt) => (
                      <button
                        key={prompt}
                        type="button"
                        onClick={() => {
                          void sendMessage(prompt);
                        }}
                        disabled={disabled || thinking}
                        className="px-2 py-1 text-left rounded-sm border border-eve-border/60 bg-eve-dark/50 text-eve-text hover:border-eve-accent/50 hover:text-eve-accent transition-colors text-xs disabled:opacity-40 disabled:cursor-not-allowed"
                      >
                        {prompt}
                      </button>
                    ))}
                  </div>
                </div>
              )}

              {messages.map((msg) => (
                <div
                  key={msg.id}
                  className={`ai-chat-bubble max-w-[95%] rounded-sm border px-2 py-1.5 text-xs select-text cursor-text ${
                    msg.role === "assistant"
                      ? "ai-chat-bubble--assistant text-eve-text"
                      : "ai-chat-bubble--user ml-auto text-eve-text whitespace-pre-wrap"
                  }`}
                >
                  {msg.role === "assistant" ? <MarkdownMessage text={msg.text} /> : msg.text}
                </div>
              ))}

              {thinking && (
                <div className="ai-chat-bubble ai-chat-bubble--assistant max-w-[95%] rounded-sm border px-2 py-1.5 text-xs text-eve-dim">
                  {t("aiThinking")}
                </div>
              )}
              <div ref={endRef} />
            </div>

            <footer className="ai-chat-footer shrink-0 p-3 border-t">
              <div className="ai-chat-composer-shell relative rounded-sm border p-2">
                <div className="flex items-end gap-2">
                  <textarea
                    value={input}
                    onChange={(e) => setInput(e.target.value)}
                    onKeyDown={(e) => {
                      if (e.key === "Escape") {
                        setPromptToolsOpen(false);
                      }
                      if (e.key === "Enter" && !e.shiftKey) {
                        e.preventDefault();
                        void sendMessage(input);
                      }
                    }}
                    placeholder={t("aiInputPlaceholder")}
                    className="ai-chat-input flex-1 min-h-[70px] max-h-[180px] resize-y rounded-sm border px-3 py-2 text-xs text-eve-text placeholder:text-eve-dim/70 focus:outline-none"
                  />
                  <div ref={promptToolsRef} className="relative flex flex-col gap-1.5">
                    <button
                      type="button"
                      onClick={() => setPromptToolsOpen((prev) => !prev)}
                      disabled={disabled || thinking}
                      className={`h-8 px-2 rounded-sm border text-[10px] uppercase tracking-wide transition-colors disabled:opacity-40 disabled:cursor-not-allowed ${
                        promptToolsOpen
                          ? "border-eve-accent/80 text-eve-accent bg-eve-accent/10"
                          : "border-eve-border/70 text-eve-dim hover:text-eve-accent hover:border-eve-accent/50"
                      }`}
                      title={t("aiPromptTools")}
                    >
                      + {t("aiTools")}
                    </button>
                    <button
                      type="button"
                      onClick={() => {
                        void sendMessage(input);
                      }}
                      disabled={thinking || disabled || !input.trim()}
                      className="h-10 px-4 rounded-sm border border-eve-accent/70 bg-eve-accent/15 text-eve-accent hover:bg-eve-accent/25 transition-colors text-xs font-semibold disabled:opacity-40 disabled:cursor-not-allowed"
                    >
                      {t("aiSend")}
                    </button>

                    {promptToolsOpen && (
                      <div className="ai-chat-tools-popover absolute right-0 bottom-[calc(100%+8px)] w-[min(300px,90vw)] rounded-sm border p-2 z-[56]">
                        <div className="flex items-center justify-between gap-2 mb-2">
                          <p className="text-[10px] uppercase tracking-wider text-eve-accent">
                            {t("aiPromptTools")}
                          </p>
                          <button
                            type="button"
                            onClick={() => setPromptToolsOpen(false)}
                            className="px-1.5 py-0.5 rounded-sm border border-eve-border/60 text-eve-dim hover:text-eve-accent hover:border-eve-accent/50 transition-colors text-[10px]"
                          >
                            ×
                          </button>
                        </div>
                        <div className="space-y-2">
                          <button
                            type="button"
                            onClick={toggleWikiTool}
                            className={`w-full flex items-center justify-between gap-2 px-2 py-1.5 rounded-sm border text-xs transition-colors ${
                              nextPromptWiki
                                ? "border-eve-accent/70 bg-eve-accent/10 text-eve-text"
                                : "border-eve-border/60 bg-eve-dark/50 text-eve-dim"
                            }`}
                          >
                            <span>{t("aiEnableWikiContext")}</span>
                            <span className="font-mono">
                              {nextPromptWiki ? t("aiWikiOn") : t("aiWikiOff")}
                            </span>
                          </button>
                          <button
                            type="button"
                            onClick={toggleWebTool}
                            className={`w-full flex items-center justify-between gap-2 px-2 py-1.5 rounded-sm border text-xs transition-colors ${
                              nextPromptWeb
                                ? "border-eve-accent/70 bg-eve-accent/10 text-eve-text"
                                : "border-eve-border/60 bg-eve-dark/50 text-eve-dim"
                            }`}
                          >
                            <span>{t("aiEnableWebResearch")}</span>
                            <span className="font-mono">
                              {nextPromptWeb ? t("aiWebOn") : t("aiWebOff")}
                            </span>
                          </button>
                        </div>
                      </div>
                    )}
                  </div>
                </div>

                <div className="mt-1.5 flex items-center justify-between gap-2 text-[10px] text-eve-dim">
                  <span>{t("aiInputHint")}</span>
                  <span className="font-mono">
                    {nextPromptWiki ? t("aiWikiOn") : t("aiWikiOff")} |{" "}
                    {nextPromptWeb ? t("aiWebOn") : t("aiWebOff")}
                  </span>
                </div>

              </div>
            </footer>
          </div>
        </div>

        <div
          aria-hidden="true"
          onPointerDown={startWindowResize("e")}
          className="absolute right-0 top-0 h-full w-1.5 z-[57] cursor-ew-resize"
        />
        <div
          aria-hidden="true"
          onPointerDown={startWindowResize("s")}
          className="absolute left-0 bottom-0 w-full h-1.5 z-[57] cursor-ns-resize"
        />
        <div
          aria-hidden="true"
          onPointerDown={startWindowResize("se")}
          className="absolute right-0 bottom-0 w-3 h-3 z-[58] cursor-nwse-resize"
        />
        <div
          aria-hidden="true"
          className="pointer-events-none absolute right-1 bottom-1 text-[9px] leading-none text-eve-dim/60 select-none"
        >
          ◢
        </div>
      </section>

      {configOpen && (
        <>
          <div className="fixed inset-0 z-[59] bg-black/70" onClick={() => setConfigOpen(false)} />
          <div className="fixed z-[60] left-1/2 top-1/2 -translate-x-1/2 -translate-y-1/2 w-[min(560px,92vw)] rounded-sm border border-eve-border bg-eve-panel p-3 shadow-eve-glow-strong">
            <div className="flex items-center justify-between">
              <h3 className="text-sm uppercase tracking-wider text-eve-text font-semibold">{t("aiConfig")}</h3>
              <button
                type="button"
                onClick={() => setConfigOpen(false)}
                className="px-2 py-0.5 rounded-sm border border-eve-border/60 text-eve-dim hover:text-eve-accent hover:border-eve-accent/50 transition-colors text-xs"
              >
                {t("close")}
              </button>
            </div>

            <div className="mt-3 grid grid-cols-1 sm:grid-cols-2 gap-3">
              <label className="text-xs text-eve-dim">
                <span className="block mb-1">{t("aiProvider")}</span>
                <input
                  value="OpenRouter"
                  disabled
                  className="w-full h-8 rounded-sm border border-eve-border bg-eve-dark px-2 text-eve-dim"
                />
              </label>

              <label className="text-xs text-eve-dim sm:col-span-2">
                <span className="block mb-1">{t("aiApiKey")}</span>
                <input
                  type="password"
                  value={cfg.apiKey}
                  onChange={(e) =>
                    setCfg((prev) => ({ ...prev, apiKey: e.target.value }))
                  }
                  placeholder="sk-or-..."
                  className="w-full h-8 rounded-sm border border-eve-border bg-eve-input px-2 text-eve-text"
                />
              </label>

              <label className="text-xs text-eve-dim">
                <span className="block mb-1">{t("aiModel")}</span>
                <select
                  value={cfg.model}
                  onChange={(e) =>
                    setCfg((prev) => ({ ...prev, model: e.target.value }))
                  }
                  className="w-full h-8 rounded-sm border border-eve-border bg-eve-input px-2 text-eve-text"
                >
                  {OPENROUTER_MODELS.map((m) => (
                    <option key={m} value={m}>
                      {m}
                    </option>
                  ))}
                </select>
              </label>

              <label className="text-xs text-eve-dim">
                <span className="block mb-1">{t("aiCustomModel")}</span>
                <input
                  value={cfg.customModel}
                  onChange={(e) =>
                    setCfg((prev) => ({ ...prev, customModel: e.target.value }))
                  }
                  placeholder="provider/model"
                  className="w-full h-8 rounded-sm border border-eve-border bg-eve-input px-2 text-eve-text"
                />
              </label>

              <label className="inline-flex items-center gap-2 text-xs text-eve-dim">
                <input
                  type="checkbox"
                  checked={cfg.useCustomModel}
                  onChange={(e) =>
                    setCfg((prev) => ({
                      ...prev,
                      useCustomModel: e.target.checked,
                    }))
                  }
                  className="accent-eve-accent"
                />
                <span>{t("aiUseCustomModel")}</span>
              </label>

              <label className="text-xs text-eve-dim">
                <span className="block mb-1">{t("aiTemperature")}</span>
                <input
                  type="number"
                  min={0}
                  max={2}
                  step={0.1}
                  value={cfg.temperature}
                  onChange={(e) =>
                    setCfg((prev) => ({
                      ...prev,
                      temperature: Math.max(0, Math.min(2, Number(e.target.value) || 0)),
                    }))
                  }
                  className="w-full h-8 rounded-sm border border-eve-border bg-eve-input px-2 text-eve-text"
                />
              </label>

              <label className="text-xs text-eve-dim">
                <span className="block mb-1">{t("aiMaxTokens")}</span>
                <input
                  type="number"
                  min={200}
                  max={AI_MAX_TOKENS_LIMIT}
                  step={50}
                  value={cfg.maxTokens}
                  onChange={(e) =>
                    setCfg((prev) => ({
                      ...prev,
                      maxTokens: Math.max(
                        200,
                        Math.min(AI_MAX_TOKENS_LIMIT, Number(e.target.value) || 200),
                      ),
                    }))
                  }
                  className="w-full h-8 rounded-sm border border-eve-border bg-eve-input px-2 text-eve-text"
                />
              </label>

              <label className="inline-flex items-center gap-2 text-xs text-eve-dim sm:col-span-2">
                <input
                  type="checkbox"
                  checked={cfg.enableWikiContext}
                  onChange={(e) =>
                    setCfg((prev) => ({
                      ...prev,
                      enableWikiContext: e.target.checked,
                    }))
                  }
                  className="accent-eve-accent"
                />
                <span>{t("aiEnableWikiContext")}</span>
              </label>

              <label className="inline-flex items-center gap-2 text-xs text-eve-dim sm:col-span-2">
                <input
                  type="checkbox"
                  checked={cfg.enableWebResearch}
                  onChange={(e) =>
                    setCfg((prev) => ({
                      ...prev,
                      enableWebResearch: e.target.checked,
                    }))
                  }
                  className="accent-eve-accent"
                />
                <span>{t("aiEnableWebResearch")}</span>
              </label>

              <label className="text-xs text-eve-dim sm:col-span-2">
                <span className="block mb-1">{t("aiWikiRepo")}</span>
                <input
                  value={cfg.wikiRepo}
                  onChange={(e) =>
                    setCfg((prev) => ({
                      ...prev,
                      wikiRepo: e.target.value,
                    }))
                  }
                  placeholder="https://github.com/owner/repo/wiki or owner/repo"
                  className="w-full h-8 rounded-sm border border-eve-border bg-eve-input px-2 text-eve-text"
                />
              </label>
            </div>

            <p className="mt-3 text-[11px] text-eve-dim">{t("aiConfigHint")}</p>
          </div>
        </>
      )}
    </>
  );
}
