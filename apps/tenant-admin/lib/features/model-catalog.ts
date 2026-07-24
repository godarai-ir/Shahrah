import "server-only";

export function isModelCatalogV2Enabled(): boolean {
  return process.env.MODEL_CATALOG_V2_ENABLED?.trim().toLowerCase() === "true";
}
