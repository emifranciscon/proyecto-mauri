import { useCallback, useEffect, useState } from "react";
import { apiFetch, type HistorialRow, type Balanza, type Asiento } from "../api/client";
import { useAuth } from "../context/AuthContext";
import { exportDataTablePdf, historialRowsToBody } from "../utils/pdfTableExport";

function fmtDate(s: string) {
  return s?.slice(0, 10) ?? "";
}

export function BalanzasPage() {
  const { token } = useAuth();
  const [list, setList] = useState<Balanza[]>([]);
  const [selected, setSelected] = useState<Balanza | null>(null);
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
    const t = await apiFetch<Balanza[]>("/api/balanzas", { token });
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
      const h = await apiFetch<HistorialRow[]>(`/api/balanzas/${id}/historial`, {
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
    if (!selected || !token) {
      setHistorial([]);
      return;
    }
    loadHistorial(selected.id).catch((e) => setError(String(e.message)));
  }, [selected, token, loadHistorial]);

  useEffect(() => {
    loadAsientos().catch(() => {});
  }, [loadAsientos]);

  function exportHistorialPdf() {
    if (!selected) return;
    setError(null);
    try {
      exportDataTablePdf({
        documentTitle: "Historial de balanzas",
        subtitle: `Balanza: ${selected.nombre}`,
        head: ["Fecha asignación", "Registro de asiento", "Descripción"],
        body: historialRowsToBody(historial),
        fileBaseName: `historial-balanza-${selected.nombre}`,
        landscape: true,
        emptyPlaceholder: ["Sin historial.", "", ""],
      });
    } catch (err) {
      setError(
        err instanceof Error ? err.message : "No se pudo generar el PDF. Probá de nuevo."
      );
    }
  }

  async function addBalanza(e: React.FormEvent) {
    e.preventDefault();
    setError(null);
    try {
      await apiFetch("/api/balanzas", {
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
      await apiFetch(`/api/balanzas/${selected.id}/historial`, {
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
      <h1 style={{ marginTop: 0 }}>Balanzas</h1>
      {error && <div className="error-banner">{error}</div>}
      <div
        style={{
          display: "grid",
          gridTemplateColumns: "minmax(200px, 280px) 1fr",
          gap: "1rem",
          alignItems: "start",
        }}
      >
        <div
          style={{
            border: "1px solid var(--border)",
            borderRadius: "var(--radius)",
            background: "var(--surface)",
            padding: "0.75rem",
          }}
        >
          <h2 style={{ fontSize: "1rem", margin: "0 0 0.75rem" }}>Balanzas</h2>
          <form onSubmit={addBalanza} style={{ marginBottom: "0.75rem" }}>
            <div className="field" style={{ marginBottom: "0.5rem" }}>
              <input
                className="input"
                placeholder="Nombre nueva balanza"
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
            {list.map((b) => (
              <li key={b.id}>
                <button
                  type="button"
                  onClick={() => setSelected(b)}
                  className="btn btn-ghost"
                  style={{
                    width: "100%",
                    justifyContent: "flex-start",
                    marginBottom: 4,
                    background:
                      selected?.id === b.id ? "rgba(61,158,120,0.15)" : undefined,
                    borderColor:
                      selected?.id === b.id ? "var(--accent-dim)" : "var(--border)",
                  }}
                >
                  {b.nombre}
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
            <p style={{ color: "var(--muted)" }}>Seleccioná una balanza.</p>
          ) : (
            <div className="table-wrap">
              <table>
                <thead>
                  <tr>
                    <th>Fecha asignación</th>
                    <th>Registro de asiento</th>
                    <th>Descripción</th>
                  </tr>
                </thead>
                <tbody>
                  {historial.length === 0 ? (
                    <tr>
                      <td colSpan={3} style={{ color: "var(--muted)" }}>
                        Sin historial.
                      </td>
                    </tr>
                  ) : (
                    historial.map((h) => (
                      <tr key={h.id}>
                        <td>{fmtDate(h.fecha_asignacion)}</td>
                        <td>
                          {h.asiento_id
                            ? `#${h.asiento_id}${h.asiento ? ` — ${fmtDate(h.asiento.fecha)}` : ""}`
                            : "—"}
                        </td>
                        <td>{h.descripcion}</td>
                      </tr>
                    ))
                  )}
                </tbody>
              </table>
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
                <label htmlFor="bfa">Fecha de asignación</label>
                <input
                  id="bfa"
                  className="input"
                  type="date"
                  value={fa}
                  onChange={(e) => setFa(e.target.value)}
                  required
                />
              </div>
              <div className="field">
                <label htmlFor="baid">Registro de asiento (opcional)</label>
                <select
                  id="baid"
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
                <label htmlFor="bhd">Descripción</label>
                <textarea
                  id="bhd"
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
