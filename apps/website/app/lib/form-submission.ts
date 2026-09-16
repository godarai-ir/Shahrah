// Server-side helpers shared by the public website form APIs (contact, on-premise request,
// BuildGuild survey), so every form parses, filters bots and delivers email the same way.
import { NextResponse } from "next/server";

export type FormBody = Record<string, unknown>;

export async function readFormBody(request: Request): Promise<FormBody | null> {
  const body = (await request.json().catch(() => null)) as unknown;
  return body && typeof body === "object" && !Array.isArray(body) ? (body as FormBody) : null;
}

// A hidden field catches basic form bots without revealing whether an email was sent.
export function isBotSubmission(body: FormBody) {
  return typeof body.website === "string" && body.website.trim() !== "";
}

export function normalizeField(value: unknown) {
  return typeof value === "string" ? value.trim() : "";
}

export function formOk() {
  return NextResponse.json({ ok: true });
}

export function formError(error: string, status = 400) {
  return NextResponse.json({ ok: false, error }, { status });
}

export async function deliverFormEmail(flow: string, send: () => Promise<void>, failureMessage: string) {
  try {
    await send();
    return formOk();
  } catch (error) {
    console.error(`[website][${flow}] failed to send email`, error);
    return formError(failureMessage, 502);
  }
}
