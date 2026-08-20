export type PlaygroundMode = "chat" | "image_generation" | "layout_parsing";

export function normalizePlaygroundMode(value: unknown): PlaygroundMode | null {
  if (typeof value !== "string") return null;
  const normalized = value.trim().toLowerCase().replaceAll("-", "_").replaceAll(" ", "_");
  switch (normalized) {
    case "chat":
    case "chat_completion":
    case "chat_completions":
      return "chat";
    case "image":
    case "images":
    case "image_generation":
    case "image_generation_api":
    case "text_to_image":
    case "image_edit":
      return "image_generation";
    case "ocr":
    case "layout":
    case "layout_parsing":
      return "layout_parsing";
    default:
      return null;
  }
}

export function modelModesSupportPlaygroundMode(values: readonly unknown[], mode: PlaygroundMode): boolean {
  if (values.length === 0) return mode === "chat";
  return values.some((value) => normalizePlaygroundMode(value) === mode);
}

export function buildPlaygroundHref(modelFqn: string, endpointMode: string): string | null {
  const mode = normalizePlaygroundMode(endpointMode);
  if (!mode) return null;
  return `/playground?${new URLSearchParams({ model: modelFqn, mode }).toString()}`;
}
