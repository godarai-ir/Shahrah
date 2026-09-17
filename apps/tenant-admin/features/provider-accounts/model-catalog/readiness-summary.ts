import type { CatalogModel, ModelAvailability } from "./types";

export type ModelReadinessSummary = "deprecated" | "pricing" | "limited" | "ready";

// Ordered by priority: the first matching state wins.
export const MODEL_READINESS_SUMMARIES: readonly ModelReadinessSummary[] = [
  "deprecated",
  "pricing",
  "limited",
  "ready",
];

export function isModelReadinessSummary(value: string): value is ModelReadinessSummary {
  return (MODEL_READINESS_SUMMARIES as readonly string[]).includes(value);
}

export function resolveModelAvailability(
  model: CatalogModel,
  accountID: string,
): ModelAvailability | null {
  return model.availability.find((item) => item.account_id === accountID) ?? model.availability[0] ?? null;
}

export function summarizeModelReadiness(
  model: CatalogModel,
  availability: ModelAvailability | null,
): ModelReadinessSummary {
  if (model.is_deprecated) return "deprecated";

  const capabilities = availability?.capability_readiness ?? model.readiness ?? [];
  if (availability?.pricing_resolved === false || capabilities.some((row) => row.state === "pricing_incomplete")) {
    return "pricing";
  }
  // blocked_lifecycle has no dedicated summary; it still means the model is not fully usable.
  if (capabilities.some((row) => row.state === "unsupported_transport" || row.state === "blocked_lifecycle")) {
    return "limited";
  }
  return "ready";
}

export function matchesModelQuery(model: CatalogModel, query: string): boolean {
  const needle = query.trim().toLowerCase();
  if (!needle) return true;
  return [
    model.model,
    model.display_name,
    model.replacement_model,
    ...model.types,
    ...(model.capabilities ?? []),
  ].some((value) => value?.toLowerCase().includes(needle));
}

export type SummarizedCatalogModel = {
  model: CatalogModel;
  availability: ModelAvailability | null;
  readiness: ModelReadinessSummary;
};

export function filterCatalogModels(
  models: CatalogModel[],
  accountID: string,
  query: string,
  readiness: ModelReadinessSummary | "",
): {
  rows: SummarizedCatalogModel[];
  counts: Record<ModelReadinessSummary, number>;
  queryMatches: number;
} {
  const counts: Record<ModelReadinessSummary, number> = { deprecated: 0, pricing: 0, limited: 0, ready: 0 };
  const rows: SummarizedCatalogModel[] = [];
  let queryMatches = 0;

  for (const model of models) {
    if (!matchesModelQuery(model, query)) continue;
    const availability = resolveModelAvailability(model, accountID);
    const summary = summarizeModelReadiness(model, availability);
    // Chip counts reflect the search, not the active chip, so switching chips never hides a count.
    counts[summary] += 1;
    queryMatches += 1;
    if (!readiness || readiness === summary) rows.push({ model, availability, readiness: summary });
  }

  return { rows, counts, queryMatches };
}
