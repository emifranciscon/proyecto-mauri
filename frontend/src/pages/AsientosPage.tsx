import { useCallback, useEffect, useState } from "react";
import { Link } from "react-router-dom";
import {
  apiFetch,
  type Asiento,
  type Balanza,
  type Tanque,
} from "../api/client";
import { useAuth } from "../context/AuthContext";

function formatAsientoFechaShort(iso: string): string {
  const d = new Date(iso);
  if (Number.isNaN(d.getTime())) return iso.slice(0, 10);
  return d.toLocaleString(undefined, { dateStyle: "medium", timeStyle: "short" });
}

function lineTipoLabel(tipo: string | undefined): string {
  return tipo?.trim().toLowerCase() === "egreso" ? "Egreso" : "Ingreso";
}

function formatTanquesCell(a: Asiento): string {
  const lines = a.asiento_tanques ?? [];
  if (lines.length === 0) return "—";
  return lines
    .map(
      (x) =>
        `${x.tanque?.nombre ?? x.tanque_id} (${lineTipoLabel(x.tipo_operacion)} ${x.cantidad})`
    )
    .join(", ");
}

function formatBalanzasCell(a: Asiento): string {
  const lines = a.asiento_balanzas ?? [];
  if (lines.length === 0) return "—";
  return lines
    .map(
      (x) =>
        `${x.balanza?.nombre ?? x.balanza_id} (${lineTipoLabel(x.tipo_operacion)} ${x.cantidad})`
    )
    .join(", ");
}

function toDatetimeLocalValue(d: Date): string {
  const pad = (n: number) => String(n).padStart(2, "0");
  return `${d.getFullYear()}-${pad(d.getMonth() + 1)}-${pad(d.getDate())}T${pad(d.getHours())}:${pad(d.getMinutes())}`;
}

type Line = { id: string; refId: string; tipo_operacion: "ingreso" | "egreso" };

let lineSeq = 0;
function newLine(): Line {
  lineSeq += 1;
  return { id: `l-${lineSeq}-${Date.now()}`, refId: "", tipo_operacion: "ingreso" };
}

export function AsientosPage() {
  const { token } = useAuth();
  const [rows, setRows] = useState<Asiento[]>([]);
  const [tanques, setTanques] = useState<Tanque[]>([]);
  const [balanzas, setBalanzas] = useState<Balanza[]>([]);
  const [error, setError] = useState<string | null>(null);
  const [modal, setModal] = useState(false);
  const [fecha, setFecha] = useState(() => toDatetimeLocalValue(new Date()));
  const [tanqueLines, setTanqueLines] = useState<Line[]>([newLine()]);
  const [balanzaLines, setBalanzaLines] = useState<Line[]>([newLine()]);
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

  function openModal() {
    setTanqueLines([newLine()]);
    setBalanzaLines([newLine()]);
    setFecha(toDatetimeLocalValue(new Date()));
    setDescripcion("");
    setModal(true);
  }

  async function create(e: React.FormEvent) {
    e.preventDefault();
    setError(null);
    const tanquesPayload = tanqueLines
      .filter((l) => l.refId)
      .map((l) => ({
        tanque_id: Number(l.refId),
        tipo_operacion: l.tipo_operacion,
      }));
    const balanzasPayload = balanzaLines
      .filter((l) => l.refId)
      .map((l) => ({
        balanza_id: Number(l.refId),
        tipo_operacion: l.tipo_operacion,
      }));
    if (tanquesPayload.length === 0 || balanzasPayload.length === 0) {
      setError("Agregá al menos un tanque y una balanza.");
      return;
    }
    try {
      await apiFetch("/api/asientos", {
        method: "POST",
        token,
        body: JSON.stringify({
          fecha,
          descripcion,
          tanques: tanquesPayload,
          balanzas: balanzasPayload,
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
        <button type="button" className="btn btn-primary" onClick={openModal}>
          Nuevo asiento
        </button>
      </div>
      {error && <div className="error-banner">{error}</div>}
      <div className="table-wrap">
        <table>
          <thead>
            <tr>
              <th>Fecha</th>
              <th>Tanques</th>
              <th>Balanzas</th>
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
                  <td>
                    <Link to={`/asientos/${r.id}`} className="asiento-list-link">
                      {formatAsientoFechaShort(r.fecha)}
                    </Link>
                  </td>
                  <td>{formatTanquesCell(r)}</td>
                  <td>{formatBalanzasCell(r)}</td>
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
            className="modal modal--asiento"
            role="dialog"
            aria-labelledby="asiento-title"
            onClick={(e) => e.stopPropagation()}
          >
            <h2 id="asiento-title">Nuevo asiento</h2>
            <form onSubmit={create}>
              <div className="field">
                <label htmlFor="fecha">Fecha y hora</label>
                <input
                  id="fecha"
                  className="input"
                  type="datetime-local"
                  value={fecha}
                  onChange={(e) => setFecha(e.target.value)}
                  required
                />
              </div>
              <div className="field">
                <div className="asiento-form-section-head">
                  <label>Tanques</label>
                  <button
                    type="button"
                    className="btn btn-ghost"
                    style={{ padding: "0.25rem 0.5rem", fontSize: "0.85rem" }}
                    onClick={() => setTanqueLines((prev) => [...prev, newLine()])}
                  >
                    + Agregar tanque
                  </button>
                </div>
                {tanqueLines.map((line, i) => (
                  <div key={line.id} className="asiento-form-line">
                    <select
                      className="text-input"
                      value={line.refId}
                      onChange={(e) => {
                        const v = e.target.value;
                        setTanqueLines((prev) =>
                          prev.map((x) => (x.id === line.id ? { ...x, refId: v } : x))
                        );
                      }}
                      required={i === 0}
                      aria-label="Tanque"
                    >
                      <option value="">Seleccionar…</option>
                      {tanques.map((t) => (
                        <option key={t.id} value={String(t.id)}>
                          {t.nombre}
                        </option>
                      ))}
                    </select>
                    <select
                      className="text-input"
                      value={line.tipo_operacion}
                      onChange={(e) => {
                        const v = e.target.value as "ingreso" | "egreso";
                        setTanqueLines((prev) =>
                          prev.map((x) => (x.id === line.id ? { ...x, tipo_operacion: v } : x))
                        );
                      }}
                      aria-label="Tipo de movimiento (tanque)"
                    >
                      <option value="ingreso">Ingreso</option>
                      <option value="egreso">Egreso</option>
                    </select>
                    {tanqueLines.length > 1 && (
                      <button
                        type="button"
                        className="btn btn-ghost"
                        onClick={() =>
                          setTanqueLines((prev) => prev.filter((x) => x.id !== line.id))
                        }
                      >
                        Quitar
                      </button>
                    )}
                  </div>
                ))}
              </div>
              <div className="field">
                <div className="asiento-form-section-head">
                  <label>Balanzas</label>
                  <button
                    type="button"
                    className="btn btn-ghost"
                    style={{ padding: "0.25rem 0.5rem", fontSize: "0.85rem" }}
                    onClick={() => setBalanzaLines((prev) => [...prev, newLine()])}
                  >
                    + Agregar balanza
                  </button>
                </div>
                {balanzaLines.map((line, i) => (
                  <div key={line.id} className="asiento-form-line">
                    <select
                      className="text-input"
                      value={line.refId}
                      onChange={(e) => {
                        const v = e.target.value;
                        setBalanzaLines((prev) =>
                          prev.map((x) => (x.id === line.id ? { ...x, refId: v } : x))
                        );
                      }}
                      required={i === 0}
                      aria-label="Balanza"
                    >
                      <option value="">Seleccionar…</option>
                      {balanzas.map((b) => (
                        <option key={b.id} value={String(b.id)}>
                          {b.nombre}
                        </option>
                      ))}
                    </select>
                    <select
                      className="text-input"
                      value={line.tipo_operacion}
                      onChange={(e) => {
                        const v = e.target.value as "ingreso" | "egreso";
                        setBalanzaLines((prev) =>
                          prev.map((x) => (x.id === line.id ? { ...x, tipo_operacion: v } : x))
                        );
                      }}
                      aria-label="Tipo de movimiento (balanza)"
                    >
                      <option value="ingreso">Ingreso</option>
                      <option value="egreso">Egreso</option>
                    </select>
                    {balanzaLines.length > 1 && (
                      <button
                        type="button"
                        className="btn btn-ghost"
                        onClick={() =>
                          setBalanzaLines((prev) => prev.filter((x) => x.id !== line.id))
                        }
                      >
                        Quitar
                      </button>
                    )}
                  </div>
                ))}
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
