import { useCallback, useEffect, useState } from "react";
import { useSearchParams } from "react-router-dom";
import { apiFetch, type HistorialRow, type Tanque, type Asiento } from "../api/client";
import { useAuth } from "../context/AuthContext";
import { HistorialTimeline } from "../components/HistorialTimeline";
import { exportDataTablePdf, historialRowsToBodyTanque } from "../utils/pdfTableExport";

function fmtDate(s: string) {
  return s?.slice(0, 10) ?? "";
}

export function TanquesPage() {
  const { token } = useAuth();
  const [searchParams] = useSearchParams();
  const [list, setList] = useState<Tanque[]>([]);
  const [selected, setSelected] = useState<Tanque | null>(null);
  const [historial, setHistorial] = useState<HistorialRow[]>([]);
  const [asientos, setAsientos] = useState<Asiento[]>([]);
  const [error, setError] = useState<string | null>(null);
  const [newNombre, setNewNombre] = useState("");
  const [histModal, setHistModal] = useState(false);
  const [fa, setFa] = useState(() => new Date().toISOString().slice(0, 10));
  const [asientoId, setAsientoId] = useState("");
  const [hDesc, setHDesc] = useState("");

  const refreshList = useCallback(async () => {
    if (!token) return;
    const t = await apiFetch<Tanque[]>("/api/tanques", { token });
    setList(t);
    setSelected((prev) => {
      if (t.length === 0) return null;
      if (prev && t.some((x) => x.id === prev.id)) return prev;
      return t[0];
    });
  }, [token]);

  const loadHistorial = useCallback(
    async (id: number) => {
      if (!token) return;
      const h = await apiFetch<HistorialRow[]>(`/api/tanques/${id}/historial`, {
        token,
      });
      setHistorial(h);
    },
    [token]
  );

  const loadAsientos = useCallback(async () => {
    if (!token) return;
    const a = await apiFetch<Asiento[]>("/api/asientos", { token });
    setAsientos(a);
  }, [token]);

  useEffect(() => {
    refreshList().catch((e) => setError(String(e.message)));
  }, [refreshList]);

  useEffect(() => {
    const tid = searchParams.get("tanque");
    if (!tid || list.length === 0) return;
    const n = Number(tid);
    if (!Number.isFinite(n)) return;
    const found = list.find((x) => x.id === n);
    if (found) setSelected(found);
  }, [list, searchParams]);

  useEffect(() => {
    if (!selected || !token) {
      setHistorial([]);
      return;
    }
    loadHistorial(selected.id).catch((e) => setError(String(e.message)));
  }, [selected, token, loadHistorial]);

  useEffect(() => {
    loadAsientos().catch(() => {});
  }, [loadAsientos]);

  async function addTanque(e: React.FormEvent) {
    e.preventDefault();
    setError(null);
    try {
      await apiFetch("/api/tanques", {
        method: "POST",
        token,
        body: JSON.stringify({ nombre: newNombre }),
      });
      setNewNombre("");
      await refreshList();
    } catch (err) {
      setError(err instanceof Error ? err.message : "Error");
    }
  }

  function exportHistorialPdf() {
    if (!selected) return;
    setError(null);
    try {
      exportDataTablePdf({
        documentTitle: "Historial de tanques",
        subtitle: `Tanque: ${selected.nombre}`,
        head: [
          "Fecha",
          "Operación",
          "Balanzas (asiento)",
          "Descripción",
          "Cantidad",
          "Nº asiento",
        ],
        body: historialRowsToBodyTanque(historial),
        fileBaseName: `historial-tanque-${selected.nombre}`,
        landscape: true,
        emptyPlaceholder: ["Sin historial.", "", "", "", "", ""],
      });
    } catch (err) {
      setError(
        err instanceof Error ? err.message : "No se pudo generar el PDF. Probá de nuevo."
      );
    }
  }

  async function addHistorial(e: React.FormEvent) {
    e.preventDefault();
    if (!selected) return;
    setError(null);
    try {
      const body: Record<string, unknown> = {
        fecha_asignacion: fa,
        descripcion: hDesc,
      };
      if (asientoId) body.asiento_id = Number(asientoId);
      await apiFetch(`/api/tanques/${selected.id}/historial`, {
        method: "POST",
        token,
        body: JSON.stringify(body),
      });
      setHistModal(false);
      setHDesc("");
      setAsientoId("");
      await loadHistorial(selected.id);
    } catch (err) {
      setError(err instanceof Error ? err.message : "Error");
    }
  }

  return (
    <div>
      <h1 style={{ marginTop: 0 }}>Tanques</h1>
      {error && <div className="error-banner">{error}</div>}
      <div
        style={{
          display: "grid",
          gridTemplateColumns: "minmax(200px, 280px) 1fr",
          gap: "1rem",
          alignItems: "start",
        }}
        className="master-detail"
      >
        <div
          style={{
            border: "1px solid var(--border)",
            borderRadius: "var(--radius)",
            background: "var(--surface)",
            padding: "0.75rem",
          }}
        >
          <h2 style={{ fontSize: "1rem", margin: "0 0 0.75rem" }}>Tanques</h2>
          <form onSubmit={addTanque} style={{ marginBottom: "0.75rem" }}>
            <div className="field" style={{ marginBottom: "0.5rem" }}>
              <input
                className="input"
                placeholder="Nombre nuevo tanque"
                value={newNombre}
                onChange={(e) => setNewNombre(e.target.value)}
                required
              />
            </div>
            <button type="submit" className="btn btn-primary" style={{ width: "100%" }}>
              Agregar
            </button>
          </form>
          <ul style={{ listStyle: "none", margin: 0, padding: 0 }}>
            {list.map((t) => (
              <li key={t.id}>
                <button
                  type="button"
                  onClick={() => setSelected(t)}
                  className="btn btn-ghost"
                  style={{
                    width: "100%",
                    justifyContent: "flex-start",
                    marginBottom: 4,
                    background:
                      selected?.id === t.id ? "rgba(61,158,120,0.15)" : undefined,
                    borderColor:
                      selected?.id === t.id ? "var(--accent-dim)" : "var(--border)",
                  }}
                >
                  {t.nombre}
                </button>
              </li>
            ))}
          </ul>
        </div>
        <div
          style={{
            border: "1px solid var(--border)",
            borderRadius: "var(--radius)",
            background: "var(--surface)",
            padding: "0.75rem",
            minHeight: 280,
          }}
        >
          <div
            style={{
              display: "flex",
              justifyContent: "space-between",
              alignItems: "center",
              marginBottom: "0.75rem",
              flexWrap: "wrap",
              gap: "0.5rem",
            }}
          >
            <h2 style={{ fontSize: "1rem", margin: 0 }}>
              Historial{selected ? `: ${selected.nombre}` : ""}
            </h2>
            {selected && (
              <div style={{ display: "flex", flexWrap: "wrap", gap: "0.5rem" }}>
                <button
                  type="button"
                  className="btn btn-ghost"
                  onClick={exportHistorialPdf}
                >
                  Exportar PDF
                </button>
                <button
                  type="button"
                  className="btn btn-primary"
                  onClick={() => setHistModal(true)}
                >
                  Nuevo registro
                </button>
              </div>
            )}
          </div>
          {!selected ? (
            <p style={{ color: "var(--muted)" }}>Seleccioná un tanque.</p>
          ) : (
            <div className="table-wrap">
              <HistorialTimeline variant="tanque" rows={historial} />
            </div>
          )}
        </div>
      </div>

      {histModal && selected && (
        <div
          className="modal-backdrop"
          role="presentation"
          onClick={() => setHistModal(false)}
        >
          <div className="modal" onClick={(e) => e.stopPropagation()}>
            <h2>Historial — {selected.nombre}</h2>
            <form onSubmit={addHistorial}>
              <div className="field">
                <label htmlFor="fa">Fecha de asignación</label>
                <input
                  id="fa"
                  className="input"
                  type="date"
                  value={fa}
                  onChange={(e) => setFa(e.target.value)}
                  required
                />
              </div>
              <div className="field">
                <label htmlFor="aid">Registro de asiento (opcional)</label>
                <select
                  id="aid"
                  className="text-input"
                  value={asientoId}
                  onChange={(e) => setAsientoId(e.target.value)}
                >
                  <option value="">—</option>
                  {asientos.map((a) => (
                    <option key={a.id} value={a.id}>
                      #{a.id} — {fmtDate(a.fecha)}
                    </option>
                  ))}
                </select>
              </div>
              <div className="field">
                <label htmlFor="hd">Descripción</label>
                <textarea
                  id="hd"
                  className="input"
                  rows={3}
                  value={hDesc}
                  onChange={(e) => setHDesc(e.target.value)}
                  required
                />
              </div>
              <div style={{ display: "flex", gap: "0.5rem", justifyContent: "flex-end" }}>
                <button
                  type="button"
                  className="btn btn-ghost"
                  onClick={() => setHistModal(false)}
                >
                  Cancelar
                </button>
                <button type="submit" className="btn btn-primary">
                  Guardar
                </button>
              </div>
            </form>
          </div>
        </div>
      )}
    </div>
  );
}
