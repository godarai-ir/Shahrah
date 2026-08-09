export function formatCatalogModelDisplayName(model: string, displayName?: string): string {
  const source = displayName?.trim() || model;
  return source
    .replace(/\bGpt\b/gi, "GPT")
    .replace(/\bTts\b/gi, "TTS")
    .replace(/\bApi\b/gi, "API")
    .replace(/\bPdf\b/gi, "PDF")
    .replace(/\bAi\b/g, "AI");
}
