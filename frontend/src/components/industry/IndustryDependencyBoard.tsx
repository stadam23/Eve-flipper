import { useEffect, useMemo, useState } from "react";
import { formatDuration, industryTaskStatusClass, type IndustryTaskDependencyBoard } from "./industryHelpers";

interface Props {
  board: IndustryTaskDependencyBoard;
}

export function IndustryDependencyBoard({ board }: Props) {
  const [rowsPerPage, setRowsPerPage] = useState(80);
  const [page, setPage] = useState(1);

  const totalPages = useMemo(
    () => Math.max(1, Math.ceil(board.rows.length / Math.max(1, rowsPerPage))),
    [board.rows.length, rowsPerPage]
  );

  useEffect(() => {
    setPage((prev) => Math.min(Math.max(1, prev), totalPages));
  }, [totalPages]);

  const visibleRows = useMemo(() => {
    const safeSize = Math.max(1, rowsPerPage);
    const start = (page - 1) * safeSize;
    return board.rows.slice(start, start + safeSize);
  }, [board.rows, page, rowsPerPage]);

  return (
    <div className="mt-2 border border-eve-border/40 rounded-sm p-2 bg-eve-dark/20">
      <div className="flex items-center justify-between gap-2 mb-1">
        <div className="text-[10px] uppercase tracking-wider text-eve-dim">Dependency Board ({board.total_edges} links)</div>
        <div className="inline-flex items-center gap-1 text-[10px] text-eve-dim">
          <span>critical path: {formatDuration(board.critical_path_sec)}</span>
          <span>rows/page</span>
          <select
            value={rowsPerPage}
            onChange={(e) => setRowsPerPage(Math.max(20, Math.min(500, Number(e.target.value) || 80)))}
            className="px-1 py-0.5 bg-eve-input border border-eve-border rounded-sm text-[10px] text-eve-text"
          >
            <option value={40}>40</option>
            <option value={80}>80</option>
            <option value={160}>160</option>
            <option value={320}>320</option>
          </select>
          <span>{page}/{totalPages}</span>
          <button
            type="button"
            onClick={() => setPage((prev) => Math.max(1, prev - 1))}
            disabled={page <= 1}
            className="px-1 border border-eve-border rounded-sm disabled:opacity-40"
          >
            {"<"}
          </button>
          <button
            type="button"
            onClick={() => setPage((prev) => Math.min(totalPages, prev + 1))}
            disabled={page >= totalPages}
            className="px-1 border border-eve-border rounded-sm disabled:opacity-40"
          >
            {">"}
          </button>
        </div>
      </div>
      <div className="grid grid-cols-2 md:grid-cols-4 lg:grid-cols-8 gap-1 text-[11px] mb-2">
        <div className="px-1.5 py-1 rounded-sm border border-eve-border/40 bg-eve-dark/30 text-eve-dim">
          tasks <span className="float-right font-mono text-eve-text">{board.total_tasks}</span>
        </div>
        <div className="px-1.5 py-1 rounded-sm border border-eve-border/40 bg-eve-dark/30 text-eve-dim">
          links <span className="float-right font-mono text-eve-accent">{board.total_edges}</span>
        </div>
        <div className="px-1.5 py-1 rounded-sm border border-eve-border/40 bg-eve-dark/30 text-eve-dim">
          roots <span className="float-right font-mono text-cyan-300">{board.roots}</span>
        </div>
        <div className="px-1.5 py-1 rounded-sm border border-eve-border/40 bg-eve-dark/30 text-eve-dim">
          leaves <span className="float-right font-mono text-emerald-300">{board.leaves}</span>
        </div>
        <div className="px-1.5 py-1 rounded-sm border border-eve-border/40 bg-eve-dark/30 text-eve-dim">
          depth <span className="float-right font-mono text-fuchsia-300">{board.max_depth}</span>
        </div>
        <div className="px-1.5 py-1 rounded-sm border border-eve-border/40 bg-eve-dark/30 text-eve-dim">
          orphans <span className="float-right font-mono text-yellow-300">{board.orphans}</span>
        </div>
        <div className="px-1.5 py-1 rounded-sm border border-eve-border/40 bg-eve-dark/30 text-eve-dim">
          self-links <span className="float-right font-mono text-amber-300">{board.self_links}</span>
        </div>
        <div className="px-1.5 py-1 rounded-sm border border-eve-border/40 bg-eve-dark/30 text-eve-dim">
          cycles <span className="float-right font-mono text-red-300">{board.cycles}</span>
        </div>
      </div>
      <div className="border border-eve-border rounded-sm max-h-[170px] overflow-auto">
        <table className="w-full text-[11px]">
          <thead className="sticky top-0 bg-eve-dark z-10">
            <tr className="text-eve-dim uppercase tracking-wider border-b border-eve-border/60">
              <th className="px-1.5 py-1 text-left">Child</th>
              <th className="px-1.5 py-1 text-left">Child status</th>
              <th className="px-1.5 py-1 text-left">Parent</th>
              <th className="px-1.5 py-1 text-left">Parent status</th>
            </tr>
          </thead>
          <tbody>
            {visibleRows.map((row) => (
              <tr key={`dep-row-${row.child_id}-${row.parent_id}`} className="border-b border-eve-border/30">
                <td className="px-1.5 py-1 text-eve-text">
                  <div className="truncate">{row.child_name}</div>
                  <div className="text-[10px] text-eve-dim">#{row.child_id}</div>
                </td>
                <td className="px-1.5 py-1">
                  <span className={`px-1.5 py-0.5 text-[10px] uppercase rounded-sm border ${industryTaskStatusClass(row.child_status)}`}>
                    {row.child_status}
                  </span>
                </td>
                <td className="px-1.5 py-1 text-eve-text">
                  <div className={`truncate ${row.parent_missing ? "text-yellow-300" : ""}`}>{row.parent_name}</div>
                  <div className="text-[10px] text-eve-dim">#{row.parent_id}</div>
                </td>
                <td className="px-1.5 py-1">
                  {row.parent_missing ? (
                    <span className="px-1.5 py-0.5 text-[10px] uppercase rounded-sm border border-yellow-500/40 text-yellow-300 bg-yellow-500/10">
                      missing
                    </span>
                  ) : (
                    <span className={`px-1.5 py-0.5 text-[10px] uppercase rounded-sm border ${industryTaskStatusClass(row.parent_status)}`}>
                      {row.parent_status}
                    </span>
                  )}
                </td>
              </tr>
            ))}
            {board.rows.length === 0 && (
              <tr>
                <td colSpan={4} className="px-2 py-2 text-center text-eve-dim">No task dependencies in snapshot.</td>
              </tr>
            )}
          </tbody>
        </table>
      </div>
    </div>
  );
}
