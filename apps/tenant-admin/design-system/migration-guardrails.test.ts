import { readdirSync, readFileSync } from "node:fs";
import { extname, join, relative } from "node:path";
import { fileURLToPath } from "node:url";

import { describe, expect, it } from "vitest";

const appRoot = fileURLToPath(new URL("..", import.meta.url));
const sourceRoots = ["app", "components", "features"];
const sourceExtensions = new Set([".css", ".ts", ".tsx"]);
const productionSources = sourceRoots.flatMap((root) => collectSourceFiles(join(appRoot, root)));

function collectSourceFiles(directory: string): string[] {
  return readdirSync(directory, { withFileTypes: true }).flatMap((entry) => {
    const path = join(directory, entry.name);
    if (entry.isDirectory()) return collectSourceFiles(path);
    if (!sourceExtensions.has(extname(entry.name)) || entry.name.includes(".test.")) return [];
    return [path];
  });
}

function findViolations(pattern: RegExp, excludedPaths: string[] = []) {
  return productionSources.flatMap((path) => {
    const relativePath = relative(appRoot, path).replaceAll("\\", "/");
    if (excludedPaths.includes(relativePath)) return [];
    return pattern.test(readFileSync(path, "utf8")) ? [relativePath] : [];
  });
}

describe("Design System migration guardrails", () => {
  it("rejects raw Tailwind palette colors", () => {
    const rawPalette = /\b(?:slate|gray|zinc|stone|neutral|red|green|blue|amber|yellow|orange|indigo|purple|violet|emerald|teal|cyan|sky|rose|pink|fuchsia|lime)-\d+/;
    expect(findViolations(rawPalette)).toEqual([]);
  });

  it("rejects arbitrary font sizes", () => {
    const arbitraryFontSize = /\btext-\[(?:\d+(?:\.\d+)?(?:px|rem)|clamp\()/;
    expect(findViolations(arbitraryFontSize)).toEqual([]);
  });

  it("rejects physical spacing and text-alignment utilities", () => {
    const physicalUtility = /\b(?:ml|mr|pl|pr)-[^\s"'`]+|\btext-(?:left|right)\b/;
    expect(findViolations(physicalUtility)).toEqual([]);
  });

  it("rejects fixed RTL outside the explicit playground example", () => {
    expect(findViolations(/dir=["']rtl["']/, ["app/design-system/page.tsx"])).toEqual([]);
  });

  it("routes product selects through AppSelectField", () => {
    const nativeSelect = /<\/?select\b/;
    expect(findViolations(nativeSelect)).toEqual([]);

    const directPrimitiveImport = /from ["']@\/components\/ui\/select["']/;
    expect(findViolations(directPrimitiveImport, ["components/app-ui/AppSelectField.tsx"])).toEqual([]);
  });
});
