import { readFile, readdir } from "node:fs/promises";
import path from "node:path";
import process from "node:process";

const appRoot = path.resolve(import.meta.dirname, "..");
const sourceRoots = ["app", "components", "features", "hooks", "i18n", "lib", "messages"];
const checkedExtensions = new Set([".json", ".mts", ".ts", ".tsx"]);
const decoder = new TextDecoder("utf-8", { fatal: true });
const failures = [];
let checkedFiles = 0;

for (const root of sourceRoots) await scan(path.join(appRoot, root));

if (failures.length > 0) {
  console.error("Source encoding validation failed:\n");
  for (const failure of failures.sort()) console.error(`- ${failure}`);
  process.exitCode = 1;
} else {
  console.log(`Source encoding validation passed for ${checkedFiles} files.`);
}

async function scan(directory) {
  const entries = await readdir(directory, { withFileTypes: true }).catch(() => []);
  for (const entry of entries) {
    if ([".next", "node_modules"].includes(entry.name)) continue;
    const fullPath = path.join(directory, entry.name);
    if (entry.isDirectory()) {
      await scan(fullPath);
      continue;
    }
    if (!checkedExtensions.has(path.extname(entry.name))) continue;

    checkedFiles += 1;
    let source;
    try {
      source = decoder.decode(await readFile(fullPath));
    } catch {
      failures.push(`${relative(fullPath)} is not valid UTF-8`);
      continue;
    }

    const checks = [
      ["replacement character", /\uFFFD/u],
      ["question-mark corruption", /\?{4,}/u],
      ["common UTF-8 mojibake", /(?:Ã.|Â.|â€|ðŸ|Ø.|Ù.)/u],
    ];
    for (const [label, pattern] of checks) {
      if (pattern.test(source)) failures.push(`${relative(fullPath)} contains ${label}`);
    }
  }
}

function relative(file) {
  return path.relative(appRoot, file).replaceAll("\\", "/");
}
