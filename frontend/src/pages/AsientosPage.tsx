import { useCallback, useEffect, useRef, useState } from "react";
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
  const t = tipo?.trim().toLowerCase();
  if (t === "egreso") return "Egreso";
  if (t === "mantenimiento") return "Mantenimiento";
  return "Ingreso";
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

type TipoOperacionLine = "ingreso" | "egreso" | "mantenimiento";
type Line = { id: string; refId: string; tipo_operacion: TipoOperacionLine };

let lineSeq = 0;
function newLine(): Line {
  lineSeq += 1;
  return { id: `l-${lineSeq}-${Date.now()}`, refId: "", tipo_operacion: "ingreso" };
}

type EntityOption = { id: number; nombre: string };

type AsientoLinesEditorProps = {
  sectionTitle: string;
  addButtonLabel: string;
  entityLabel: string;
  lines: Line[];
  setLines: React.Dispatch<React.SetStateAction<Line[]>>;
  options: EntityOption[];
  disabled: boolean;
};

function AsientoLinesEditor({
  sectionTitle,
  addButtonLabel,
  entityLabel,
  lines,
  setLines,
  options,
  disabled,
}: AsientoLinesEditorProps) {
  const canRemove = lines.length > 1;
  return (
    <section className="asiento-modal-section" aria-label={sectionTitle}>
      <div className="asiento-modal-section__head">
        <h3 className="asiento-modal-section__title">{sectionTitle}</h3>
        <button
          type="button"
          className="btn btn-ghost btn-add-line"
          disabled={disabled}
          onClick={() => setLines((prev) => [...prev, newLine()])}
        >
          {addButtonLabel}
        </button>
      </div>
      <ul className="asiento-line-list">
        {lines.map((line, i) => (
          <li key={line.id}>
            <div className="asiento-line-card">
              <div
                className={
                  "asiento-line-card__grid" +
                  (canRemove ? " asiento-line-card__grid--with-remove" : "")
                }
              >
                <div className="asiento-line-field">
                  <label htmlFor={`${line.id}-entity`}>{entityLabel}</label>
                  <select
                    id={`${line.id}-entity`}
                    className="text-input"
                    value={line.refId}
                    disabled={disabled}
                    onChange={(e) => {
                      const v = e.target.value;
                      setLines((prev) =>
                        prev.map((x) => (x.id === line.id ? { ...x, refId: v } : x))
                      );
                    }}
                    required={i === 0}
                  >
                    <option value="">Elegir…</option>
                    {options.map((o) => (
                      <option key={o.id} value={String(o.id)}>
                        {o.nombre}
                      </option>
                    ))}
                  </select>
                </div>
                <div className="asiento-line-field">
                  <label htmlFor={`${line.id}-tipo`}>Operación</label>
                  <select
                    id={`${line.id}-tipo`}
                    className="text-input"
                    value={line.tipo_operacion}
                    disabled={disabled}
                    onChange={(e) => {
                      const v = e.target.value as TipoOperacionLine;
                      setLines((prev) =>
                        prev.map((x) => (x.id === line.id ? { ...x, tipo_operacion: v } : x))
                      );
                    }}
                  >
                    <option value="ingreso">Ingreso</option>
                    <option value="egreso">Egreso</option>
                    <option value="mantenimiento">Mantenimiento</option>
                  </select>
                </div>
                {canRemove && (
                  <button
                    type="button"
                    className="btn btn-ghost asiento-line-remove"
                    disabled={disabled}
                    onClick={() =>
                      setLines((prev) => prev.filter((x) => x.id !== line.id))
                    }
                  >
                    Quitar línea
                  </button>
                )}
              </div>
            </div>
          </li>
        ))}
      </ul>
    </section>
  );
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
  const [savingAsiento, setSavingAsiento] = useState(false);
  const savingAsientoRef = useRef(false);

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
    setError(null);
    setTanqueLines([newLine()]);
    setBalanzaLines([newLine()]);
    setFecha(toDatetimeLocalValue(new Date()));
    setDescripcion("");
    setModal(true);
  }

  async function create(e: React.FormEvent) {
    e.preventDefault();
    if (savingAsientoRef.current) return;
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
      setError("Agregá al menos un tanque y una balanza con datos completos.");
      return;
    }
    savingAsientoRef.current = true;
    setSavingAsiento(true);
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
      setError(null);
      await load();
    } catch (err) {
      setError(err instanceof Error ? err.message : "Error");
    } finally {
      savingAsientoRef.current = false;
      setSavingAsiento(false);
    }
  }

  const tanqueOptions: EntityOption[] = tanques.map((t) => ({
    id: t.id,
    nombre: t.nombre,
  }));
  const balanzaOptions: EntityOption[] = balanzas.map((b) => ({
    id: b.id,
    nombre: b.nombre,
  }));

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
      {error && !modal && <div className="error-banner">{error}</div>}
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
        <div
          className="modal-backdrop"
          role="presentation"
          onClick={() => !savingAsiento && setModal(false)}
        >
          <div
            className="modal modal--asiento"
            role="dialog"
            aria-modal="true"
            aria-labelledby="asiento-title"
            onClick={(e) => e.stopPropagation()}
          >
            <button
              type="button"
              className="asiento-modal__close"
              aria-label="Cerrar"
              disabled={savingAsiento}
              onClick={() => setModal(false)}
            >
              ×
            </button>
            <div className="asiento-modal__header">
              <h2 id="asiento-title">Nuevo asiento</h2>
              <p className="asiento-modal__header-lead">
                Completá fecha, líneas de <strong>tanques</strong> y <strong>balanzas</strong>, y una
                descripción. <strong>Mantenimiento</strong> registra el movimiento en el historial{" "}
                <strong>sin cambiar el saldo</strong> del tanque ni de la balanza.
              </p>
            </div>

            <form onSubmit={create} className="asiento-modal__form" noValidate>
              <div className="asiento-modal__body">
                {error && (
                  <div className="asiento-modal__error" role="alert">
                    {error}
                  </div>
                )}

                <fieldset className="asiento-modal__fieldset asiento-modal__datetime">
                  <legend className="asiento-modal__legend">Fecha y hora del asiento</legend>
                  <div className="field" style={{ marginBottom: 0 }}>
                    <input
                      id="fecha"
                      className="input"
                      type="datetime-local"
                      aria-label="Fecha y hora del asiento"
                      value={fecha}
                      disabled={savingAsiento}
                      onChange={(e) => setFecha(e.target.value)}
                      required
                    />
                  </div>
                </fieldset>

                <div className="asiento-modal__split">
                  <AsientoLinesEditor
                    sectionTitle="Tanques"
                    addButtonLabel="+ Otra línea"
                    entityLabel="Tanque"
                    lines={tanqueLines}
                    setLines={setTanqueLines}
                    options={tanqueOptions}
                    disabled={savingAsiento}
                  />
                  <AsientoLinesEditor
                    sectionTitle="Balanzas"
                    addButtonLabel="+ Otra línea"
                    entityLabel="Balanza"
                    lines={balanzaLines}
                    setLines={setBalanzaLines}
                    options={balanzaOptions}
                    disabled={savingAsiento}
                  />
                </div>

                <fieldset className="asiento-modal__fieldset">
                  <legend className="asiento-modal__legend">Descripción</legend>
                  <div className="field" style={{ marginBottom: 0 }}>
                    <textarea
                      id="desc"
                      className="input"
                      rows={4}
                      placeholder="Detalle del movimiento, referencias, observaciones…"
                      aria-label="Descripción del asiento"
                      value={descripcion}
                      disabled={savingAsiento}
                      onChange={(e) => setDescripcion(e.target.value)}
                      required
                    />
                  </div>
                </fieldset>
              </div>

              <div className="asiento-modal__footer">
                <button
                  type="button"
                  className="btn btn-ghost"
                  disabled={savingAsiento}
                  onClick={() => setModal(false)}
                >
                  Cancelar
                </button>
                <button type="submit" className="btn btn-primary" disabled={savingAsiento}>
                  {savingAsiento ? "Guardando…" : "Guardar asiento"}
                </button>
              </div>
            </form>
          </div>
        </div>
      )}
    </div>
  );
}
