import { NextRequest, NextResponse } from "next/server";

import { readAuthFromRequest } from "@/lib/auth/request-auth";
import { proxyToControlPlane } from "@/lib/control-plane";

export const dynamic = "force-dynamic";

type PromptRunPayload = {
  request?: Record<string, unknown>;
  cache_config?: Record<string, unknown> | null;
  guardrails?: {
    input?: string[];
    output?: string[];
    scope?: "all" | "last" | "";
  } | null;
};

export async function POST(request: NextRequest) {
  const auth = await readAuthFromRequest(request);
  if (!auth.ok) {
    return NextResponse.json({ error: { message: auth.message } }, { status: auth.status });
  }
  const body = (await request.json().catch(() => null)) as PromptRunPayload | null;
  if (!body || typeof body !== "object" || !body.request || typeof body.request !== "object") {
    return NextResponse.json({ error: { message: "Invalid payload." } }, { status: 400 });
  }

  const upstreamHeaders = new Headers();
  if (body.cache_config && Object.keys(body.cache_config).length > 0) {
    upstreamHeaders.set("x-shahrah-cache-config", JSON.stringify(body.cache_config));
  }

  const input = Array.isArray(body.guardrails?.input) ? body.guardrails?.input.filter(Boolean) : [];
  const output = Array.isArray(body.guardrails?.output) ? body.guardrails?.output.filter(Boolean) : [];
  if (input.length > 0 || output.length > 0) {
    upstreamHeaders.set(
      "x-shahrah-guardrails",
      JSON.stringify({
        llm_input_guardrails: input,
        llm_output_guardrails: output,
      }),
    );
    if (body.guardrails?.scope) {
      upstreamHeaders.set("x-shahrah-guardrails-scope", body.guardrails.scope);
    }
  }

  return proxyToControlPlane({
    request,
    method: "POST",
    path: "/v1/chat/completions",
    body: body.request,
    userID: auth.user.id,
    upstreamHeaders,
  });
}
