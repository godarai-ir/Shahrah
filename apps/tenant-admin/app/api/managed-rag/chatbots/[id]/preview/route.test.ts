import { NextRequest } from "next/server";
import { beforeEach, describe, expect, it, vi } from "vitest";

import { POST } from "./route";

const readAuthFromRequestMock = vi.hoisted(() => vi.fn());
const fetchControlPlaneJSONMock = vi.hoisted(() => vi.fn());
const chatbotID = "11111111-1111-4111-8111-111111111111";
const sessionID = "33333333-3333-4333-8333-333333333333";

vi.mock("@/lib/auth/request-auth", () => ({
  readAuthFromRequest: readAuthFromRequestMock,
}));

vi.mock("@/lib/control-plane", () => ({
  fetchControlPlaneJSON: fetchControlPlaneJSONMock,
}));

describe("managed RAG AI SDK preview route", () => {
  beforeEach(() => {
    vi.clearAllMocks();
    readAuthFromRequestMock.mockResolvedValue({
      ok: true,
      tenantSlug: "tenant1",
      user: { id: "user-1" },
    });
    fetchControlPlaneJSONMock.mockResolvedValue({
      ok: true,
      payload: {
        request_id: "rag-request-1",
        trace_id: "rag-trace-1",
        session_id: sessionID,
        answer: "پاسخ مستند [1].",
        citations: [{
          id: "1",
          document_id: "22222222-2222-4222-8222-222222222222",
          title: "راهنما",
          page: 3,
          quote: "متن شاهد",
        }],
        refusal: false,
        actual_model: "model-1",
      },
    });
  });

  it("returns an atomic AI SDK data stream with validated citations", async () => {
    const response = await POST(buildRequest({
      id: "chat-1",
      messages: [{ id: "user-message-1", role: "user", parts: [{ type: "text", text: "پرسش من" }] }],
    }), routeContext());

    expect(response.status).toBe(200);
    expect(response.headers.get("content-type")).toContain("text/event-stream");
    expect(fetchControlPlaneJSONMock).toHaveBeenCalledWith(expect.objectContaining({
      tenantSlug: "tenant1",
      userID: "user-1",
      method: "POST",
      path: `/api/v1/chatbots/${chatbotID}/preview-responses`,
      body: { message: "پرسش من" },
    }));

    const stream = await response.text();
    expect(stream).toContain("پاسخ مستند [1].");
    expect(stream).toContain("data-result");
    expect(stream).toContain("راهنما");
  });

  it("continues the RAG session returned by the previous assistant response", async () => {
    const response = await POST(buildRequest({
      id: "chat-1",
      messages: [
        { id: "user-message-1", role: "user", parts: [{ type: "text", text: "پرسش اول" }] },
        {
          id: "assistant-message-1",
          role: "assistant",
          parts: [{
            type: "data-result",
            data: {
              citations: [],
              refusal: false,
              actualModel: "model-1",
              requestID: "rag-request-1",
              traceID: "rag-trace-1",
              sessionID,
            },
          }],
        },
        { id: "user-message-2", role: "user", parts: [{ type: "text", text: "پرسش دوم" }] },
      ],
    }), routeContext());

    expect(response.status).toBe(200);
    expect(fetchControlPlaneJSONMock).toHaveBeenCalledWith(expect.objectContaining({
      path: `/api/v1/chatbots/${chatbotID}/preview-responses`,
      body: { message: "پرسش دوم", session_id: sessionID },
    }));
  });

  it("rejects malformed UI messages before calling the RAG service", async () => {
    const response = await POST(buildRequest({
      messages: [{ id: "user-message-1", role: "user", parts: [{ type: "text", text: 42 }] }],
    }), routeContext());

    expect(response.status).toBe(400);
    expect(fetchControlPlaneJSONMock).not.toHaveBeenCalled();
  });

  it("rejects oversized chat input before calling the RAG service", async () => {
    const response = await POST(buildRequest({
      messages: [{ id: "user-message-1", role: "user", parts: [{ type: "text", text: "x".repeat(8001) }] }],
    }), routeContext());

    expect(response.status).toBe(400);
    expect(fetchControlPlaneJSONMock).not.toHaveBeenCalled();
  });
});

function buildRequest(body: unknown) {
  return new NextRequest("https://tenant1.example.com/api/managed-rag/chatbots/bot/preview", {
    method: "POST",
    headers: { "content-type": "application/json" },
    body: JSON.stringify(body),
  });
}

function routeContext() {
  return { params: Promise.resolve({ id: chatbotID }) };
}
