import { useState } from "react";
import { formatISK } from "@/lib/format";
import type { MaterialNode } from "@/lib/types";

interface IndustryMaterialTreeProps {
  node: MaterialNode;
}

export function IndustryMaterialTree({ node }: IndustryMaterialTreeProps) {
  return (
    <div className="p-2">
      <IndustryTreeNode node={node} />
    </div>
  );
}

function IndustryTreeNode({ node, level = 0 }: { node: MaterialNode; level?: number }) {
  const [expanded, setExpanded] = useState(level < 2);
  const hasChildren = node.children && node.children.length > 0;
  const indent = level * 20;
  const decisionDelta = node.buy_price - node.build_cost;
  const decisionHint = !node.is_base
    ? node.should_build
      ? node.buy_price > 0
        ? `Build wins by ${formatISK(Math.max(0, decisionDelta))} (job: ${formatISK(node.job_cost || 0)})`
        : `Build selected (no market buy price, job: ${formatISK(node.job_cost || 0)})`
      : `Buy wins by ${formatISK(Math.max(0, -decisionDelta))}`
    : "";

  return (
    <div>
      <div
        className={`flex items-center py-1 px-2 hover:bg-eve-accent/5 rounded-sm ${
          node.should_build ? "" : "opacity-70"
        }`}
        style={{ paddingLeft: Math.min(indent + 8, 120) }}
      >
        {hasChildren ? (
          <button
            onClick={() => setExpanded(!expanded)}
            className="w-4 h-4 flex items-center justify-center text-eve-dim hover:text-eve-accent mr-1"
          >
            {expanded ? "▼" : "▶"}
          </button>
        ) : (
          <span className="w-4 h-4 mr-1" />
        )}

        <span className="flex-1 text-sm text-eve-text truncate">
          {node.type_name}
          <span className="text-eve-dim ml-2">x{node.quantity.toLocaleString()}</span>
        </span>

        <span className="text-xs text-eve-dim mx-2">
          Buy: {formatISK(node.buy_price)}
        </span>
        {!node.is_base && (
          <span className="text-xs text-eve-dim mx-2">
            Build: {formatISK(node.build_cost)}
          </span>
        )}

        {!node.is_base && (
          <span
            className={`text-[10px] px-2 py-0.5 rounded-sm ${
              node.should_build
                ? "bg-green-500/20 text-green-400"
                : "bg-blue-500/20 text-blue-400"
            }`}
            title={decisionHint}
          >
            {node.should_build ? "BUILD" : "BUY"}
          </span>
        )}
        {node.is_base && (
          <span className="text-[10px] px-2 py-0.5 rounded-sm bg-eve-dim/20 text-eve-dim">
            BASE
          </span>
        )}
        {decisionHint && (
          <span className="text-[10px] text-eve-dim ml-2 hidden xl:inline">
            {decisionHint}
          </span>
        )}
      </div>

      {expanded && hasChildren && (
        <div>
          {node.children!.map((child, index) => (
            <IndustryTreeNode key={`${child.type_id}-${index}`} node={child} level={level + 1} />
          ))}
        </div>
      )}
    </div>
  );
}
