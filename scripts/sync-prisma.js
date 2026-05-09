/**
 * Apply Goose migrations and sync Prisma schemas/clients.
 *
 * Usage:
 *   DATABASE_URL=postgresql://... node scripts/sync-prisma.js
 */

const path = require("path");
const { spawnSync } = require("child_process");

const ROOT_DIR = path.join(__dirname, "..");
const MIGRATIONS_DIR = path.join(ROOT_DIR, "db", "main", "migrations");

function log(message) {
  process.stdout.write(`${message}\n`);
}

function fail(message) {
  process.stderr.write(`${message}\n`);
  process.exit(1);
}

function run(command, args, options = {}) {
  const shell = process.platform === "win32";
  const result = spawnSync(command, args, {
    cwd: options.cwd || ROOT_DIR,
    stdio: "inherit",
    shell,
    env: options.env || process.env,
  });

  if (result.status !== 0) {
    const rendered = `${command} ${args.join(" ")}`.trim();
    fail(`Command failed: ${rendered}`);
  }
}

function createPrismaEnv() {
  const env = { ...process.env };
  const databaseURL = env.DATABASE_URL;
  if (!databaseURL) {
    return env;
  }

  try {
    const parsed = new URL(databaseURL);
    if (parsed.hostname === "localhost") {
      parsed.hostname = "127.0.0.1";
      env.DATABASE_URL = parsed.toString();
    }
  } catch {
    // Keep the original value if URL parsing fails.
  }

  return env;
}

function ensureGoose() {
  const check = spawnSync("goose", ["-version"], {
    cwd: ROOT_DIR,
    stdio: "ignore",
    shell: process.platform === "win32",
    env: process.env,
  });

  if (check.status === 0) {
    return;
  }

  log("goose not found. Installing goose CLI...");
  run("go", ["install", "github.com/pressly/goose/v3/cmd/goose@latest"]);
}

function applyGooseMigrations(databaseURL) {
  ensureGoose();
  log("Applying Goose migrations...");
  run("goose", ["postgres", databaseURL, "up"], { cwd: MIGRATIONS_DIR });
}

function syncPrisma() {
  const prismaEnv = createPrismaEnv();

  const prismaJobs = [
    {
      name: "@shahrah/database",
      cwd: path.join(ROOT_DIR, "packages", "database"),
      schema: "./schema.prisma",
      pull: true,
    },
    {
      name: "@shahrah/support-admin",
      cwd: path.join(ROOT_DIR, "apps", "support-admin"),
      schema: "./prisma/schema.prisma",
      pull: false,
    },
  ];

  for (const job of prismaJobs) {
    if (job.pull) {
      log(`Running Prisma db pull for ${job.name}...`);
      run(
        "pnpm",
        ["prisma", "db", "pull", `--schema=${job.schema}`],
        { cwd: job.cwd, env: prismaEnv },
      );
    }

    log(`Formatting schema (${job.name}): ${job.schema}`);
    run(
      "pnpm",
      ["prisma", "format", `--schema=${job.schema}`],
      { cwd: job.cwd, env: prismaEnv },
    );

    log(`Generating Prisma client: ${job.name}`);
    run("pnpm", ["run", "prisma:generate"], { cwd: job.cwd, env: prismaEnv });
  }
}

function main() {
  const databaseURL = process.env.DATABASE_URL;
  if (!databaseURL) {
    fail("Missing DATABASE_URL. Set it before running this script.");
  }

  log("Starting Goose migration + Prisma sync...");
  applyGooseMigrations(databaseURL);
  syncPrisma();
  log("Done.");
}

try {
  main();
} catch (error) {
  const message = error instanceof Error ? error.stack || error.message : String(error);
  fail(`Sync failed:\n${message}`);
}
