export const ADMIN_SESSION_COOKIE = 'support_admin_session';
export const SESSION_TTL_SECONDS = 60 * 60 * 12;
export const APP_BASE_PATH = '/catalog';
export const ADMIN_COOKIE_PATH = APP_BASE_PATH;

export function withBasePath(path: string): string {
  const normalizedPath = path.startsWith('/') ? path : `/${path}`;
  return `${APP_BASE_PATH}${normalizedPath}`;
}
