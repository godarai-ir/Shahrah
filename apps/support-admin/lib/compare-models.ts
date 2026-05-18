export const COMPARE_MODELS_STORAGE_KEY = 'support_admin_compare_models_v1';
export const COMPARE_MODELS_CHANGED_EVENT = 'support-admin:compare-models-changed';
export const MAX_COMPARE_MODELS = 4;
const EMPTY_COMPARE_IDS: CompareModelId[] = [];
let cachedRawCompareIds: string | null = null;
let cachedCompareIds: CompareModelId[] = EMPTY_COMPARE_IDS;

export type CompareModelId = string;

export type ComparePricing = {
  inputPricePerMillionTokens?: number | null;
  outputPricePerMillionTokens?: number | null;
  cachedInputPricePerMillionTokens?: number | null;
};

export type CompareModel = {
  canonicalId: string;
  displayName?: string | null;
  status?: string | null;
  releaseDate?: string | null;
  contextWindow?: number | null;
  maxOutputTokens?: number | null;
  openSource?: boolean | null;
  free?: boolean | null;
  license?: string | null;
  officialDocsUrl?: string | null;
  provider?: {
    slug?: string | null;
    name?: string | null;
  } | null;
  modelSeries?: {
    slug?: string | null;
    name?: string | null;
  } | null;
  pricing?: ComparePricing[] | null;
  inputModalities?: Array<{ modality?: { slug?: string | null; name?: string | null } | null }> | null;
  outputModalities?: Array<{ modality?: { slug?: string | null; name?: string | null } | null }> | null;
  capabilities?: Array<{ capability?: { slug?: string | null; name?: string | null } | null }> | null;
  supportedParams?: Array<{ supportedParameter?: { slug?: string | null; name?: string | null } | null }> | null;
  regions?: Array<{ region?: { slug?: string | null; name?: string | null } | null }> | null;
  benchmarks?: Array<{
    score?: number | null;
    percentile?: number | null;
    benchmark?: { slug?: string | null; name?: string | null } | null;
  }> | null;
};

export function parseCompareIds(input: string | null | undefined): CompareModelId[] {
  if (!input) return [];
  const seen = new Set<string>();
  const ids: string[] = [];
  for (const raw of input.split(',')) {
    const id = raw.trim();
    if (!id || seen.has(id)) continue;
    seen.add(id);
    ids.push(id);
    if (ids.length >= MAX_COMPARE_MODELS) break;
  }
  return ids;
}

export function serializeCompareIds(ids: CompareModelId[]): string {
  return ids.join(',');
}

export function dedupeCompareIds(ids: CompareModelId[]): CompareModelId[] {
  return Array.from(new Set(ids.filter(Boolean)));
}

export function readStoredCompareIds(): CompareModelId[] {
  if (typeof window === 'undefined') return [];
  try {
    const raw = window.localStorage.getItem(COMPARE_MODELS_STORAGE_KEY);
    if (!raw) return [];
    const parsed = JSON.parse(raw);
    if (!Array.isArray(parsed)) return [];
    return dedupeCompareIds(parsed.map((id) => String(id))).slice(0, MAX_COMPARE_MODELS);
  } catch {
    return [];
  }
}

export function getStoredCompareIdsSnapshot(): CompareModelId[] {
  if (typeof window === 'undefined') return EMPTY_COMPARE_IDS;
  const raw = window.localStorage.getItem(COMPARE_MODELS_STORAGE_KEY);
  if (raw === cachedRawCompareIds) return cachedCompareIds;
  cachedRawCompareIds = raw;
  cachedCompareIds = readStoredCompareIds();
  return cachedCompareIds;
}

export function getStoredCompareIdsServerSnapshot(): CompareModelId[] {
  return EMPTY_COMPARE_IDS;
}

export function writeStoredCompareIds(ids: CompareModelId[]): void {
  if (typeof window === 'undefined') return;
  window.localStorage.setItem(COMPARE_MODELS_STORAGE_KEY, JSON.stringify(ids.slice(0, MAX_COMPARE_MODELS)));
  window.dispatchEvent(new Event(COMPARE_MODELS_CHANGED_EVENT));
}

export function subscribeStoredCompareIds(onStoreChange: () => void): () => void {
  if (typeof window === 'undefined') return () => {};
  const handleChange = () => onStoreChange();
  window.addEventListener('storage', handleChange);
  window.addEventListener(COMPARE_MODELS_CHANGED_EVENT, handleChange);
  return () => {
    window.removeEventListener('storage', handleChange);
    window.removeEventListener(COMPARE_MODELS_CHANGED_EVENT, handleChange);
  };
}

export function normalizeNameList(entries: Array<{ slug?: string | null; name?: string | null } | null | undefined> | null | undefined): string[] {
  return (entries ?? [])
    .map((entry) => entry?.name || entry?.slug)
    .filter((value): value is string => Boolean(value));
}
