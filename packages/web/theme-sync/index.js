export const SHARED_THEME_STORAGE_KEY = "godarai-theme";
export const SHARED_THEME_COOKIE_NAME = "godarai-theme";
export const SHARED_THEME_CHANGE_EVENT = "godarai-theme-change";

const COOKIE_MAX_AGE_SECONDS = 60 * 60 * 24 * 365;
const PRODUCTION_COOKIE_DOMAIN = ".godarai.ir";

export function normalizeThemePreference(value) {
  if (typeof value !== "string") return undefined;

  const normalized = value.toLowerCase();
  return normalized === "light" || normalized === "dark" || normalized === "system"
    ? normalized
    : undefined;
}

export function resolveThemePreference(preference, prefersDark = false) {
  const normalized = normalizeThemePreference(preference) ?? "system";
  return normalized === "system" ? (prefersDark ? "dark" : "light") : normalized;
}

export function readThemeCookie(cookieSource, cookieName = SHARED_THEME_COOKIE_NAME) {
  if (!cookieSource) return undefined;

  const encodedName = `${encodeURIComponent(cookieName)}=`;
  for (const part of cookieSource.split(";")) {
    const cookie = part.trim();
    if (!cookie.startsWith(encodedName)) continue;

    try {
      return normalizeThemePreference(decodeURIComponent(cookie.slice(encodedName.length)));
    } catch {
      return undefined;
    }
  }

  return undefined;
}

export function getThemeCookieDomain(hostname) {
  const normalizedHostname = hostname?.toLowerCase();
  if (
    normalizedHostname === "godarai.ir" ||
    normalizedHostname?.endsWith(".godarai.ir")
  ) {
    return PRODUCTION_COOKIE_DOMAIN;
  }

  return undefined;
}

export function selectThemePreference({
  cookieTheme,
  storedTheme,
  legacyThemes = [],
  defaultTheme = "system",
} = {}) {
  return (
    normalizeThemePreference(cookieTheme) ??
    normalizeThemePreference(storedTheme) ??
    legacyThemes.map(normalizeThemePreference).find(Boolean) ??
    normalizeThemePreference(defaultTheme) ??
    "system"
  );
}

function safelyReadStorage(storage, key) {
  try {
    return storage?.getItem(key) ?? undefined;
  } catch {
    return undefined;
  }
}

function safelyWriteStorage(storage, key, value) {
  try {
    storage?.setItem(key, value);
  } catch {
    // Theme persistence is best-effort when storage is blocked.
  }
}

function readLegacyThemes(storage, legacyStorageKeys, legacyJsonStorage) {
  const themes = legacyStorageKeys.map((key) => safelyReadStorage(storage, key));

  for (const entry of legacyJsonStorage) {
    const raw = safelyReadStorage(storage, entry.key);
    if (!raw) continue;

    try {
      const parsed = JSON.parse(raw);
      themes.push(parsed?.[entry.property]);
    } catch {
      // Ignore malformed legacy settings and continue to the next source.
    }
  }

  return themes;
}

export function getBrowserThemePreference({
  defaultTheme = "system",
  legacyStorageKeys = [],
  legacyJsonStorage = [],
} = {}) {
  if (typeof window === "undefined" || typeof document === "undefined") {
    return normalizeThemePreference(defaultTheme) ?? "system";
  }

  return selectThemePreference({
    cookieTheme: readThemeCookie(document.cookie),
    storedTheme: safelyReadStorage(window.localStorage, SHARED_THEME_STORAGE_KEY),
    legacyThemes: readLegacyThemes(
      window.localStorage,
      legacyStorageKeys,
      legacyJsonStorage,
    ),
    defaultTheme,
  });
}

function writeThemeCookie(preference) {
  if (typeof window === "undefined" || typeof document === "undefined") return;

  const domain = getThemeCookieDomain(window.location.hostname);
  const attributes = [
    `${encodeURIComponent(SHARED_THEME_COOKIE_NAME)}=${encodeURIComponent(preference)}`,
    "Path=/",
    `Max-Age=${COOKIE_MAX_AGE_SECONDS}`,
    "SameSite=Lax",
  ];

  if (domain) attributes.push(`Domain=${domain}`);
  if (window.location.protocol === "https:") attributes.push("Secure");

  document.cookie = attributes.join("; ");
}

export function setBrowserThemePreference(preference, {emit = true} = {}) {
  const normalized = normalizeThemePreference(preference) ?? "system";
  if (typeof window === "undefined" || typeof document === "undefined") {
    return normalized;
  }

  safelyWriteStorage(window.localStorage, SHARED_THEME_STORAGE_KEY, normalized);
  writeThemeCookie(normalized);

  if (emit) {
    window.dispatchEvent(
      new CustomEvent(SHARED_THEME_CHANGE_EVENT, {detail: {preference: normalized}}),
    );
  }

  return normalized;
}

export function migrateBrowserThemePreference(options = {}) {
  const preference = getBrowserThemePreference(options);
  setBrowserThemePreference(preference, {emit: false});
  return preference;
}

export function getSystemPrefersDark() {
  return (
    typeof window !== "undefined" &&
    window.matchMedia("(prefers-color-scheme: dark)").matches
  );
}

export function applyThemePreference(preference, targetDocument = document) {
  const normalized = normalizeThemePreference(preference) ?? "system";
  const resolved = resolveThemePreference(normalized, getSystemPrefersDark());
  const root = targetDocument.documentElement;

  root.classList.remove("light", "dark");
  root.classList.add(resolved);
  root.dataset.theme = resolved;
  root.dataset.themeChoice = normalized;
  root.style.colorScheme = resolved;

  return resolved;
}

export function subscribeToThemePreference(callback) {
  if (typeof window === "undefined") return () => {};

  const media = window.matchMedia("(prefers-color-scheme: dark)");
  const notify = () => callback();
  const notifyWhenVisible = () => {
    if (document.visibilityState === "visible") notify();
  };
  const onStorage = (event) => {
    if (event.key === null || event.key === SHARED_THEME_STORAGE_KEY) notify();
  };

  window.addEventListener("storage", onStorage);
  window.addEventListener(SHARED_THEME_CHANGE_EVENT, notify);
  window.addEventListener("focus", notify);
  window.addEventListener("pageshow", notify);
  document.addEventListener("visibilitychange", notifyWhenVisible);
  media.addEventListener("change", notify);

  return () => {
    window.removeEventListener("storage", onStorage);
    window.removeEventListener(SHARED_THEME_CHANGE_EVENT, notify);
    window.removeEventListener("focus", notify);
    window.removeEventListener("pageshow", notify);
    document.removeEventListener("visibilitychange", notifyWhenVisible);
    media.removeEventListener("change", notify);
  };
}

export function createThemeBootstrapScript({
  defaultTheme = "system",
  legacyStorageKeys = [],
  legacyJsonStorage = [],
  mirrorStorageKey,
} = {}) {
  const config = {
    cookieName: SHARED_THEME_COOKIE_NAME,
    storageKey: SHARED_THEME_STORAGE_KEY,
    cookieMaxAge: COOKIE_MAX_AGE_SECONDS,
    cookieDomain: PRODUCTION_COOKIE_DOMAIN,
    defaultTheme: normalizeThemePreference(defaultTheme) ?? "system",
    legacyStorageKeys,
    legacyJsonStorage,
    mirrorStorageKey,
  };

  return `;(${themeBootstrap.toString()})(${JSON.stringify(config)});`;
}

function themeBootstrap(config) {
  function normalize(value) {
    if (typeof value !== "string") return undefined;
    var normalized = value.toLowerCase();
    return normalized === "light" || normalized === "dark" || normalized === "system"
      ? normalized
      : undefined;
  }

  function readStorage(key) {
    try {
      return window.localStorage.getItem(key) || undefined;
    } catch (_error) {
      return undefined;
    }
  }

  function writeStorage(key, value) {
    try {
      if (value === undefined) window.localStorage.removeItem(key);
      else window.localStorage.setItem(key, value);
    } catch (_error) {
      // Storage can be unavailable in privacy-restricted contexts.
    }
  }

  function readCookie() {
    var prefix = encodeURIComponent(config.cookieName) + "=";
    var parts = document.cookie ? document.cookie.split(";") : [];
    for (var index = 0; index < parts.length; index += 1) {
      var cookie = parts[index].trim();
      if (!cookie.startsWith(prefix)) continue;
      try {
        return normalize(decodeURIComponent(cookie.slice(prefix.length)));
      } catch (_error) {
        return undefined;
      }
    }
    return undefined;
  }

  function readLegacyTheme() {
    for (var index = 0; index < config.legacyStorageKeys.length; index += 1) {
      var theme = normalize(readStorage(config.legacyStorageKeys[index]));
      if (theme) return theme;
    }

    for (var jsonIndex = 0; jsonIndex < config.legacyJsonStorage.length; jsonIndex += 1) {
      var entry = config.legacyJsonStorage[jsonIndex];
      var raw = readStorage(entry.key);
      if (!raw) continue;
      try {
        var parsed = JSON.parse(raw);
        var parsedTheme = normalize(parsed && parsed[entry.property]);
        if (parsedTheme) return parsedTheme;
      } catch (_error) {
        // Continue to the next legacy source.
      }
    }
    return undefined;
  }

  var preference =
    readCookie() ||
    normalize(readStorage(config.storageKey)) ||
    readLegacyTheme() ||
    config.defaultTheme;
  var prefersDark = window.matchMedia("(prefers-color-scheme: dark)").matches;
  var resolved = preference === "system" ? (prefersDark ? "dark" : "light") : preference;
  var root = document.documentElement;

  writeStorage(config.storageKey, preference);
  if (config.mirrorStorageKey) {
    writeStorage(
      config.mirrorStorageKey,
      preference === "system" ? undefined : preference,
    );
  }

  var hostname = window.location.hostname.toLowerCase();
  var attributes = [
    encodeURIComponent(config.cookieName) + "=" + encodeURIComponent(preference),
    "Path=/",
    "Max-Age=" + config.cookieMaxAge,
    "SameSite=Lax",
  ];
  if (hostname === "godarai.ir" || hostname.endsWith(".godarai.ir")) {
    attributes.push("Domain=" + config.cookieDomain);
  }
  if (window.location.protocol === "https:") attributes.push("Secure");
  document.cookie = attributes.join("; ");

  root.classList.remove("light", "dark");
  root.classList.add(resolved);
  root.setAttribute("data-theme", resolved);
  root.setAttribute("data-theme-choice", preference);
  root.style.colorScheme = resolved;
}
