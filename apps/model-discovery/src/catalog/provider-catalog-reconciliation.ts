export function findMissingProviderSlugs(
  databaseProviderSlugs: string[],
  catalogProviderSlugs: string[],
): string[] {
  if (catalogProviderSlugs.length === 0) {
    throw new Error(
      'Refusing to reconcile missing providers against an empty provider catalog',
    );
  }

  const catalogSlugs = new Set(catalogProviderSlugs);
  return databaseProviderSlugs.filter((slug) => !catalogSlugs.has(slug)).sort();
}
