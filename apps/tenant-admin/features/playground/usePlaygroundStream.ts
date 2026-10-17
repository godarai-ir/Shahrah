"use client";

import { useCallback } from "react";

export type PlaygroundStreamSnapshot = {
  responseText: string;
  toolCalls: unknown[];
  finishReason?: string;
  usage?: Record<string, unknown>;
  responseID?: string;
  responseModel?: string;
  firstTokenAt?: number;
  costUSD?: number;
  credits?: string;
};

type FrameScheduler = (callback: () => void) => () => void;

type ConsumePlaygroundStreamOptions = {
  signal: AbortSignal;
  disconnectedMessage: string;
  onUpdate: (snapshot: PlaygroundStreamSnapshot) => void;
  scheduleFrame?: FrameScheduler;
  now?: () => number;
};

const MAX_FINISH_FRAMES = 30;

export function usePlaygroundStream() {
  return useCallback((response: Response, options: ConsumePlaygroundStreamOptions) => consumePlaygroundStream(response, options), []);
}

export async function consumePlaygroundStream(response: Response, options: ConsumePlaygroundStreamOptions): Promise<PlaygroundStreamSnapshot> {
  if (options.signal.aborted) throw new DOMException("The operation was aborted.", "AbortError");
  const reader = response.body?.getReader();
  if (!reader) throw new Error(options.disconnectedMessage);

  const snapshot: PlaygroundStreamSnapshot = { responseText: "", toolCalls: [] };
  const now = options.now ?? (() => performance.now());
  const queue = createDeltaRenderQueue((delta) => {
    snapshot.responseText += delta;
    options.onUpdate(cloneSnapshot(snapshot));
  }, options.scheduleFrame ?? scheduleBrowserFrame, options.signal);
  const decoder = new TextDecoder();
  let buffer = "";
  let receivedDone = false;

  try {
    while (true) {
      const { done, value } = await reader.read();
      if (done) break;
      buffer += decoder.decode(value, { stream: true });
      const frames = buffer.split(/\r?\n\r?\n/);
      buffer = frames.pop() ?? "";

      for (const frame of frames) {
        const data = frame.split(/\r?\n/).filter((line) => line.startsWith("data:")).map((line) => line.slice(5).trimStart()).join("\n");
        if (!data) continue;
        if (data.trim() === "[DONE]") {
          receivedDone = true;
          continue;
        }

        const chunk = JSON.parse(data) as Record<string, unknown>;
        const streamError = extractStreamError(chunk, options.disconnectedMessage);
        if (streamError) throw new Error(streamError);
        snapshot.responseID = typeof chunk.id === "string" ? chunk.id : snapshot.responseID;
        snapshot.responseModel = typeof chunk.model === "string" ? chunk.model : snapshot.responseModel;
        if (chunk.usage && typeof chunk.usage === "object") snapshot.usage = chunk.usage as Record<string, unknown>;
        if (chunk.object === "godarai.telemetry") {
          snapshot.costUSD = finiteNumber(chunk.cost_usd) ?? snapshot.costUSD;
          snapshot.credits = typeof chunk.credits === "string" && chunk.credits ? chunk.credits : snapshot.credits;
        }

        const choice = Array.isArray(chunk.choices) ? chunk.choices[0] as Record<string, unknown> | undefined : undefined;
        const delta = choice?.delta && typeof choice.delta === "object" ? choice.delta as Record<string, unknown> : undefined;
        const text = extractPlaygroundText(delta?.content);
        if (text) {
          if (snapshot.firstTokenAt === undefined) snapshot.firstTokenAt = now();
          queue.enqueue(text);
        }
        if (Array.isArray(delta?.tool_calls)) {
          snapshot.toolCalls = mergeToolCallDeltas(snapshot.toolCalls, delta.tool_calls);
          options.onUpdate(cloneSnapshot(snapshot));
        }
        if (typeof choice?.finish_reason === "string") snapshot.finishReason = choice.finish_reason;
        if (!text && (chunk.usage || choice?.finish_reason || chunk.object === "godarai.telemetry")) options.onUpdate(cloneSnapshot(snapshot));
      }
    }

    if (!receivedDone) throw new Error(options.disconnectedMessage);
    await queue.finish();
    if (options.signal.aborted) throw new DOMException("The operation was aborted.", "AbortError");

    return cloneSnapshot(snapshot);
  } catch (error) {
    queue.cancel();
    throw error;
  } finally {
    reader.releaseLock();
  }
}

function extractStreamError(chunk: Record<string, unknown>, disconnectedMessage: string): string | undefined {
  if (!chunk.error || typeof chunk.error !== "object" || Array.isArray(chunk.error)) return undefined;
  const error = chunk.error as Record<string, unknown>;
  const code = typeof error.code === "string" ? error.code : "";
  if (code === "stream_incomplete" || code === "stream_interrupted") return disconnectedMessage;
  return typeof error.message === "string" && error.message.trim() ? error.message : disconnectedMessage;
}

export function createDeltaRenderQueue(onAppend: (delta: string) => void, scheduleFrame: FrameScheduler = scheduleBrowserFrame, signal?: AbortSignal) {
  const pending: string[] = [];
  let cancelFrame: (() => void) | null = null;
  let finishing = false;
  let framesRemaining = 0;
  let cancelled = false;
  let resolveFinish: (() => void) | null = null;

  const resolveIfFinished = () => {
    if (finishing && pending.length === 0) {
      resolveFinish?.();
      resolveFinish = null;
    }
  };

  const schedule = () => {
    if (cancelled || cancelFrame || pending.length === 0) return;
    cancelFrame = scheduleFrame(() => {
      cancelFrame = null;
      if (cancelled) return;

      const count = finishing
        ? Math.max(1, Math.ceil(pending.length / Math.max(1, framesRemaining)))
        : 1;
      const delta = pending.splice(0, count).join("");
      if (framesRemaining > 0) framesRemaining -= 1;
      if (delta) onAppend(delta);

      if (pending.length > 0) schedule();
      else resolveIfFinished();
    });
  };

  const cancel = () => {
    if (cancelled) return;
    cancelled = true;
    cancelFrame?.();
    cancelFrame = null;
    pending.length = 0;
    resolveFinish?.();
    resolveFinish = null;
  };

  signal?.addEventListener("abort", cancel, { once: true });

  return {
    enqueue(delta: string) {
      if (!delta || cancelled || finishing) return;
      pending.push(delta);
      schedule();
    },
    finish() {
      if (cancelled || pending.length === 0) return Promise.resolve();
      finishing = true;
      framesRemaining = Math.min(MAX_FINISH_FRAMES, pending.length);
      schedule();
      return new Promise<void>((resolve) => { resolveFinish = resolve; });
    },
    cancel,
  };
}

export function extractPlaygroundText(content: unknown): string {
  if (typeof content === "string") return content;
  if (Array.isArray(content)) return content.map((part) => part && typeof part === "object" && "text" in part ? String((part as { text?: unknown }).text ?? "") : "").join("");
  return content == null ? "" : JSON.stringify(content);
}

function mergeToolCallDeltas(current: unknown[], deltas: unknown[]): unknown[] {
  const next = current.map((item) => structuredClone(item)) as Array<Record<string, unknown>>;
  for (const rawDelta of deltas) {
    if (!rawDelta || typeof rawDelta !== "object") continue;
    const delta = rawDelta as Record<string, unknown>;
    const index = typeof delta.index === "number" ? delta.index : next.length;
    const existing = next[index] ?? { index };
    if (typeof delta.id === "string") existing.id = delta.id;
    if (typeof delta.type === "string") existing.type = delta.type;
    if (delta.function && typeof delta.function === "object") {
      const functionDelta = delta.function as Record<string, unknown>;
      const oldFunction = existing.function && typeof existing.function === "object" ? existing.function as Record<string, unknown> : {};
      existing.function = {
        name: typeof functionDelta.name === "string" ? `${oldFunction.name ?? ""}${functionDelta.name}` : oldFunction.name,
        arguments: typeof functionDelta.arguments === "string" ? `${oldFunction.arguments ?? ""}${functionDelta.arguments}` : oldFunction.arguments,
      };
    }
    next[index] = existing;
  }
  return next;
}

function cloneSnapshot(snapshot: PlaygroundStreamSnapshot): PlaygroundStreamSnapshot {
  return { ...snapshot, toolCalls: structuredClone(snapshot.toolCalls), usage: snapshot.usage ? { ...snapshot.usage } : undefined };
}

function finiteNumber(value: unknown): number | undefined {
  return typeof value === "number" && Number.isFinite(value) ? value : undefined;
}

function scheduleBrowserFrame(callback: () => void) {
  if (typeof window === "undefined" || typeof window.requestAnimationFrame !== "function") {
    const timeout = setTimeout(callback, 16);
    return () => clearTimeout(timeout);
  }

  let pending = true;
  let frame = 0;
  let timeout = 0;
  const run = () => {
    if (!pending) return;
    pending = false;
    window.cancelAnimationFrame(frame);
    window.clearTimeout(timeout);
    callback();
  };
  frame = window.requestAnimationFrame(run);
  timeout = window.setTimeout(run, 50);

  return () => {
    pending = false;
    window.cancelAnimationFrame(frame);
    window.clearTimeout(timeout);
  };
}
