import {
  createUIMessageStream,
  createUIMessageStreamResponse,
} from "ai";
import { type NextRequest, NextResponse } from "next/server";

import type { ManagedRAGUIMessage } from "@/features/managed-rag/chat-types";
import { readAuthFromRequest } from "@/lib/auth/request-auth";
import { chatbotResponseSchema } from "@/lib/contracts/managed-rag";
import { fetchControlPlaneJSON } from "@/lib/control-plane";

export const dynamic = "force-dynamic";

type RouteContext = { params: Promise<{ id: string }> };

const maxQueryCharacters = 8000;

export async function POST(request: NextRequest, context: RouteContext) {
  const auth = await readAuthFromRequest(request);
  if (!auth.ok) {
    return NextResponse.json({ error: { message: auth.message } }, { status: auth.status });
  }

  const { id } = await context.params;
  const body = await request.json().catch(() => null) as {
    id?: unknown;
    messages?: unknown;
  } | null;
  const messages = parseMessages(body?.messages);
  if (!messages) {
    return NextResponse.json({ error: { message: "Invalid chat messages." } }, { status: 400 });
  }

  const message = lastUserText(messages);
  if (!message) {
    return NextResponse.json({ error: { message: "A non-empty user message is required." } }, { status: 400 });
  }

  if ([...message].length > maxQueryCharacters) {
    return NextResponse.json({ error: { message: `Chat messages must not exceed ${maxQueryCharacters} characters.` } }, { status: 400 });
  }

  const sessionID = lastSessionID(messages);
  const stream = createUIMessageStream<ManagedRAGUIMessage>({
    originalMessages: messages,
    execute: async ({ writer }) => {
      writer.write({ type: "data-status", data: { stage: "retrieving" }, transient: true });
      writer.write({ type: "data-status", data: { stage: "reranking" }, transient: true });

      const upstream = await fetchControlPlaneJSON({
        tenantSlug: auth.tenantSlug,
        userID: auth.user.id,
        method: "POST",
        path: `/api/v1/chatbots/${encodeURIComponent(id)}/preview-responses`,
        body: sessionID ? { message, session_id: sessionID } : { message },
        requestHeaders: request.headers,
        signal: request.signal,
      });
      if (!upstream.ok) {
        throw new Error(upstream.errorMessage ?? "Managed RAG response failed.");
      }

      const result = chatbotResponseSchema.parse(upstream.payload);
      const textID = crypto.randomUUID();
      writer.write({ type: "text-start", id: textID });
      writer.write({ type: "text-delta", id: textID, delta: result.answer });
      writer.write({ type: "text-end", id: textID });
      writer.write({
        type: "data-result",
        data: {
          citations: result.citations,
          refusal: result.refusal,
          actualModel: result.actual_model,
          requestID: result.request_id,
          traceID: result.trace_id,
          sessionID: result.session_id,
        },
      });
    },
    onError: (error) => error instanceof Error ? error.message : "Managed RAG response failed.",
  });

  return createUIMessageStreamResponse({ stream });
}

function lastSessionID(messages: ManagedRAGUIMessage[]): string | null {
  for (let messageIndex = messages.length - 1; messageIndex >= 0; messageIndex -= 1) {
    const message = messages[messageIndex];
    if (message?.role !== "assistant") continue;

    for (let partIndex = message.parts.length - 1; partIndex >= 0; partIndex -= 1) {
      const part = message.parts[partIndex];
      if (part?.type !== "data-result") continue;

      const sessionID = chatbotResponseSchema.shape.session_id.safeParse(part.data.sessionID);
      if (sessionID.success) return sessionID.data;
    }
  }

  return null;
}

function lastUserText(messages: ManagedRAGUIMessage[]): string {
  for (let index = messages.length - 1; index >= 0; index -= 1) {
    const message = messages[index];
    if (message?.role !== "user") continue;

    return message.parts
      .filter((part) => part.type === "text")
      .map((part) => part.text.trim())
      .filter(Boolean)
      .join("\n");
  }

  return "";
}

function parseMessages(value: unknown): ManagedRAGUIMessage[] | null {
  if (!Array.isArray(value) || value.length > 100) return null;

  for (const message of value) {
    if (!isRecord(message) || typeof message.id !== "string" ||
      !["system", "user", "assistant"].includes(String(message.role)) || !Array.isArray(message.parts)) {
      return null;
    }

    for (const part of message.parts) {
      if (!isRecord(part) || typeof part.type !== "string") return null;
      if (part.type === "text" && typeof part.text !== "string") return null;
    }
  }

  return value as ManagedRAGUIMessage[];
}

function isRecord(value: unknown): value is Record<string, unknown> {
  return typeof value === "object" && value !== null;
}
