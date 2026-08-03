export type ThemePreference = "light" | "dark" | "system";
export type ResolvedTheme = Exclude<ThemePreference, "system">;

export type LegacyJsonStorage = {
  key: string;
  property: string;
};

export type ThemeReadOptions = {
  defaultTheme?: ThemePreference;
  legacyStorageKeys?: string[];
  legacyJsonStorage?: LegacyJsonStorage[];
};

export const SHARED_THEME_STORAGE_KEY: "godarai-theme";
export const SHARED_THEME_COOKIE_NAME: "godarai-theme";
export const SHARED_THEME_CHANGE_EVENT: "godarai-theme-change";

export function normalizeThemePreference(value: unknown): ThemePreference | undefined;
export function resolveThemePreference(
  preference: ThemePreference,
  prefersDark?: boolean,
): ResolvedTheme;
export function readThemeCookie(
  cookieSource: string,
  cookieName?: string,
): ThemePreference | undefined;
export function getThemeCookieDomain(hostname?: string): string | undefined;
export function selectThemePreference(options?: {
  cookieTheme?: unknown;
  storedTheme?: unknown;
  legacyThemes?: unknown[];
  defaultTheme?: ThemePreference;
}): ThemePreference;
export function getBrowserThemePreference(options?: ThemeReadOptions): ThemePreference;
export function setBrowserThemePreference(
  preference: ThemePreference,
  options?: {emit?: boolean},
): ThemePreference;
export function migrateBrowserThemePreference(options?: ThemeReadOptions): ThemePreference;
export function getSystemPrefersDark(): boolean;
export function applyThemePreference(
  preference: ThemePreference,
  targetDocument?: Document,
): ResolvedTheme;
export function subscribeToThemePreference(callback: () => void): () => void;
export function createThemeBootstrapScript(
  options?: ThemeReadOptions & {mirrorStorageKey?: string},
): string;
