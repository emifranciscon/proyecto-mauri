import { useState } from "react";
import { historialDescripcionPlain } from "../utils/pdfTableExport";
import type { HistorialRow } from "../api/client";

type Props = { row: HistorialRow };

/** ~2 lines clamped; optional expand for long descriptions. */
export function HistorialDescripcionCell({ row }: Props) {
  const text = historialDescripcionPlain(row);
  const [open, setOpen] = useState(false);
  if (text === "—") {
    return <span style={{ color: "var(--muted)" }}>—</span>;
  }
  return (
    <div className="historial-desc-wrap">
      <div className={open ? "historial-desc historial-desc--open" : "historial-desc"}>
        {text}
      </div>
      {text.length > 120 && (
        <button
          type="button"
          className="historial-desc-more"
          onClick={() => setOpen((v) => !v)}
        >
          {open ? "Ver menos" : "Ver más"}
        </button>
      )}
    </div>
  );
}
