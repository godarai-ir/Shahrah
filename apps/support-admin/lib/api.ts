export const API_BASE =
  process.env.MODEL_DISCOVERY_API_BASE_INTERNAL || 'http://localhost:3002/api/v1';

export function resolveAssetUrl(path?: string | null): string | null {
  if (!path) return null;
  if (path.startsWith('http://') || path.startsWith('https://')) return path;
  const normalized = path.startsWith('/') ? path : `/${path}`;
  const discoveryPath = normalized.startsWith('/api/v1/')
    ? normalized.slice('/api/v1/'.length)
    : normalized.slice(1);
  return `/api/admin/model-discovery-assets/${discoveryPath}`;
}
