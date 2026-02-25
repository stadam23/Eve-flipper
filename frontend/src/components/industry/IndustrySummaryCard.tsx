interface IndustrySummaryCardProps {
  label: string;
  value: string;
  subtext?: string;
  color?: string;
}

export function IndustrySummaryCard({
  label,
  value,
  subtext,
  color = "text-eve-accent",
}: IndustrySummaryCardProps) {
  return (
    <div className="bg-eve-panel border border-eve-border rounded-sm p-3">
      <div className="text-[10px] uppercase tracking-wider text-eve-dim mb-1">{label}</div>
      <div className={`text-lg font-mono font-semibold ${color}`}>{value}</div>
      {subtext && <div className="text-xs text-eve-dim">{subtext}</div>}
    </div>
  );
}
