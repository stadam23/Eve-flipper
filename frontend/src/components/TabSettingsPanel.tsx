import { useState, type ReactNode } from "react";
import { TabHelp } from "./TabHelp";

interface TabSettingsPanelProps {
  title: string;
  hint?: string;
  icon?: string;
  defaultExpanded?: boolean;
  /** Optional help: step keys and wiki slug for ? button */
  help?: { stepKeys: string[]; wikiSlug: string };
  children: ReactNode;
}

/**
 * Unified collapsible settings panel for tab-specific parameters.
 * Used to maintain consistent UI across all tabs.
 */
export function TabSettingsPanel({
  title,
  hint,
  icon = "⚙",
  defaultExpanded = false,
  help,
  children,
}: TabSettingsPanelProps) {
  const [expanded, setExpanded] = useState(defaultExpanded);

  return (
    <div className="bg-eve-panel border border-eve-border rounded-sm">
      <button
        onClick={() => setExpanded(!expanded)}
        className="w-full px-3 py-2 flex items-center justify-between text-left hover:bg-eve-accent/5 transition-colors"
      >
        <div className="flex items-center gap-2">
          <span className="text-eve-accent text-sm">{icon}</span>
          <span className="text-sm font-medium text-eve-text">{title}</span>
          {hint && <span className="text-xs text-eve-dim">— {hint}</span>}
        </div>
        <div className="flex items-center gap-2" onClick={(e) => e.stopPropagation()}>
          {help && <TabHelp stepKeys={help.stepKeys} wikiSlug={help.wikiSlug} />}
          <span className="text-eve-dim text-xs">{expanded ? "▲" : "▼"}</span>
        </div>
      </button>

      {expanded && (
        <div className="px-3 pb-3 border-t border-eve-border/50">
          <div className="pt-3">{children}</div>
        </div>
      )}
    </div>
  );
}

// --- Reusable input components for settings ---

interface FieldProps {
  label: string;
  children: ReactNode;
}

export function SettingsField({ label, children }: FieldProps) {
  return (
    <div className="flex flex-col gap-1">
      <label className="text-[11px] uppercase tracking-wider text-eve-dim font-medium">
        {label}
      </label>
      {children}
    </div>
  );
}

interface NumberInputProps {
  value: number;
  onChange: (v: number) => void;
  min?: number;
  max?: number;
  step?: number;
  placeholder?: string;
}

export function SettingsNumberInput({
  value,
  onChange,
  min = 0,
  max = 999999999,
  step = 1,
  placeholder,
}: NumberInputProps) {
  return (
    <input
      type="number"
      value={value}
      onChange={(e) => {
        const v = parseFloat(e.target.value);
        if (!isNaN(v) && v >= min && v <= max) onChange(v);
      }}
      min={min}
      max={max}
      step={step}
      placeholder={placeholder}
      className="w-full px-3 py-1.5 bg-eve-input border border-eve-border rounded-sm text-eve-text text-sm font-mono
                 focus:outline-none focus:border-eve-accent focus:ring-1 focus:ring-eve-accent/30
                 transition-colors
                 [appearance:textfield] [&::-webkit-outer-spin-button]:appearance-none [&::-webkit-inner-spin-button]:appearance-none"
    />
  );
}

interface SelectInputProps {
  value: string | number;
  onChange: (v: string) => void;
  options: { value: string | number; label: string }[];
}

export function SettingsSelect({ value, onChange, options }: SelectInputProps) {
  return (
    <select
      value={value}
      onChange={(e) => onChange(e.target.value)}
      className="w-full px-3 py-1.5 bg-eve-input border border-eve-border rounded-sm text-eve-text text-sm font-mono
                 focus:outline-none focus:border-eve-accent focus:ring-1 focus:ring-eve-accent/30
                 transition-colors"
    >
      {options.map((opt) => (
        <option key={opt.value} value={opt.value}>
          {opt.label}
        </option>
      ))}
    </select>
  );
}

interface CheckboxInputProps {
  checked: boolean;
  onChange: (v: boolean) => void;
  label?: string;
}

export function SettingsCheckbox({ checked, onChange, label }: CheckboxInputProps) {
  return (
    <div className="flex items-center h-[34px]">
      <label className="relative inline-flex items-center cursor-pointer gap-2">
        <input
          type="checkbox"
          checked={checked}
          onChange={(e) => onChange(e.target.checked)}
          className="sr-only peer"
        />
        <div className="w-9 h-5 bg-eve-input border border-eve-border rounded-full peer 
                      peer-checked:bg-eve-accent/30 peer-checked:border-eve-accent
                      after:content-[''] after:absolute after:top-[2px] after:left-[2px]
                      after:bg-eve-dim after:rounded-full after:h-4 after:w-4
                      after:transition-all peer-checked:after:translate-x-4 peer-checked:after:bg-eve-accent" />
        {label && <span className="text-xs text-eve-dim">{label}</span>}
      </label>
    </div>
  );
}

// Grid wrapper for settings fields
interface SettingsGridProps {
  children: ReactNode;
  cols?: number;
}

export function SettingsGrid({ children, cols = 4 }: SettingsGridProps) {
  const colsClass = {
    2: "grid-cols-2",
    3: "grid-cols-2 sm:grid-cols-3",
    4: "grid-cols-2 sm:grid-cols-3 md:grid-cols-4",
    5: "grid-cols-2 sm:grid-cols-3 md:grid-cols-4 lg:grid-cols-5",
  }[cols] ?? "grid-cols-2 sm:grid-cols-4";

  return (
    <div className={`grid ${colsClass} gap-x-3 gap-y-3 items-end`}>
      {children}
    </div>
  );
}

// Hints section at the bottom of settings
interface SettingsHintsProps {
  hints: string[];
}

export function SettingsHints({ hints }: SettingsHintsProps) {
  if (hints.length === 0) return null;
  return (
    <div className="mt-3 pt-2 border-t border-eve-border/30 text-xs text-eve-dim space-y-1">
      {hints.map((hint, i) => (
        <p key={i}>• {hint}</p>
      ))}
    </div>
  );
}
