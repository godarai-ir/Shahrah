import { readFile, readdir, stat } from "node:fs/promises";
import path from "node:path";
import process from "node:process";

const appRoot = process.cwd();
const budget = JSON.parse(
  await readFile(path.join(appRoot, "performance-budget.json"), "utf8"),
);
const budgetKeys = [
  "maxFirstLoadUncompressedJsBytes",
  "maxInitialChunkBytes",
  "maxStaticJsBytes",
];
for (const key of budgetKeys) {
  if (!Number.isFinite(budget[key]) || budget[key] <= 0) {
    throw new Error(`Performance budget ${key} must be a positive number.`);
  }
}
const statsPath = path.join(appRoot, ".next", "diagnostics", "route-bundle-stats.json");

let routeStats;
try {
  routeStats = JSON.parse(await readFile(statsPath, "utf8"));
} catch (error) {
  console.error("Performance diagnostics are unavailable. Run `pnpm build` first.");
  throw error;
}
if (!Array.isArray(routeStats) || routeStats.length === 0) {
  throw new Error("Performance route diagnostics are empty or invalid.");
}

const largestRoute = [...routeStats].sort(
  (left, right) => right.firstLoadUncompressedJsBytes - left.firstLoadUncompressedJsBytes,
)[0];
const initialChunkPaths = new Set(routeStats.flatMap((route) => route.firstLoadChunkPaths));
if (initialChunkPaths.size === 0) {
  throw new Error("Performance route diagnostics do not contain initial chunks.");
}
const initialChunkSizes = await Promise.all(
  [...initialChunkPaths].map(async (chunkPath) => ({
    chunkPath,
    bytes: (await stat(path.join(appRoot, chunkPath))).size,
  })),
);
const largestInitialChunk = initialChunkSizes.sort((left, right) => right.bytes - left.bytes)[0];
const staticJsRoot = path.join(appRoot, ".next", "static", "chunks");
const staticJsBytes = await sumJavaScriptBytes(staticJsRoot);
const violations = [];

if (largestRoute.firstLoadUncompressedJsBytes > budget.maxFirstLoadUncompressedJsBytes) {
  violations.push(
    `${largestRoute.route} first-load JS is ${formatBytes(largestRoute.firstLoadUncompressedJsBytes)}; budget is ${formatBytes(budget.maxFirstLoadUncompressedJsBytes)}.`,
  );
}
if (largestInitialChunk.bytes > budget.maxInitialChunkBytes) {
  violations.push(
    `${largestInitialChunk.chunkPath} is ${formatBytes(largestInitialChunk.bytes)}; initial chunk budget is ${formatBytes(budget.maxInitialChunkBytes)}.`,
  );
}
if (staticJsBytes > budget.maxStaticJsBytes) {
  violations.push(
    `Static JS totals ${formatBytes(staticJsBytes)}; deployment budget is ${formatBytes(budget.maxStaticJsBytes)}.`,
  );
}

console.log(`Largest route: ${largestRoute.route} (${formatBytes(largestRoute.firstLoadUncompressedJsBytes)})`);
console.log(`Largest initial chunk: ${formatBytes(largestInitialChunk.bytes)}`);
console.log(`Static JS total: ${formatBytes(staticJsBytes)}`);

if (violations.length > 0) {
  console.error(`\nPerformance budget failed:\n- ${violations.join("\n- ")}`);
  process.exitCode = 1;
} else {
  console.log("Performance budget passed.");
}

async function sumJavaScriptBytes(directory) {
  const entries = await readdir(directory, { withFileTypes: true });
  const sizes = await Promise.all(entries.map(async (entry) => {
    const entryPath = path.join(directory, entry.name);
    if (entry.isDirectory()) return sumJavaScriptBytes(entryPath);
    if (!entry.name.endsWith(".js")) return 0;
    return (await stat(entryPath)).size;
  }));
  return sizes.reduce((total, size) => total + size, 0);
}

function formatBytes(bytes) {
  return `${(bytes / 1024).toFixed(1)} KiB`;
}
