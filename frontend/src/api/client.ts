/** Vercel: NEXT_PUBLIC_API_URL. Fallback VITE_API_URL. Sin valor: rutas relativas + proxy en dev. */
function apiBaseUrl(): string {
  const raw =
    import.meta.env.NEXT_PUBLIC_API_URL ||
    import.meta.env.VITE_API_URL ||
    "";
  return String(raw).replace(/\/$/, "");
}

const apiBase = apiBaseUrl();

function authHeader(token: string | null): HeadersInit {
  const h: Record<string, string> = { "Content-Type": "application/json" };
  if (token) h.Authorization = `Bearer ${token}`;
  return h;
}

export async function apiFetch<T>(
  path: string,
  opts: RequestInit & { token?: string | null } = {}
): Promise<T> {
  const { token, ...init } = opts;
  const url = `${apiBase}${path.startsWith("/") ? path : `/${path}`}`;
  const res = await fetch(url, {
    ...init,
    headers: { ...authHeader(token ?? null), ...(init.headers as object) },
  });
  if (res.status === 204) return undefined as T;
  const text = await res.text();
  let body: unknown = null;
  if (text) {
    try {
      body = JSON.parse(text);
    } catch {
      body = { error: text };
    }
  }
  if (!res.ok) {
    const err = (body as { error?: string })?.error || res.statusText;
    throw new Error(err);
  }
  return body as T;
}

export type User = { id: number; email: string };

export type Tanque = { id: number; nombre: string; saldo?: string };
export type Balanza = { id: number; nombre: string; saldo?: string };

export type AsientoTanqueRow = {
  id?: number;
  asiento_id?: number;
  tanque_id: number;
  cantidad: string;
  tipo_operacion?: string;
  tanque?: Tanque;
};

export type AsientoBalanzaRow = {
  id?: number;
  asiento_id?: number;
  balanza_id: number;
  cantidad: string;
  tipo_operacion?: string;
  balanza?: Balanza;
};

export type Asiento = {
  id: number;
  fecha: string;
  descripcion: string;
  asiento_tanques?: AsientoTanqueRow[];
  asiento_balanzas?: AsientoBalanzaRow[];
};

export type HistorialRow = {
  id: number;
  tanque_id?: number;
  balanza_id?: number;
  fecha_asignacion: string;
  asiento_id?: number | null;
  descripcion: string;
  balanzas_resumen?: string;
  tanques_resumen?: string;
  cantidad_movimiento?: string;
  tipo_operacion?: string;
  tanque?: Tanque;
  balanza?: Balanza;
  asiento?: Asiento | null;
};
