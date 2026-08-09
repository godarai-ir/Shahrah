export type ModelCatalogTab = "providers" | "models";
export type ModelCatalogSort =
  | "name"
  | "provider"
  | "availability"
  | "context"
  | "input_price"
  | "output_price";
export type SortOrder = "asc" | "desc";
export type ProviderDetailSection = "overview" | "models" | "access" | "credentials";

export type ModelCatalogURLState = {
  tab: ModelCatalogTab;
  q: string;
  providers: string[];
  accounts: string[];
  credential: string;
  readiness: string;
  sort: ModelCatalogSort;
  order: SortOrder;
  account: string;
  section: ProviderDetailSection;
  modelProvider: string;
  model: string;
  modelAccount: string;
};

const MODEL_CATALOG_SORTS = new Set<ModelCatalogSort>([
  "name",
  "provider",
  "availability",
  "context",
  "input_price",
  "output_price",
]);
const DETAIL_SECTIONS = new Set<ProviderDetailSection>([
  "overview",
  "models",
  "access",
  "credentials",
]);

export function parseModelCatalogURLState(params: URLSearchParams): ModelCatalogURLState {
  const rawTab = params.get("tab");
  const rawSort = params.get("sort") as ModelCatalogSort | null;
  const rawSection = params.get("section") as ProviderDetailSection | null;

  return {
    tab: rawTab === "models" ? "models" : "providers",
    q: params.get("q")?.trim() ?? "",
    providers: parseCSV(params.get("providers")),
    accounts: parseCSV(params.get("accounts")),
    credential: params.get("credential")?.trim() ?? "",
    readiness: params.get("readiness")?.trim() ?? "",
    sort: rawSort && MODEL_CATALOG_SORTS.has(rawSort) ? rawSort : "name",
    order: params.get("order") === "desc" ? "desc" : "asc",
    account: params.get("account")?.trim() ?? "",
    section: rawSection && DETAIL_SECTIONS.has(rawSection) ? rawSection : "overview",
    modelProvider: params.get("model_provider")?.trim() ?? "",
    model: params.get("model")?.trim() ?? "",
    modelAccount: params.get("model_account")?.trim() ?? "",
  };
}

export function patchModelCatalogURL(
  current: URLSearchParams,
  patch: Record<string, string | string[] | null | undefined>,
): string {
  const next = new URLSearchParams(current.toString());
  for (const [key, value] of Object.entries(patch)) {
    const serialized = Array.isArray(value) ? normalizeList(value).join(",") : value?.trim();
    if (serialized) next.set(key, serialized);
    else next.delete(key);
  }
  return next.toString();
}

export function parseCSV(value: string | null): string[] {
  return normalizeList(value?.split(",") ?? []);
}

function normalizeList(values: string[]): string[] {
  return [...new Set(values.map((value) => value.trim()).filter(Boolean))];
}
