type ModelDiscoveryErrorPayload = {
  error?: { message?: string };
  message?: string;
};

export function getModelDiscoveryBaseURL(): string {
  const raw =
    process.env.SHAHRAH_MODEL_DISCOVERY_URL ??
    process.env.NEXT_PUBLIC_MODEL_CATALOG_API_BASE ??
    "http://127.0.0.1:3002/api/v1";
  return raw.endsWith("/") ? raw : `${raw}/`;
}

export async function fetchModelDiscoveryJSON<T>(path: string, revalidateSeconds = 1800): Promise<T> {
  const url = new URL(path.startsWith("/") ? path.slice(1) : path, getModelDiscoveryBaseURL()).toString();
  const response = await fetch(url, {
    headers: { Accept: "application/json" },
    next: { revalidate: revalidateSeconds },
  });
  const text = await response.text();
  let payload: unknown = {};
  if (text.trim()) {
    try {
      payload = JSON.parse(text) as unknown;
    } catch {
      payload = { message: text };
    }
  }
  if (!response.ok) {
    const err = payload as ModelDiscoveryErrorPayload;
    throw new Error(err.error?.message ?? err.message ?? `Model discovery request failed (${response.status})`);
  }
  return payload as T;
}

