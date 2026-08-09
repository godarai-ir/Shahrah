import { access, readFile, readdir } from "node:fs/promises";
import path from "node:path";
import process from "node:process";

const appRoot = path.resolve(import.meta.dirname, "..");
const sourceRoots = ["app", "components", "features", "hooks", "i18n", "lib"];
const sourceExtensions = new Set([".ts", ".tsx", ".mts"]);
const resolutionSuffixes = ["", ".ts", ".tsx", ".mts", "/index.ts", "/index.tsx", "/index.mts"];

const sourceFiles = (await Promise.all(sourceRoots.map((root) => walk(path.join(appRoot, root))))).flat();
const sourceSet = new Set(sourceFiles.map(normalize));
const sourceByFile = new Map(
  await Promise.all(sourceFiles.map(async (file) => [normalize(file), await readFile(file, "utf8")])),
);
const graph = new Map();
const violations = [];

for (const file of sourceFiles) {
  const normalizedFile = normalize(file);
  const source = sourceByFile.get(normalizedFile) ?? "";
  const imports = collectImports(source);
  const dependencies = [];

  validateFilePlacement(file, source);

  for (const specifier of imports) {
    const resolved = await resolveLocalImport(file, specifier);
    if (resolved) dependencies.push(resolved);
    validateLayerImport(file, specifier);

    if (isClientModule(source) && isServerImport(specifier, resolved)) {
      report(file, `client module imports server-only code through "${specifier}"`);
    }
  }

  graph.set(normalizedFile, dependencies);
}

function validateFilePlacement(file, source) {
  const filePath = relative(file);
  if (filePath.startsWith("app/components/")) {
    report(file, "feature implementation must live under features/<feature>, not app/components");
  }

  if (
    filePath.startsWith("features/")
    && /\bfetch\s*\(/u.test(source)
    && !/(^|\/)api(?:\.ts|\/)/u.test(filePath)
  ) {
    report(file, "direct fetch belongs in the feature API adapter");
  }

  if (
    filePath.startsWith("features/")
    && /(?:async\s+)?function\s+(?:requestJSON|fetchJSON)\s*</u.test(source)
  ) {
    report(file, "generic HTTP helpers belong in lib/http; expose named feature operations instead");
  }

  if (
    filePath.startsWith("features/")
    && /from\s+["']recharts["']/u.test(source)
    && !/charts?\.tsx$/iu.test(filePath)
  ) {
    report(file, "Recharts imports belong in a feature chart module behind a dynamic boundary");
  }
}

findCycles(graph);

if (violations.length > 0) {
  console.error("Architecture validation failed:\n");
  for (const violation of violations.sort()) console.error(`- ${violation}`);
  process.exitCode = 1;
} else {
  console.log(`Architecture validation passed for ${sourceFiles.length} source files.`);
}

function validateLayerImport(importer, specifier) {
  if (!specifier.startsWith("@/")) return;

  const importerPath = relative(importer);
  const importerLayer = importerPath.split("/")[0];
  const importedPath = specifier.slice(2);
  const importedLayer = importedPath.split("/")[0];

  if (importerLayer === "features" && importedLayer === "app") {
    report(importer, `feature code must not import the app layer: "${specifier}"`);
  }

  if (["components", "hooks", "lib"].includes(importerLayer) && ["app", "features"].includes(importedLayer)) {
    report(importer, `${importerLayer} is a shared layer and must not import ${importedLayer}: "${specifier}"`);
  }

  if (importerLayer === "features" && importedLayer === "features") {
    const importerFeature = importerPath.split("/")[1];
    const importedFeature = importedPath.split("/")[1];
    if (importerFeature && importedFeature && importerFeature !== importedFeature) {
      report(importer, `feature "${importerFeature}" must not import feature "${importedFeature}" directly`);
    }
  }
}

function findCycles(dependencies) {
  const visiting = new Set();
  const visited = new Set();
  const stack = [];
  const reportedCycles = new Set();

  function visit(file) {
    if (visited.has(file)) return;
    if (visiting.has(file)) {
      const cycleStart = stack.indexOf(file);
      const cycle = [...stack.slice(cycleStart), file].map(relative);
      const key = [...new Set(cycle)].sort().join("|");
      if (!reportedCycles.has(key)) {
        reportedCycles.add(key);
        violations.push(`dependency cycle: ${cycle.join(" -> ")}`);
      }
      return;
    }

    visiting.add(file);
    stack.push(file);
    for (const dependency of dependencies.get(file) ?? []) visit(dependency);
    stack.pop();
    visiting.delete(file);
    visited.add(file);
  }

  for (const file of dependencies.keys()) visit(file);
}

async function resolveLocalImport(importer, specifier) {
  let basePath;
  if (specifier.startsWith("@/")) {
    basePath = path.join(appRoot, specifier.slice(2));
  } else if (specifier.startsWith(".")) {
    basePath = path.resolve(path.dirname(importer), specifier);
  } else {
    return null;
  }

  for (const suffix of resolutionSuffixes) {
    const candidate = normalize(`${basePath}${suffix}`);
    if (!sourceSet.has(candidate)) continue;
    try {
      await access(candidate);
      return candidate;
    } catch {
      // Keep trying TypeScript resolution candidates.
    }
  }
  return null;
}

function collectImports(source) {
  const imports = new Set();
  const patterns = [
    /(?:from\s+|import\s*\(\s*)["']([^"']+)["']/g,
    /import\s+["']([^"']+)["']/g,
  ];

  for (const pattern of patterns) {
    for (const match of source.matchAll(pattern)) imports.add(match[1]);
  }
  return imports;
}

function isClientModule(source) {
  return /^\s*["']use client["'];?/u.test(source);
}

function isServerImport(specifier, resolved) {
  if (/(^|\/)server(?:\/|$)/u.test(specifier) || specifier === "server-only") return true;
  if (!resolved) return false;
  return /^\s*import\s+["']server-only["'];?/mu.test(sourceByFile.get(resolved) ?? "");
}

async function walk(directory) {
  const entries = await readdir(directory, { withFileTypes: true }).catch(() => []);
  const files = [];
  for (const entry of entries) {
    if ([".next", "node_modules"].includes(entry.name)) continue;
    const fullPath = path.join(directory, entry.name);
    if (entry.isDirectory()) files.push(...(await walk(fullPath)));
    else if (sourceExtensions.has(path.extname(entry.name))) files.push(fullPath);
  }
  return files;
}

function report(file, message) {
  violations.push(`${relative(file)}: ${message}`);
}

function relative(file) {
  return normalize(path.relative(appRoot, file));
}

function normalize(file) {
  return file.replaceAll("\\", "/");
}
