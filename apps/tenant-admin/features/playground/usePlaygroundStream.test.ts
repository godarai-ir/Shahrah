import { describe, expect, it, vi } from "vitest";

import { consumePlaygroundStream, createDeltaRenderQueue } from "./usePlaygroundStream";

describe("playground stream rendering", () => {
  it("renders coalesced deltas over separate frames", async () => {
    const callbacks: Array<() => void> = [];
    const appended: string[] = [];
    const queue = createDeltaRenderQueue(
      (delta) => appended.push(delta),
      (callback) => {
        callbacks.push(callback);
        return () => undefined;
      },
    );

    queue.enqueue("one");
    queue.enqueue(" two");
    queue.enqueue(" three");
    const finished = queue.finish();

    expect(appended).toEqual([]);
    callbacks.shift()?.();
    expect(appended).toEqual(["one"]);
    callbacks.shift()?.();
    expect(appended).toEqual(["one", " two"]);
    callbacks.shift()?.();
    await finished;
    expect(appended).toEqual(["one", " two", " three"]);
  });

  it("parses OpenAI SSE metadata while publishing text updates", async () => {
    const encoder = new TextEncoder();
    const body = [
      ": gateway-ready",
      'data: {"id":"response-1","model":"model-1","choices":[{"delta":{"content":"Hello"},"finish_reason":null}]}',
      ": keep-alive",
      'data: {"id":"response-1","model":"model-1","choices":[{"delta":{"content":" world"},"finish_reason":"stop"}]}',
      'event: godarai.telemetry\ndata: {"object":"godarai.telemetry","usage":{"prompt_tokens":12,"completion_tokens":7,"total_tokens":19},"cost_usd":0.00042,"credits":"0.42"}',
      "data: [DONE]",
      "",
    ].join("\n\n");
    const response = new Response(new ReadableStream({
      start(controller) {
        controller.enqueue(encoder.encode(body));
        controller.close();
      },
    }), { headers: { "Content-Type": "text/event-stream" } });
    const updates: string[] = [];

    const result = await consumePlaygroundStream(response, {
      signal: new AbortController().signal,
      disconnectedMessage: "disconnected",
      onUpdate: (snapshot) => updates.push(snapshot.responseText),
      scheduleFrame: (callback) => {
        queueMicrotask(callback);
        return () => undefined;
      },
      now: vi.fn(() => 42),
    });

    expect(updates).toContain("Hello");
    expect(updates.at(-1)).toBe("Hello world");
    expect(result).toMatchObject({
      responseText: "Hello world",
      responseID: "response-1",
      responseModel: "model-1",
      finishReason: "stop",
      firstTokenAt: 42,
      usage: { prompt_tokens: 12, completion_tokens: 7, total_tokens: 19 },
      costUSD: 0.00042,
      credits: "0.42",
    });
  });

  it("surfaces an interrupted stream event with the localized disconnect message", async () => {
    const encoder = new TextEncoder();
    const response = new Response(new ReadableStream({
      start(controller) {
        controller.enqueue(encoder.encode([
          'data: {"choices":[{"delta":{"content":"partial"}}]}',
          'data: {"error":{"message":"The upstream response stream was interrupted.","type":"streaming_error","code":"stream_interrupted"}}',
          "",
        ].join("\n\n")));
        controller.close();
      },
    }), { headers: { "Content-Type": "text/event-stream" } });

    await expect(consumePlaygroundStream(response, {
      signal: new AbortController().signal,
      disconnectedMessage: "localized disconnect",
      onUpdate: () => undefined,
      scheduleFrame: (callback) => {
        queueMicrotask(callback);
        return () => undefined;
      },
    })).rejects.toThrow("localized disconnect");
  });
});
