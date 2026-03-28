const apiBase = (import.meta.env.VITE_API_URL ?? "").replace(/\/$/, "");

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

export type Tanque = { id: number; nombre: string };
export type Balanza = { id: number; nombre: string };
export type Asiento = {
  id: number;
  fecha: string;
  tanque_id: number;
  balanza_id: number;
  descripcion: string;
  tanque?: Tanque;
  balanza?: Balanza;
};

export type HistorialRow = {
  id: number;
  fecha_asignacion: string;
  asiento_id?: number | null;
  descripcion: string;
  asiento?: Asiento | null;
};
