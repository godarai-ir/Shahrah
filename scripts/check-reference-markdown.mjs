#!/usr/bin/env node

import { open, readdir } from "node:fs/promises";
import path from "node:path";
import { fileURLToPath } from "node:url";

const scriptDirectory = path.dirname(fileURLToPath(import.meta.url));
const repositoryRoot = path.resolve(scriptDirectory, "..");
const roots = process.argv.slice(2);
const scanRoots = roots.length > 0 ? roots : [".portkey-docs/markdown"];
const invalidFiles = [];

for (const scanRoot of scanRoots) {
  await scanDirectory(path.resolve(repositoryRoot, scanRoot));
}

if (invalidFiles.length > 0) {
  console.error("HTML snapshots found in Markdown reference files:");
  for (const filePath of invalidFiles) {
    console.error(`- ${path.relative(repositoryRoot, filePath)}`);
  }
  process.exit(1);
}

console.log(`Reference Markdown check passed (${scanRoots.join(", ")}).`);

async function scanDirectory(directory) {
  for (const entry of await readdir(directory, { withFileTypes: true })) {
    const entryPath = path.join(directory, entry.name);
    if (entry.isDirectory()) {
      await scanDirectory(entryPath);
      continue;
    }
    if (!entry.isFile() || path.extname(entry.name).toLowerCase() !== ".md") {
      continue;
    }

    const handle = await open(entryPath, "r");
    try {
      const buffer = Buffer.alloc(512);
      const { bytesRead } = await handle.read(buffer, 0, buffer.length, 0);
      const prefix = buffer.subarray(0, bytesRead).toString("utf8");
      if (/^\s*(?:<!doctype\s+html|<html\b)/i.test(prefix)) {
        invalidFiles.push(entryPath);
      }
    } finally {
      await handle.close();
    }
  }
}
