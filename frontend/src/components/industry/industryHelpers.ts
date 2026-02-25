import type { IndustryPlanPatch } from "@/lib/types";

export type IndustryPlannerWarningSource = "preview" | "apply" | "gate";

export interface IndustryPlannerWarningEvent {
  id: number;
  source: IndustryPlannerWarningSource;
  message: string;
  created_at: string;
}

export interface IndustryTaskDependencyRow {
  child_id: number;
  child_name: string;
  child_status: string;
  parent_id: number;
  parent_name: string;
  parent_status: string;
  parent_missing: boolean;
}

export interface IndustryTaskDependencyBoard {
  total_tasks: number;
  total_edges: number;
  roots: number;
  leaves: number;
  max_depth: number;
  critical_path_sec: number;
  orphans: number;
  cycles: number;
  self_links: number;
  depth_by_task: Record<number, number>;
  parent_by_task: Record<number, number>;
  parent_missing_by_task: Record<number, boolean>;
  critical_task_ids: Set<number>;
  rows: IndustryTaskDependencyRow[];
}

export function formatDuration(seconds: number): string {
  if (seconds <= 0) return "—";
  const d = Math.floor(seconds / 86400);
  const h = Math.floor((seconds % 86400) / 3600);
  const m = Math.floor((seconds % 3600) / 60);
  const s = seconds % 60;
  const parts: string[] = [];
  if (d > 0) parts.push(`${d}d`);
  if (h > 0) parts.push(`${h}h`);
  if (m > 0) parts.push(`${m}m`);
  if (parts.length === 0) parts.push(`${s}s`);
  return parts.join(" ");
}

export function formatUtcShort(value: string): string {
  const trimmed = value?.trim();
  if (!trimmed) return "—";
  const date = new Date(trimmed);
  if (Number.isNaN(date.getTime())) return trimmed;
  const yyyy = date.getUTCFullYear();
  const mm = String(date.getUTCMonth() + 1).padStart(2, "0");
  const dd = String(date.getUTCDate()).padStart(2, "0");
  const hh = String(date.getUTCHours()).padStart(2, "0");
  const mi = String(date.getUTCMinutes()).padStart(2, "0");
  return `${yyyy}-${mm}-${dd} ${hh}:${mi} UTC`;
}

export function industryJobStatusClass(status: string): string {
  switch (status) {
    case "completed":
      return "bg-emerald-500/20 text-emerald-400 border-emerald-500/30";
    case "active":
      return "bg-blue-500/20 text-blue-400 border-blue-500/30";
    case "queued":
    case "planned":
      return "bg-amber-500/20 text-amber-400 border-amber-500/30";
    case "paused":
      return "bg-indigo-500/20 text-indigo-400 border-indigo-500/30";
    case "failed":
      return "bg-red-500/20 text-red-400 border-red-500/30";
    case "cancelled":
      return "bg-eve-dim/20 text-eve-dim border-eve-dim/30";
    default:
      return "bg-eve-dim/20 text-eve-dim border-eve-dim/30";
  }
}

export function industryTaskStatusClass(status: string): string {
  switch (status) {
    case "completed":
      return "bg-emerald-500/20 text-emerald-400 border-emerald-500/30";
    case "active":
      return "bg-blue-500/20 text-blue-400 border-blue-500/30";
    case "ready":
      return "bg-amber-500/20 text-amber-400 border-amber-500/30";
    case "planned":
      return "bg-eve-dim/20 text-eve-dim border-eve-dim/30";
    case "blocked":
      return "bg-red-500/20 text-red-400 border-red-500/30";
    case "paused":
      return "bg-indigo-500/20 text-indigo-400 border-indigo-500/30";
    case "cancelled":
      return "bg-zinc-500/20 text-zinc-300 border-zinc-500/30";
    default:
      return "bg-eve-dim/20 text-eve-dim border-eve-dim/30";
  }
}

function stableSerialize(value: unknown): string {
  if (value === null || value === undefined) return "null";
  if (typeof value !== "object") return JSON.stringify(value);
  if (Array.isArray(value)) {
    return `[${value.map((item) => stableSerialize(item)).join(",")}]`;
  }
  const record = value as Record<string, unknown>;
  const keys = Object.keys(record).sort();
  return `{${keys.map((key) => `${JSON.stringify(key)}:${stableSerialize(record[key])}`).join(",")}}`;
}

export function planPatchSignature(patch: IndustryPlanPatch | null): string {
  if (!patch) return "";
  return stableSerialize(patch);
}

export function taskConstraintRecord(value: unknown): Record<string, unknown> {
  if (!value || typeof value !== "object" || Array.isArray(value)) {
    return {};
  }
  return { ...(value as Record<string, unknown>) };
}

export function taskConstraintNumber(value: unknown, key: string): number {
  const constraints = taskConstraintRecord(value);
  const raw = constraints[key];
  if (typeof raw === "number" && Number.isFinite(raw)) return raw;
  if (typeof raw === "string") {
    const parsed = Number(raw.trim());
    if (Number.isFinite(parsed)) return parsed;
  }
  return 0;
}

export function industryPlannerWarningSourceClass(source: IndustryPlannerWarningSource): string {
  switch (source) {
    case "preview":
      return "border-yellow-500/40 text-yellow-300 bg-yellow-500/10";
    case "apply":
      return "border-amber-500/40 text-amber-300 bg-amber-500/10";
    case "gate":
      return "border-red-500/40 text-red-300 bg-red-500/10";
    default:
      return "border-eve-border text-eve-dim bg-eve-dark/30";
  }
}
