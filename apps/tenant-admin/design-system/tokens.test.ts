import { readFileSync } from "node:fs";

import { describe, expect, it } from "vitest";

type ThemeName = "light" | "dark";
type RGB = readonly [number, number, number];
type Color = { rgb: RGB; alpha: number };
type Oklab = readonly [number, number, number];

const tokensCSS = readFileSync(new URL("./tokens.css", import.meta.url), "utf8");
const globalsCSS = readFileSync(new URL("../app/globals.css", import.meta.url), "utf8");
const lightTokens = parseDeclarations(extractBlocks(tokensCSS, ":root").join("\n"));
const darkTokens = { ...lightTokens, ...parseDeclarations(extractBlocks(tokensCSS, ".dark").join("\n")) };

const TEXT_CONTRAST = 4.5;
const NON_TEXT_CONTRAST = 3;
const statusNames = ["destructive", "success", "warning", "info"] as const;

describe.each<ThemeName>(["light", "dark"])("%s design-token contrast", (theme) => {
  it("keeps solid semantic foreground pairs at WCAG AA contrast", () => {
    const semanticPairs = [
      ["foreground", "background"],
      ["card-foreground", "card"],
      ["popover-foreground", "popover"],
      ["primary-foreground", "primary"],
      ["secondary-foreground", "secondary"],
      ...statusNames.map((status) => [`${status}-foreground`, status]),
    ] as const;

    for (const [foreground, surface] of semanticPairs) {
      expectContrast(theme, foreground, surface, TEXT_CONTRAST);
    }
  });

  it("keeps status text readable on page, card, badge, and alert surfaces", () => {
    for (const status of statusNames) {
      const text = tokenColor(theme, status);

      for (const surface of ["background", "card"] as const) {
        const base = tokenColor(theme, surface);
        expect(contrast(text, base), `${status} text on ${surface}`).toBeGreaterThanOrEqual(TEXT_CONTRAST);

        const badgeSurface = composite({ ...text, alpha: 0.1 }, base);
        expect(contrast(text, badgeSurface), `${status} badge on ${surface}`).toBeGreaterThanOrEqual(TEXT_CONTRAST);
      }

      if (status === "destructive") {
        for (const surface of ["background", "card"] as const) {
          const base = tokenColor(theme, surface);
          const alertSurface = composite({ ...text, alpha: 0.05 }, base);
          expect(contrast(text, alertSurface), `${status} alert on ${surface}`).toBeGreaterThanOrEqual(TEXT_CONTRAST);
        }
      } else {
        expectContrast(theme, status, `${status}-muted`, TEXT_CONTRAST);
      }

      for (const surface of ["background", "card"] as const) {
        expectContrast(theme, `${status}-soft-foreground`, `${status}-soft`, TEXT_CONTRAST, surface);
      }
    }
  });

  it("keeps at least one layer of the two-color focus indicator visible", () => {
    const ring = tokenColor(theme, "ring");
    const ringOffset = tokenColor(theme, "ring-offset");

    for (const surface of ["background", "card", "popover", "primary", "secondary", ...statusNames]) {
      const surfaceColor = tokenColor(theme, surface);
      const strongestLayer = Math.max(contrast(ring, surfaceColor), contrast(ringOffset, surfaceColor));
      expect(strongestLayer, `focus indicator on ${surface}`).toBeGreaterThanOrEqual(NON_TEXT_CONTRAST);
    }

    for (const surface of ["background", "card"] as const) {
      const surfaceColor = tokenColor(theme, surface);
      const translucentRing = composite({ ...ring, alpha: 0.5 }, surfaceColor);
      expect(contrast(translucentRing, surfaceColor), `50% focus ring on ${surface}`).toBeGreaterThanOrEqual(NON_TEXT_CONTRAST);
    }
  });

  it("keeps the branded sidebar text and selection pairs readable", () => {
    expectContrast(theme, "sidebar-foreground", "sidebar", TEXT_CONTRAST);
    expectContrast(theme, "sidebar-primary-foreground", "sidebar-primary", TEXT_CONTRAST);
    expectContrast(theme, "sidebar-accent-foreground", "sidebar-accent", TEXT_CONTRAST);
  });

  it("resolves five distinct brand-derived chart colors", () => {
    const resolved = Array.from({ length: 5 }, (_, index) => {
      const color = tokenColor(theme, `chart-${index + 1}`);
      return color.rgb.map((channel) => Math.round(channel * 255)).join(",");
    });

    expect(new Set(resolved).size).toBe(5);
  });
});

describe("GodarAI brand primitives", () => {
  it("matches the canonical web-assets palette", () => {
    expect(lightTokens).toMatchObject({
      "brand-cyan": "#00C9EB",
      "brand-cyan-bright": "#05D1EE",
      "brand-cyan-deep": "#00BCE1",
      "brand-navy": "#012651",
      "brand-navy-top": "#022A57",
      "brand-navy-bottom": "#011D40",
      "brand-dark-background": "#06182D",
      "brand-off-white": "#F6FBFF",
      "brand-ice": "#D9EEFF",
    });
  });
});

describe("global focus indicator", () => {
  it("renders a two-color two-pixel focus perimeter", () => {
    expect(lightTokens["focus-ring-width"]).toBe("2px");
    expect(globalsCSS).toMatch(/:focus-visible\s*{[\s\S]*?outline:[^;]*var\(--ring\)/);
    expect(globalsCSS).toMatch(/:focus-visible\s*{[\s\S]*?box-shadow:[^;]*var\(--ring-offset\)/);
  });
});

function expectContrast(
  theme: ThemeName,
  foreground: string,
  background: string,
  minimum: number,
  compositeSurface?: "background" | "card",
) {
  const foregroundColor = tokenColor(theme, foreground);
  let backgroundColor = tokenColor(theme, background);
  if (backgroundColor.alpha < 1) {
    backgroundColor = composite(backgroundColor, tokenColor(theme, compositeSurface ?? "background"));
  }
  expect(contrast(foregroundColor, backgroundColor), `${foreground} on ${background}`).toBeGreaterThanOrEqual(minimum);
}

function tokenColor(theme: ThemeName, name: string): Color {
  const tokens = theme === "dark" ? darkTokens : lightTokens;
  const value = tokens[name];
  if (!value) throw new Error(`Missing ${theme} token: --${name}`);
  return parseColor(value, tokens, new Set([name]));
}

function parseColor(value: string, tokens: Record<string, string>, seen: Set<string>): Color {
  const trimmed = value.trim();
  const variable = trimmed.match(/^var\(--([a-z0-9-]+)\)$/i);
  if (variable) {
    const name = variable[1];
    if (seen.has(name)) throw new Error(`Circular token reference: --${name}`);
    const resolved = tokens[name];
    if (!resolved) throw new Error(`Unknown token reference: --${name}`);
    return parseColor(resolved, tokens, new Set([...seen, name]));
  }

  if (trimmed === "white") return { rgb: [1, 1, 1], alpha: 1 };
  if (trimmed === "black") return { rgb: [0, 0, 0], alpha: 1 };
  if (trimmed === "transparent") return { rgb: [0, 0, 0], alpha: 0 };

  const hex = trimmed.match(/^#([0-9a-f]{6})$/i);
  if (hex) {
    return {
      rgb: [0, 2, 4].map((offset) => Number.parseInt(hex[1].slice(offset, offset + 2), 16) / 255) as unknown as RGB,
      alpha: 1,
    };
  }

  const oklch = trimmed.match(/^oklch\(\s*([\d.]+)\s+([\d.]+)\s+([\d.]+)\s*\)$/i);
  if (oklch) {
    const lightness = Number(oklch[1]);
    const chroma = Number(oklch[2]);
    const hue = Number(oklch[3]) * Math.PI / 180;
    return { rgb: oklabToSrgb([lightness, chroma * Math.cos(hue), chroma * Math.sin(hue)]), alpha: 1 };
  }

  const colorMix = trimmed.match(/^color-mix\(in oklab,\s*(.+?)\s+([\d.]+)%,\s*(.+?)\s*\)$/i);
  if (colorMix) {
    const first = parseColor(colorMix[1], tokens, seen);
    const weight = Number(colorMix[2]) / 100;
    const second = parseColor(colorMix[3], tokens, seen);

    if (second.alpha === 0) return { ...first, alpha: weight };

    const firstLab = srgbToOklab(first.rgb);
    const secondLab = srgbToOklab(second.rgb);
    return {
      rgb: oklabToSrgb([
        firstLab[0] * weight + secondLab[0] * (1 - weight),
        firstLab[1] * weight + secondLab[1] * (1 - weight),
        firstLab[2] * weight + secondLab[2] * (1 - weight),
      ]),
      alpha: 1,
    };
  }

  throw new Error(`Unsupported color value in contrast test: ${trimmed}`);
}

function contrast(first: Color, second: Color): number {
  const firstLuminance = relativeLuminance(first.rgb);
  const secondLuminance = relativeLuminance(second.rgb);
  return (Math.max(firstLuminance, secondLuminance) + 0.05) / (Math.min(firstLuminance, secondLuminance) + 0.05);
}

function composite(foreground: Color, background: Color): Color {
  const alpha = foreground.alpha + background.alpha * (1 - foreground.alpha);
  const rgb = foreground.rgb.map((channel, index) => (
    (channel * foreground.alpha + background.rgb[index] * background.alpha * (1 - foreground.alpha)) / alpha
  )) as unknown as RGB;
  return { rgb, alpha };
}

function relativeLuminance(rgb: RGB): number {
  const [red, green, blue] = rgb.map((channel) => (
    channel <= 0.04045 ? channel / 12.92 : ((channel + 0.055) / 1.055) ** 2.4
  ));
  return 0.2126 * red + 0.7152 * green + 0.0722 * blue;
}

function oklabToSrgb([lightness, axisA, axisB]: Oklab): RGB {
  const long = lightness + 0.3963377774 * axisA + 0.2158037573 * axisB;
  const medium = lightness - 0.1055613458 * axisA - 0.0638541728 * axisB;
  const short = lightness - 0.0894841775 * axisA - 1.291485548 * axisB;
  const longCubed = long ** 3;
  const mediumCubed = medium ** 3;
  const shortCubed = short ** 3;
  const linear: RGB = [
    4.0767416621 * longCubed - 3.3077115913 * mediumCubed + 0.2309699292 * shortCubed,
    -1.2684380046 * longCubed + 2.6097574011 * mediumCubed - 0.3413193965 * shortCubed,
    -0.0041960863 * longCubed - 0.7034186147 * mediumCubed + 1.707614701 * shortCubed,
  ];
  return linear.map((channel) => {
    const clipped = Math.max(0, Math.min(1, channel));
    return clipped <= 0.0031308 ? 12.92 * clipped : 1.055 * clipped ** (1 / 2.4) - 0.055;
  }) as unknown as RGB;
}

function srgbToOklab(rgb: RGB): Oklab {
  const [red, green, blue] = rgb.map((channel) => (
    channel <= 0.04045 ? channel / 12.92 : ((channel + 0.055) / 1.055) ** 2.4
  ));
  const long = Math.cbrt(0.4122214708 * red + 0.5363325363 * green + 0.0514459929 * blue);
  const medium = Math.cbrt(0.2119034982 * red + 0.6806995451 * green + 0.1073969566 * blue);
  const short = Math.cbrt(0.0883024619 * red + 0.2817188376 * green + 0.6299787005 * blue);
  return [
    0.2104542553 * long + 0.793617785 * medium - 0.0040720468 * short,
    1.9779984951 * long - 2.428592205 * medium + 0.4505937099 * short,
    0.0259040371 * long + 0.7827717662 * medium - 0.808675766 * short,
  ];
}

function extractBlocks(css: string, selector: string): string[] {
  const escapedSelector = selector.replace(/[.*+?^${}()|[\]\\]/g, "\\$&");
  const matches = [...css.matchAll(new RegExp(`${escapedSelector}\\s*{([\\s\\S]*?)}`, "g"))];
  if (matches.length === 0) throw new Error(`Missing CSS block: ${selector}`);
  return matches.map((match) => match[1]);
}

function parseDeclarations(block: string): Record<string, string> {
  return Object.fromEntries(
    [...block.matchAll(/--([a-z0-9-]+):\s*([^;]+);/gi)].map((match) => [match[1], match[2].trim()]),
  );
}
