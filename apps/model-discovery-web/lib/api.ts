const PUBLIC_API_BASE =
  process.env.NEXT_PUBLIC_MODEL_DISCOVERY_API_BASE ||
  '/catalog/api/v1';
const INTERNAL_API_BASE =
  process.env.MODEL_DISCOVERY_API_BASE_INTERNAL || 'http://localhost:3002/api/v1';

export const API_BASE =
  typeof window === 'undefined' ? INTERNAL_API_BASE : PUBLIC_API_BASE;
export const API_ORIGIN =
  PUBLIC_API_BASE.startsWith('http://') || PUBLIC_API_BASE.startsWith('https://')
    ? new URL(PUBLIC_API_BASE).origin
    : '';

export function resolveAssetUrl(path?: string | null): string | null {
  if (!path) return null;
  if (path.startsWith('http://') || path.startsWith('https://')) return path;
  if (path.startsWith('/')) return API_ORIGIN ? `${API_ORIGIN}${path}` : path;
  return API_ORIGIN ? `${API_ORIGIN}/${path}` : `/${path}`;
}

export async function apiGet<T>(path: string, init?: RequestInit): Promise<T> {
  let res: Response;
  try {
    res = await fetch(`${API_BASE}${path}`, { ...init, cache: 'no-store' });
  } catch (error) {
    const message = error instanceof Error ? error.message : 'Unknown fetch error';
    throw new Error(`API fetch failed for ${API_BASE}${path}: ${message}`);
  }
  if (!res.ok) throw new Error(`API ${res.status}: ${await res.text()}`);
  return res.json() as Promise<T>;
}

export async function apiPost<T>(path: string, body: unknown, token?: string): Promise<T> {
  const res = await fetch(`${API_BASE}${path}`, {
    method: 'POST',
    headers: {
      'content-type': 'application/json',
      ...(token ? { authorization: `Bearer ${token}` } : {}),
    },
    body: JSON.stringify(body),
  });
  if (!res.ok) throw new Error(`API ${res.status}: ${await res.text()}`);
  return res.json() as Promise<T>;
}
