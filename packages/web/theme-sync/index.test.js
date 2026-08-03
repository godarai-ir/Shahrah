import assert from "node:assert/strict";
import test from "node:test";
import vm from "node:vm";

import {
  createThemeBootstrapScript,
  getThemeCookieDomain,
  normalizeThemePreference,
  readThemeCookie,
  resolveThemePreference,
  selectThemePreference,
} from "./index.js";

test("normalizes the shared light, dark, and system values", () => {
  assert.equal(normalizeThemePreference("DARK"), "dark");
  assert.equal(normalizeThemePreference("light"), "light");
  assert.equal(normalizeThemePreference("SYSTEM"), "system");
  assert.equal(normalizeThemePreference("sepia"), undefined);
});

test("uses the shared cookie before local and legacy values", () => {
  assert.equal(
    selectThemePreference({
      cookieTheme: "dark",
      storedTheme: "light",
      legacyThemes: ["system"],
    }),
    "dark",
  );
  assert.equal(
    selectThemePreference({legacyThemes: ["invalid", "DARK"]}),
    "dark",
  );
});

test("reads the shared theme cookie without matching similarly named cookies", () => {
  assert.equal(readThemeCookie("other=light; godarai-theme=system"), "system");
  assert.equal(readThemeCookie("godarai-theme-old=dark"), undefined);
});

test("shares cookies only across the production GodarAI domain", () => {
  assert.equal(getThemeCookieDomain("godarai.ir"), ".godarai.ir");
  assert.equal(getThemeCookieDomain("tenant.godarai.ir"), ".godarai.ir");
  assert.equal(getThemeCookieDomain("localhost"), undefined);
});

test("resolves system preference and emits a self-contained bootstrap", () => {
  assert.equal(resolveThemePreference("system", true), "dark");
  assert.equal(resolveThemePreference("system", false), "light");

  const script = createThemeBootstrapScript({
    legacyStorageKeys: ["shahrah-theme"],
  });
  assert.match(script, /shahrah-theme/);
  assert.match(script, /godarai-theme/);
  assert.doesNotMatch(script, /themeBootstrap is not defined/);
});

test("bootstrap migrates a legacy JSON preference before the page renders", () => {
  const values = new Map([
    [
      "model_discovery_web_display_settings_v1",
      JSON.stringify({themeMode: "DARK", currencyUnit: "IRR"}),
    ],
  ]);
  const classes = new Set();
  let writtenCookie = "";
  const document = {
    get cookie() {
      return "";
    },
    set cookie(value) {
      writtenCookie = value;
    },
    documentElement: {
      classList: {
        add: (...tokens) => tokens.forEach((token) => classes.add(token)),
        remove: (...tokens) => tokens.forEach((token) => classes.delete(token)),
      },
      setAttribute(name, value) {
        this[name] = value;
      },
      style: {},
    },
  };
  const window = {
    document,
    localStorage: {
      getItem: (key) => values.get(key) ?? null,
      setItem: (key, value) => values.set(key, value),
      removeItem: (key) => values.delete(key),
    },
    location: {hostname: "catalog.godarai.ir", protocol: "https:"},
    matchMedia: () => ({matches: false}),
  };

  vm.runInNewContext(
    createThemeBootstrapScript({
      legacyJsonStorage: [
        {key: "model_discovery_web_display_settings_v1", property: "themeMode"},
      ],
    }),
    {document, window, encodeURIComponent, decodeURIComponent, JSON},
  );

  assert.equal(values.get("godarai-theme"), "dark");
  assert.equal(classes.has("dark"), true);
  assert.equal(document.documentElement["data-theme-choice"], "dark");
  assert.match(writtenCookie, /^godarai-theme=dark;/);
  assert.match(writtenCookie, /Domain=.godarai.ir/);
  assert.match(writtenCookie, /Secure/);
});
