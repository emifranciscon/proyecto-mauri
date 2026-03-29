import { useCallback, useEffect, useState } from "react";
import {
  apiFetch,
  type Asiento,
  type Balanza,
  type Tanque,
} from "../api/client";
import { useAuth } from "../context/AuthContext";

function fmtDate(s: string) {
  return s?.slice(0, 10) ?? "";
}

export function AsientosPage() {
  const { token } = useAuth();
  const [rows, setRows] = useState<Asiento[]>([]);
  const [tanques, setTanques] = useState<Tanque[]>([]);
  const [balanzas, setBalanzas] = useState<Balanza[]>([]);
  const [error, setError] = useState<string | null>(null);
  const [modal, setModal] = useState(false);
  const [fecha, setFecha] = useState(() =>
    new Date().toISOString().slice(0, 10)
  );
  const [tanqueId, setTanqueId] = useState("");
  const [balanzaId, setBalanzaId] = useState("");
  const [descripcion, setDescripcion] = useState("");

  const load = useCallback(async () => {
    if (!token) return;
    const [a, t, b] = await Promise.all([
      apiFetch<Asiento[]>("/api/asientos", { token }),
      apiFetch<Tanque[]>("/api/tanques", { token }),
      apiFetch<Balanza[]>("/api/balanzas", { token }),
    ]);
    setRows(a);
    setTanques(t);
    setBalanzas(b);
  }, [token]);

  useEffect(() => {
    load().catch((e) => setError(String(e.message)));
  }, [load]);

  async function create(e: React.FormEvent) {
    e.preventDefault();
    setError(null);
    try {
      await apiFetch("/api/asientos", {
        method: "POST",
        token,
        body: JSON.stringify({
          fecha,
          tanque_id: Number(tanqueId),
          balanza_id: Number(balanzaId),
          descripcion,
        }),
      });
      setModal(false);
      setDescripcion("");
      await load();
    } catch (err) {
      setError(err instanceof Error ? err.message : "Error");
    }
  }

  return (
    <div>
      <div
        style={{
          display: "flex",
          justifyContent: "space-between",
          alignItems: "center",
          marginBottom: "1rem",
          flexWrap: "wrap",
          gap: "0.75rem",
        }}
      >
        <h1 style={{ margin: 0 }}>Asientos</h1>
        <button type="button" className="btn btn-primary" onClick={() => setModal(true)}>
          Nuevo asiento
        </button>
      </div>
      {error && <div className="error-banner">{error}</div>}
      <div className="table-wrap">
        <table>
          <thead>
            <tr>
              <th>Fecha</th>
              <th>Tanque</th>
              <th>Balanza</th>
              <th>Descripción</th>
            </tr>
          </thead>
          <tbody>
            {rows.length === 0 ? (
              <tr>
                <td colSpan={4} style={{ color: "var(--muted)" }}>
                  No hay registros.
                </td>
              </tr>
            ) : (
              rows.map((r) => (
                <tr key={r.id}>
                  <td>{fmtDate(r.fecha)}</td>
                  <td>{r.tanque?.nombre ?? r.tanque_id}</td>
                  <td>{r.balanza?.nombre ?? r.balanza_id}</td>
                  <td>{r.descripcion}</td>
                </tr>
              ))
            )}
          </tbody>
        </table>
      </div>

      {modal && (
        <div className="modal-backdrop" role="presentation" onClick={() => setModal(false)}>
          <div
            className="modal"
            role="dialog"
            aria-labelledby="asiento-title"
            onClick={(e) => e.stopPropagation()}
          >
            <h2 id="asiento-title">Nuevo asiento</h2>
            <form onSubmit={create}>
              <div className="field">
                <label htmlFor="fecha">Fecha</label>
                <input
                  id="fecha"
                  className="input"
                  type="date"
                  value={fecha}
                  onChange={(e) => setFecha(e.target.value)}
                  required
                />
              </div>
              <div className="field">
                <label htmlFor="tanque">Tanque</label>
                <select
                  id="tanque"
                  className="text-input"
                  value={tanqueId}
                  onChange={(e) => setTanqueId(e.target.value)}
                  required
                >
                  <option value="">Seleccionar…</option>
                  {tanques.map((t) => (
                    <option key={t.id} value={t.id}>
                      {t.nombre}
                    </option>
                  ))}
                </select>
              </div>
              <div className="field">
                <label htmlFor="balanza">Balanza</label>
                <select
                  id="balanza"
                  className="text-input"
                  value={balanzaId}
                  onChange={(e) => setBalanzaId(e.target.value)}
                  required
                >
                  <option value="">Seleccionar…</option>
                  {balanzas.map((b) => (
                    <option key={b.id} value={b.id}>
                      {b.nombre}
                    </option>
                  ))}
                </select>
              </div>
              <div className="field">
                <label htmlFor="desc">Descripción</label>
                <textarea
                  id="desc"
                  className="input"
                  rows={3}
                  value={descripcion}
                  onChange={(e) => setDescripcion(e.target.value)}
                  required
                />
              </div>
              <div style={{ display: "flex", gap: "0.5rem", justifyContent: "flex-end" }}>
                <button type="button" className="btn btn-ghost" onClick={() => setModal(false)}>
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
