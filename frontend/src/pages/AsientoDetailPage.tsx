import { useCallback, useEffect, useState } from "react";
import { Link, useParams } from "react-router-dom";
import { apiFetch, type Asiento, type AsientoBalanzaRow, type AsientoTanqueRow } from "../api/client";
import { useAuth } from "../context/AuthContext";

function formatAsientoFecha(iso: string): string {
  const d = new Date(iso);
  if (Number.isNaN(d.getTime())) return iso?.slice(0, 16) ?? "—";
  return d.toLocaleString(undefined, {
    weekday: "long",
    year: "numeric",
    month: "long",
    day: "numeric",
    hour: "2-digit",
    minute: "2-digit",
  });
}

function lineTipoShort(tipo: string | undefined): string {
  return tipo?.trim().toLowerCase() === "egreso" ? "Egreso" : "Ingreso";
}

function ChipTanque({ row }: { row: AsientoTanqueRow }) {
  const name = row.tanque?.nombre ?? `Tanque #${row.tanque_id}`;
  return (
    <li className="asiento-chip asiento-chip--tanque">
      <span className="asiento-chip-name">{name}</span>
      <span className="asiento-chip-meta">
        {lineTipoShort(row.tipo_operacion)} × {row.cantidad}
      </span>
    </li>
  );
}

function ChipBalanza({ row }: { row: AsientoBalanzaRow }) {
  const name = row.balanza?.nombre ?? `Balanza #${row.balanza_id}`;
  return (
    <li className="asiento-chip asiento-chip--balanza">
      <span className="asiento-chip-name">{name}</span>
      <span className="asiento-chip-meta">
        {lineTipoShort(row.tipo_operacion)} × {row.cantidad}
      </span>
    </li>
  );
}

export function AsientoDetailPage() {
  const { id } = useParams();
  const { token } = useAuth();
  const [row, setRow] = useState<Asiento | null>(null);
  const [error, setError] = useState<string | null>(null);
  const [loading, setLoading] = useState(true);

  const load = useCallback(async () => {
    if (!token || !id) return;
    setLoading(true);
    setError(null);
    try {
      const a = await apiFetch<Asiento>(`/api/asientos/${id}`, { token });
      setRow(a);
    } catch (e) {
      setRow(null);
      setError(e instanceof Error ? e.message : "No se encontró el asiento.");
    } finally {
      setLoading(false);
    }
  }, [token, id]);

  useEffect(() => {
    load().catch(() => {});
  }, [load]);

  const tanques = row?.asiento_tanques ?? [];
  const balanzas = row?.asiento_balanzas ?? [];

  return (
    <div className="asiento-detail-page">
      <nav className="asiento-detail-nav" aria-label="Navegación">
        <Link to="/asientos" className="asiento-detail-back">
          ← Asientos
        </Link>
      </nav>

      {loading && (
        <p className="asiento-detail-loading" style={{ color: "var(--muted)" }}>
          Cargando…
        </p>
      )}

      {error && !loading && <div className="error-banner">{error}</div>}

      {row && !error && !loading && (
        <>
          <header className="asiento-detail-hero">
            <span className="asiento-detail-id-badge">#{row.id}</span>
            <h1 className="asiento-detail-h1">Asiento</h1>
            <p className="asiento-detail-when">{formatAsientoFecha(row.fecha)}</p>
          </header>

          <div className="asiento-detail-columns">
            <section className="asiento-panel">
              <h2 className="asiento-panel-title">Información general</h2>
              <dl className="asiento-kv">
                <div className="asiento-kv-row">
                  <dt>ID</dt>
                  <dd>{row.id}</dd>
                </div>
              </dl>
              <div className="asiento-desc-block">
                <h3 className="asiento-desc-label">Descripción</h3>
                <p className="asiento-desc-text">{row.descripcion}</p>
              </div>
            </section>

            <section className="asiento-panel">
              <h2 className="asiento-panel-title">Tanques</h2>
              {tanques.length === 0 ? (
                <p className="asiento-panel-empty">Ninguno.</p>
              ) : (
                <ul className="asiento-chip-list" role="list">
                  {tanques.map((t, i) => (
                    <ChipTanque
                      key={t.id != null ? `t-${t.id}` : `t-${i}-${t.tanque_id}`}
                      row={t}
                    />
                  ))}
                </ul>
              )}
            </section>

            <section className="asiento-panel">
              <h2 className="asiento-panel-title">Balanzas</h2>
              {balanzas.length === 0 ? (
                <p className="asiento-panel-empty">Ninguna.</p>
              ) : (
                <ul className="asiento-chip-list" role="list">
                  {balanzas.map((b, i) => (
                    <ChipBalanza
                      key={b.id != null ? `b-${b.id}` : `b-${i}-${b.balanza_id}`}
                      row={b}
                    />
                  ))}
                </ul>
              )}
            </section>
          </div>
        </>
      )}
    </div>
  );
}
