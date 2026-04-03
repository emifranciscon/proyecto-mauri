import { jsPDF } from "jspdf";
import autoTable from "jspdf-autotable";
import type { HistorialRow } from "../api/client";

function fmtDate(s: string) {
  return s?.slice(0, 10) ?? "";
}

function fmtHistorialMeta(s: string | undefined): string {
  if (s == null || String(s).trim() === "") return "—";
  return String(s);
}

function descripcionHistorialCell(h: HistorialRow): string {
  const d = h.descripcion?.trim() || h.asiento?.descripcion?.trim();
  return d || "—";
}

/** Same text as table/PDF description column (UI + export). */
export function historialDescripcionPlain(h: HistorialRow): string {
  return descripcionHistorialCell(h);
}

/** Label for timeline/PDF from persisted tipo_operacion (legacy empty → Ingreso). */
export function historialOperacionLabel(h: HistorialRow): string {
  const t = h.tipo_operacion?.trim().toLowerCase();
  if (t === "egreso") return "Egreso";
  return "Ingreso";
}

function asientoLineTipo(tipo: string | undefined): string {
  return tipo?.trim().toLowerCase() === "egreso" ? "Egreso" : "Ingreso";
}

function fallbackBalanzasFromAsiento(h: HistorialRow): string {
  const a = h.asiento;
  if (!a?.asiento_balanzas?.length) return "—";
  return a.asiento_balanzas
    .map(
      (x) =>
        `${x.balanza?.nombre ?? x.balanza_id} (${asientoLineTipo(x.tipo_operacion)} ${x.cantidad})`
    )
    .join(", ");
}

function fallbackTanquesFromAsiento(h: HistorialRow): string {
  const a = h.asiento;
  if (!a?.asiento_tanques?.length) return "—";
  return a.asiento_tanques
    .map(
      (x) =>
        `${x.tanque?.nombre ?? x.tanque_id} (${asientoLineTipo(x.tipo_operacion)} ${x.cantidad})`
    )
    .join(", ");
}

/** UI + PDF: balanzas column for historial de tanques. */
export function balanzasResumenDisplay(h: HistorialRow): string {
  if (h.balanzas_resumen?.trim()) return h.balanzas_resumen.trim();
  return fallbackBalanzasFromAsiento(h);
}

/** UI + PDF: tanques column for historial de balanzas. */
export function tanquesResumenDisplay(h: HistorialRow): string {
  if (h.tanques_resumen?.trim()) return h.tanques_resumen.trim();
  return fallbackTanquesFromAsiento(h);
}

/** PDF rows for historial de tanques (tanque omitted; shown in PDF subtitle). */
export function historialRowsToBodyTanque(rows: HistorialRow[]): string[][] {
  return rows.map((h) => [
    fmtDate(h.fecha_asignacion),
    historialOperacionLabel(h),
    balanzasResumenDisplay(h),
    descripcionHistorialCell(h),
    fmtHistorialMeta(h.cantidad_movimiento),
    h.asiento_id != null ? String(h.asiento_id) : "—",
  ]);
}

/** PDF rows for historial de balanzas (balanza omitted; shown in PDF subtitle). */
export function historialRowsToBodyBalanza(rows: HistorialRow[]): string[][] {
  return rows.map((h) => [
    fmtDate(h.fecha_asignacion),
    historialOperacionLabel(h),
    tanquesResumenDisplay(h),
    descripcionHistorialCell(h),
    fmtHistorialMeta(h.cantidad_movimiento),
    h.asiento_id != null ? String(h.asiento_id) : "—",
  ]);
}

function sanitizeFileBase(name: string): string {
  const trimmed = name.trim().slice(0, 120);
  return trimmed.replace(/[/\\?%*:|"<>]/g, "_").replace(/\s+/g, "_") || "export";
}

export type ExportDataTablePdfParams = {
  documentTitle: string;
  subtitle?: string;
  head: string[];
  body: string[][];
  fileBaseName: string;
  landscape?: boolean;
  /** When body is empty, one placeholder row (one cell per column). */
  emptyPlaceholder?: string[];
};

/**
 * Client-side PDF with grid borders, bold header row, multi-page support,
 * and repeated table headers on each page (jspdf-autotable).
 */
export function exportDataTablePdf(params: ExportDataTablePdfParams): void {
  const {
    documentTitle,
    subtitle,
    head,
    body,
    fileBaseName,
    landscape = true,
    emptyPlaceholder,
  } = params;

  const colCount = head.length;
  const placeholder =
    emptyPlaceholder && emptyPlaceholder.length === colCount
      ? emptyPlaceholder
      : Array.from({ length: colCount }, (_, i) => (i === 0 ? "Sin datos." : ""));

  const tableBody = body.length > 0 ? body : [placeholder];

  const doc = new jsPDF({
    orientation: landscape ? "landscape" : "portrait",
    unit: "mm",
    format: "a4",
  });

  const margin = 14;
  let y = 16;
  doc.setFont("helvetica", "bold");
  doc.setFontSize(14);
  doc.text(documentTitle, margin, y);
  y += 7;
  if (subtitle) {
    doc.setFont("helvetica", "normal");
    doc.setFontSize(10);
    const subtitleLines = doc.splitTextToSize(
      subtitle,
      doc.internal.pageSize.getWidth() - margin * 2
    );
    doc.text(subtitleLines, margin, y);
    y += Math.max(6, subtitleLines.length * 5);
  }
  doc.setFontSize(8);
  doc.setTextColor(100);
  doc.text(`Generado: ${new Date().toISOString().slice(0, 10)}`, margin, y);
  doc.setTextColor(0);
  y += 6;

  autoTable(doc, {
    startY: y,
    head: [head],
    body: tableBody,
    theme: "grid",
    styles: {
      fontSize: 9,
      cellPadding: 1.8,
      valign: "top",
      overflow: "linebreak",
    },
    headStyles: {
      fillColor: [61, 158, 120],
      textColor: 255,
      fontStyle: "bold",
      halign: "left",
    },
    bodyStyles: { fontStyle: "normal", halign: "left" },
    showHead: "everyPage",
    margin: { left: margin, right: margin, top: margin, bottom: 18 },
    didDrawPage: (data) => {
      const pageH = doc.internal.pageSize.getHeight();
      doc.setFont("helvetica", "normal");
      doc.setFontSize(8);
      doc.setTextColor(120);
      doc.text(`Página ${data.pageNumber}`, margin, pageH - 8);
      doc.setTextColor(0);
    },
  });

  doc.save(`${sanitizeFileBase(fileBaseName)}.pdf`);
}
