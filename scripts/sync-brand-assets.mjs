import { createHash } from "node:crypto";
import {
  copyFileSync,
  existsSync,
  mkdirSync,
  readFileSync,
  readdirSync,
  rmSync,
  statSync,
  writeFileSync,
} from "node:fs";
import { dirname, extname, join, relative, resolve } from "node:path";
import { fileURLToPath } from "node:url";

const repositoryRoot = resolve(dirname(fileURLToPath(import.meta.url)), "..");
const sourceRoot = join(repositoryRoot, "web-assets");
const checkOnly = process.argv.includes("--check");

const apps = [
  {
    name: "docs",
    publicRoot: "apps/docs/static",
    manifest: {
      name: "مستندات گدارAI",
      short_name: "مستندات گدارAI",
      description: "مستندات فنی پلتفرم گدارAI",
      start_url: "/docs/",
      scope: "/docs/",
    },
  },
  {
    name: "model-discovery-web",
    publicRoot: "apps/model-discovery-web/public",
    manifest: {
      name: "کاتالوگ مدل‌های گدارAI",
      short_name: "کاتالوگ گدارAI",
      description: "کشف و مقایسه مدل‌های هوش مصنوعی در گدارAI",
      start_url: "/catalog/",
      scope: "/catalog/",
    },
  },
  {
    name: "blog",
    publicRoot: "apps/blog/public",
    manifest: {
      name: "بلاگ گدارAI",
      short_name: "بلاگ گدارAI",
      description: "مقاله‌ها، تحلیل‌ها و آموزش‌های فنی گدارAI",
      start_url: "/blog/",
      scope: "/blog/",
    },
  },
  {
    name: "website",
    publicRoot: "apps/website/public",
    manifest: {
      name: "گدارAI",
      short_name: "گدارAI",
      description: "درگاه مدل‌های زبانی برای تیم‌های ایرانی",
      start_url: "/",
      scope: "/",
    },
  },
  {
    name: "tenant-admin",
    publicRoot: "apps/tenant-admin/public",
    manifest: {
      name: "پنل مدیریت گدارAI",
      short_name: "پنل گدارAI",
      description: "پنل مدیریت فضای کار گدارAI",
      start_url: "/",
      scope: "/",
    },
  },
];

const sharedFiles = [
  ["svg/godarai-mark-light.svg", "brand/svg/godarai-mark-light.svg"],
  ["svg/godarai-mark-dark.svg", "brand/svg/godarai-mark-dark.svg"],
  ["svg/godarai-mark-adaptive.svg", "brand/svg/godarai-mark-adaptive.svg"],
  ["svg/godarai-mark-current-color.svg", "brand/svg/godarai-mark-current-color.svg"],
  ["svg/godarai-mark-mono-navy.svg", "brand/svg/godarai-mark-mono-navy.svg"],
  ["svg/godarai-mark-mono-white.svg", "brand/svg/godarai-mark-mono-white.svg"],
  ["svg/safari-pinned-tab.svg", "brand/svg/safari-pinned-tab.svg"],
  ["icons/favicon.svg", "brand/icons/favicon.svg"],
  ["icons/favicon.ico", "brand/icons/favicon.ico"],
  ["icons/favicon-light.ico", "brand/icons/favicon-light.ico"],
  ["icons/favicon-dark.ico", "brand/icons/favicon-dark.ico"],
  ["icons/favicon-16x16.png", "brand/icons/favicon-16x16.png"],
  ["icons/favicon-32x32.png", "brand/icons/favicon-32x32.png"],
  ["icons/apple-touch-icon.png", "brand/icons/apple-touch-icon.png"],
  ["icons/android-chrome-192x192.png", "brand/icons/android-chrome-192x192.png"],
  ["icons/android-chrome-512x512.png", "brand/icons/android-chrome-512x512.png"],
  ["icons/maskable-icon-192x192.png", "brand/icons/maskable-icon-192x192.png"],
  ["icons/maskable-icon-512x512.png", "brand/icons/maskable-icon-512x512.png"],
  ["icons/mstile-150x150.png", "brand/icons/mstile-150x150.png"],
  ["social/og-image-light-1200x630.png", "brand/social/og-image-light-1200x630.png"],
  ["social/og-image-dark-1200x630.png", "brand/social/og-image-dark-1200x630.png"],
  ["png/natural/light/godarai-mark-light-256x188.png", "brand/png/godarai-mark-light-256x188.png"],
  ["web/browserconfig.xml", "brand/web/browserconfig.xml"],
];

const compatibilityAliases = [
  ["icons/favicon.ico", "favicon.ico"],
  ["icons/favicon.svg", "favicon.svg"],
  ["icons/favicon-16x16.png", "favicon-16x16.png"],
  ["icons/favicon-32x32.png", "favicon-32x32.png"],
  ["icons/apple-touch-icon.png", "apple-touch-icon.png"],
];

const repositoryAliases = [
  ["icons/favicon.ico", "apps/model-discovery-web/app/favicon.ico"],
  ["icons/android-chrome-192x192.png", "apps/website/public/android-chrome-192x192.png"],
  ["icons/android-chrome-512x512.png", "apps/website/public/android-chrome-512x512.png"],
  ["icons/app-icon-1024x1024.png", "apps/tenant-admin/app/icon.png"],
  ["icons/apple-touch-icon.png", "apps/tenant-admin/app/apple-icon.png"],
];

const legacyFiles = [
  "apps/docs/static/img/mini_logo.svg",
  "apps/docs/static/img/logo.svg",
  "apps/docs/static/img/favicon.ico",
  "apps/model-discovery-web/public/mini_logo.svg",
  "apps/website/public/mini_logo.svg",
  "apps/website/public/mini_logo.png",
  "apps/tenant-admin/public/brand/mini_logo.svg",
  "apps/tenant-admin/public/brand/mini_logo.png",
  "apps/tenant-admin/public/brand/shahrah-logo.png",
  "apps/website/public/site.webmanifest",
];

const manifestIcons = [
  {
    src: "../icons/android-chrome-192x192.png",
    sizes: "192x192",
    type: "image/png",
    purpose: "any",
  },
  {
    src: "../icons/android-chrome-512x512.png",
    sizes: "512x512",
    type: "image/png",
    purpose: "any",
  },
  {
    src: "../icons/maskable-icon-192x192.png",
    sizes: "192x192",
    type: "image/png",
    purpose: "maskable",
  },
  {
    src: "../icons/maskable-icon-512x512.png",
    sizes: "512x512",
    type: "image/png",
    purpose: "maskable",
  },
];

const failures = [];
const expectedPngDimensions = new Map([
  ["icons/android-chrome-192x192.png", [192, 192]],
  ["icons/android-chrome-512x512.png", [512, 512]],
  ["icons/maskable-icon-192x192.png", [192, 192]],
  ["icons/maskable-icon-512x512.png", [512, 512]],
  ["icons/apple-touch-icon.png", [180, 180]],
  ["social/og-image-light-1200x630.png", [1200, 630]],
  ["social/og-image-dark-1200x630.png", [1200, 630]],
]);

for (const [sourceRelative, expected] of expectedPngDimensions) {
  const source = join(sourceRoot, sourceRelative);
  if (!existsSync(source)) {
    failures.push(`Missing source asset: ${sourceRelative}`);
    continue;
  }

  const actual = readPngDimensions(source);
  if (actual[0] !== expected[0] || actual[1] !== expected[1]) {
    failures.push(
      `Unexpected PNG dimensions: ${sourceRelative} is ${actual.join("x")}, expected ${expected.join("x")}`,
    );
  }
}

for (const app of apps) {
  const publicRoot = join(repositoryRoot, app.publicRoot);

  for (const [sourceRelative, destinationRelative] of [...sharedFiles, ...compatibilityAliases]) {
    syncFile(
      join(sourceRoot, sourceRelative),
      join(publicRoot, destinationRelative),
      `${app.name}:${destinationRelative}`,
    );
  }

  const manifest = `${JSON.stringify(
    {
      ...app.manifest,
      lang: "fa",
      dir: "rtl",
      display: "standalone",
      background_color: "#06182D",
      theme_color: "#012651",
      icons: manifestIcons,
    },
    null,
    2,
  )}\n`;
  syncText(join(publicRoot, "brand/web/site.webmanifest"), manifest, `${app.name}:manifest`);
}

for (const [sourceRelative, destinationRelative] of repositoryAliases) {
  syncFile(
    join(sourceRoot, sourceRelative),
    join(repositoryRoot, destinationRelative),
    destinationRelative,
  );
}

if (!checkOnly) {
  for (const legacyRelative of legacyFiles) {
    const legacyPath = join(repositoryRoot, legacyRelative);
    if (existsSync(legacyPath)) rmSync(legacyPath);
  }
}

for (const legacyRelative of legacyFiles) {
  if (existsSync(join(repositoryRoot, legacyRelative))) {
    failures.push(`Legacy brand asset still exists: ${legacyRelative}`);
  }
}

for (const sourceFile of collectSourceFiles()) {
  const contents = readFileSync(sourceFile, "utf8");
  if (contents.includes("mini_logo") || contents.includes("shahrah-logo")) {
    failures.push(`Legacy brand reference: ${relative(repositoryRoot, sourceFile)}`);
  }
}

if (failures.length > 0) {
  console.error(failures.join("\n"));
  process.exitCode = 1;
} else {
  console.log(
    checkOnly
      ? "Brand assets, manifests, and legacy-reference checks passed."
      : "Brand assets and manifests synchronized successfully.",
  );
}

function syncFile(source, destination, label) {
  if (!existsSync(source)) {
    failures.push(`Missing source asset: ${relative(repositoryRoot, source)}`);
    return;
  }

  if (checkOnly) {
    if (!existsSync(destination)) {
      failures.push(`Missing synchronized asset: ${label}`);
      return;
    }
    if (digest(source) !== digest(destination)) {
      failures.push(`Out-of-date synchronized asset: ${label}`);
    }
    return;
  }

  mkdirSync(dirname(destination), { recursive: true });
  copyFileSync(source, destination);
}

function syncText(destination, expected, label) {
  if (checkOnly) {
    if (!existsSync(destination)) {
      failures.push(`Missing generated file: ${label}`);
      return;
    }
    if (readFileSync(destination, "utf8") !== expected) {
      failures.push(`Out-of-date generated file: ${label}`);
    }
    return;
  }

  mkdirSync(dirname(destination), { recursive: true });
  writeFileSync(destination, expected, "utf8");
}

function digest(path) {
  return createHash("sha256").update(readFileSync(path)).digest("hex");
}

function readPngDimensions(path) {
  const png = readFileSync(path);
  const signature = "89504e470d0a1a0a";
  if (png.length < 24 || png.subarray(0, 8).toString("hex") !== signature) {
    return [-1, -1];
  }
  return [png.readUInt32BE(16), png.readUInt32BE(20)];
}

function collectSourceFiles() {
  const sourceExtensions = new Set([".astro", ".css", ".html", ".js", ".jsx", ".json", ".md", ".mdx", ".mjs", ".ts", ".tsx"]);
  const ignoredDirectories = new Set([".astro", ".docusaurus", ".next", ".turbo", "build", "dist", "node_modules", "public", "static"]);
  const roots = apps.map((app) => join(repositoryRoot, "apps", app.name));

  return roots.flatMap((root) => walk(root));

  function walk(directory) {
    return readdirSync(directory, { withFileTypes: true }).flatMap((entry) => {
      if (ignoredDirectories.has(entry.name)) return [];
      const path = join(directory, entry.name);
      if (entry.isDirectory()) return walk(path);
      if (!entry.isFile() || !sourceExtensions.has(extname(entry.name))) return [];
      return [path];
    });
  }
}
