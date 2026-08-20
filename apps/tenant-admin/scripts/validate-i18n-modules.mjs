import fs from "node:fs";
import path from "node:path";

const root = path.resolve(process.cwd(), "messages");
const modulesRoot = path.join(root, "modules");
const requestConfigPath = path.resolve(process.cwd(), "i18n", "request.ts");
const locales = ["fa", "en"];

function fail(message) {
  console.error(`i18n validation failed: ${message}`);
  process.exit(1);
}

if (!fs.existsSync(modulesRoot)) {
  fail(`modules directory not found at ${modulesRoot}`);
}

const namespaces = fs
  .readdirSync(modulesRoot, { withFileTypes: true })
  .filter((entry) => entry.isDirectory())
  .map((entry) => entry.name)
  .sort();

if (namespaces.length === 0) {
  fail("no namespaces found in messages/modules");
}

if (!fs.existsSync(requestConfigPath)) {
  fail(`request config not found at ${requestConfigPath}`);
}

const requestConfig = fs.readFileSync(requestConfigPath, "utf8");
const manifestMatch = requestConfig.match(/const MESSAGE_NAMESPACES = \[([\s\S]*?)\] as const;/);
if (!manifestMatch) {
  fail(`MESSAGE_NAMESPACES manifest not found in ${requestConfigPath}`);
}

const registeredNamespaces = [...manifestMatch[1].matchAll(/["']([^"']+)["']/g)]
  .map((match) => match[1])
  .sort();
const missingRegistrations = namespaces.filter((namespace) => !registeredNamespaces.includes(namespace));
const missingDirectories = registeredNamespaces.filter((namespace) => !namespaces.includes(namespace));
if (missingRegistrations.length > 0 || missingDirectories.length > 0) {
  fail(
    `namespace registry mismatch. Missing registrations: ${missingRegistrations.join(", ") || "none"}; ` +
      `missing directories: ${missingDirectories.join(", ") || "none"}`,
  );
}

for (const locale of locales) {
  const basePath = path.join(root, `${locale}.json`);
  if (!fs.existsSync(basePath)) fail(`missing base file ${basePath}`);

  let baseJson;
  try {
    baseJson = JSON.parse(fs.readFileSync(basePath, "utf8"));
  } catch (error) {
    fail(`invalid JSON in ${basePath}: ${error instanceof Error ? error.message : String(error)}`);
  }

  const baseKeys = Object.keys(baseJson);
  if (baseKeys.length > 0) {
    fail(`${locale}.json should not define namespaces after phase 2 cleanup. Found: ${baseKeys.join(", ")}`);
  }
}

for (const namespace of namespaces) {
  for (const locale of locales) {
    const modulePath = path.join(modulesRoot, namespace, `${locale}.json`);
    if (!fs.existsSync(modulePath)) {
      fail(`missing module file ${modulePath}`);
    }

    try {
      JSON.parse(fs.readFileSync(modulePath, "utf8"));
    } catch (error) {
      fail(`invalid JSON in ${modulePath}: ${error instanceof Error ? error.message : String(error)}`);
    }
  }
}

console.log(`i18n validation passed for ${namespaces.length} namespaces.`);
