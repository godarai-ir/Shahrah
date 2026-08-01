type ControlPlaneResult<T> = { ok: true; payload: T } | { ok: false; status: number; message: string };

export async function callSocialAuthControlPlane<T>(path: string, body: unknown): Promise<ControlPlaneResult<T>> {
  const masterKey = process.env.SHAHRAH_MASTER_KEY ?? process.env.MASTER_KEY;
  if (!masterKey) {
    return { ok: false, status: 500, message: "Control-plane credentials are not configured." };
  }
  const baseURL = process.env.SHAHRAH_CONTROL_PLANE_URL ?? "http://127.0.0.1:4000";
  try {
    const response = await fetch(new URL(path.replace(/^\//, ""), ensureTrailingSlash(baseURL)), {
      method: "POST",
      headers: {
        authorization: `Bearer ${masterKey}`,
        accept: "application/json",
        "content-type": "application/json",
      },
      body: JSON.stringify(body),
      cache: "no-store",
    });
    const text = await response.text();
    const payload = safeJSON(text);
    if (!response.ok) {
      return { ok: false, status: response.status, message: extractMessage(payload) ?? "Social authentication failed." };
    }
    return { ok: true, payload: payload as T };
  } catch {
    return { ok: false, status: 502, message: "Control-plane is unavailable." };
  }
}

function ensureTrailingSlash(value: string) { return value.endsWith("/") ? value : `${value}/`; }
function safeJSON(value: string): unknown { try { return value ? JSON.parse(value) : {}; } catch { return { message: value }; } }
function extractMessage(value: unknown) {
  if (!value || typeof value !== "object") return undefined;
  const payload = value as { message?: string; error?: string | { message?: string } };
  return typeof payload.error === "string" ? payload.error : payload.error?.message ?? payload.message;
}
