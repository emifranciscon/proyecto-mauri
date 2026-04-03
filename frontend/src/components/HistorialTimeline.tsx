import { Link } from "react-router-dom";
import type { HistorialRow } from "../api/client";
import { HistorialDescripcionCell } from "./HistorialDescripcionCell";
import {
  balanzasResumenDisplay,
  historialOperacionLabel,
  tanquesResumenDisplay,
} from "../utils/pdfTableExport";

function formatHistorialFecha(s: string): string {
  const d = new Date(s);
  if (Number.isNaN(d.getTime())) return s?.slice(0, 10) ?? "—";
  return d.toLocaleString(undefined, { dateStyle: "medium", timeStyle: "short" });
}

type OperacionVisual = "ingreso" | "egreso" | "mantenimiento";

function operacionVisual(h: HistorialRow): OperacionVisual {
  const t = h.tipo_operacion?.trim().toLowerCase();
  if (t === "egreso") return "egreso";
  if (t === "mantenimiento") return "mantenimiento";
  return "ingreso";
}

function fmtMeta(s: string | undefined): string {
  if (s == null || String(s).trim() === "") return "—";
  return String(s);
}

type Props = {
  variant: "tanque" | "balanza";
  rows: HistorialRow[];
};

/**
 * Vertical timeline for tanque or balanza historial (same data as former table).
 * Order follows `rows` (API: más reciente primero).
 */
export function HistorialTimeline({ variant, rows }: Props) {
  if (rows.length === 0) {
    return (
      <p className="historial-timeline-empty" style={{ color: "var(--muted)" }}>
        Sin historial.
      </p>
    );
  }

  return (
    <div className="historial-timeline">
      <p className="historial-timeline-hint">Orden: más reciente arriba (igual que antes en la tabla).</p>
      <ol className="historial-timeline-list" role="list">
        {rows.map((h) => {
          const op = operacionVisual(h);
          const ctx =
            variant === "tanque" ? balanzasResumenDisplay(h) : tanquesResumenDisplay(h);
          const ctxLabel = variant === "tanque" ? "Balanzas (asiento)" : "Tanques (asiento)";
          const dotClass =
            op === "egreso"
              ? "historial-timeline-dot historial-timeline-dot--egreso"
              : op === "mantenimiento"
                ? "historial-timeline-dot historial-timeline-dot--mantenimiento"
                : "historial-timeline-dot historial-timeline-dot--ingreso";
          const opClass =
            op === "egreso"
              ? "historial-timeline-op historial-timeline-op--egreso"
              : op === "mantenimiento"
                ? "historial-timeline-op historial-timeline-op--mantenimiento"
                : "historial-timeline-op historial-timeline-op--ingreso";
          const arrow =
            op === "egreso" ? "↓" : op === "mantenimiento" ? "◆" : "↑";

          return (
            <li key={h.id} className="historial-timeline-item">
              <span className={dotClass} aria-hidden />
              <article className="historial-timeline-card">
                <header className="historial-timeline-head">
                  <time className="historial-timeline-time" dateTime={h.fecha_asignacion}>
                    {formatHistorialFecha(h.fecha_asignacion)}
                  </time>
                  <span className={opClass}>
                    <span className="historial-timeline-arrow" aria-hidden>
                      {arrow}
                    </span>
                    {historialOperacionLabel(h)}
                  </span>
                  {h.asiento_id != null && (
                    <Link
                      to={`/asientos/${h.asiento_id}`}
                      className="historial-timeline-asiento-link"
                    >
                      Asiento #{h.asiento_id}
                    </Link>
                  )}
                </header>
                <div className="historial-timeline-meta">
                  <span>Cantidad: {fmtMeta(h.cantidad_movimiento)}</span>
                </div>
                <p className="historial-timeline-ctx-label">{ctxLabel}</p>
                <p className="historial-timeline-ctx">{ctx}</p>
                <div className="historial-timeline-desc">
                  <HistorialDescripcionCell row={h} />
                </div>
              </article>
            </li>
          );
        })}
      </ol>
    </div>
  );
}
