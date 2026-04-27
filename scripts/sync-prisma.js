/**
 * Apply SQL migrations and sync Prisma schemas/clients.
 *
 * Usage:
 *   DATABASE_URL=postgresql://... node scripts/sync-prisma.js
 */

const crypto = require("crypto");
const fs = require("fs");
const path = require("path");
const { spawnSync } = require("child_process");

const ROOT_DIR = path.join(__dirname, "..");
const MIGRATIONS_DIR = path.join(ROOT_DIR, "db", "main", "migrations");
const MIGRATIONS_TABLE = "shahrah_sql_migrations";

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

function listMigrationFiles() {
  if (!fs.existsSync(MIGRATIONS_DIR)) {
    fail(`Migrations directory not found: ${MIGRATIONS_DIR}`);
  }

  return fs
    .readdirSync(MIGRATIONS_DIR)
    .filter((file) => file.endsWith(".sql"))
    .sort();
}

function loadPgClient() {
  const pgPath = path.join(ROOT_DIR, "packages", "database", "node_modules", "pg");
  try {
    const pg = require(pgPath);
    return pg.Client;
  } catch {
    fail(
      "Cannot load pg client from packages/database/node_modules/pg. Run `pnpm install` first.",
    );
  }
}

function checksum(content) {
  return crypto.createHash("sha256").update(content).digest("hex");
}

async function ensureMigrationsTable(client) {
  await client.query(`
    CREATE TABLE IF NOT EXISTS ${MIGRATIONS_TABLE} (
      filename TEXT PRIMARY KEY,
      checksum TEXT NOT NULL,
      applied_at TIMESTAMPTZ NOT NULL DEFAULT NOW()
    )
  `);
}

async function readAppliedMigrations(client) {
  const { rows } = await client.query(
    `SELECT filename, checksum FROM ${MIGRATIONS_TABLE}`,
  );

  const map = new Map();
  for (const row of rows) {
    map.set(row.filename, row.checksum);
  }
  return map;
}

async function applyMigrations(databaseURL) {
  const Client = loadPgClient();
  const client = new Client({ connectionString: databaseURL });

  const files = listMigrationFiles();
  if (files.length === 0) {
    log("No SQL migrations found.");
    return;
  }

  log(`Found ${files.length} SQL migrations.`);
  files.forEach((file) => log(` - ${file}`));

  await client.connect();

  try {
    await ensureMigrationsTable(client);
    const applied = await readAppliedMigrations(client);

    for (const file of files) {
      const fullPath = path.join(MIGRATIONS_DIR, file);
      const sql = fs.readFileSync(fullPath, "utf8");
      const digest = checksum(sql);
      const appliedChecksum = applied.get(file);

      if (appliedChecksum && appliedChecksum === digest) {
        log(`Skipping already-applied migration: ${file}`);
        continue;
      }

      if (appliedChecksum && appliedChecksum !== digest) {
        fail(
          `Migration checksum mismatch for ${file}. Do not edit applied migrations; create a new migration file instead.`,
        );
      }

      log(`Applying migration: ${file}`);
      await client.query("BEGIN");
      try {
        await client.query(sql);
        await client.query(
          `INSERT INTO ${MIGRATIONS_TABLE} (filename, checksum) VALUES ($1, $2)`,
          [file, digest],
        );
        await client.query("COMMIT");
        log(`Applied migration: ${file}`);
      } catch (error) {
        await client.query("ROLLBACK");
        throw error;
      }
    }
  } finally {
    await client.end();
  }
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
      name: "@shahrah/tenant-admin",
      cwd: path.join(ROOT_DIR, "apps", "tenant-admin"),
      schema: "./prisma/schema.prisma",
      pull: false,
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

async function main() {
  const databaseURL = process.env.DATABASE_URL;
  if (!databaseURL) {
    fail("Missing DATABASE_URL. Set it before running this script.");
  }

  log("Starting SQL migration + Prisma sync...");
  await applyMigrations(databaseURL);
  syncPrisma();
  log("Done.");
}

main().catch((error) => {
  const message = error instanceof Error ? error.stack || error.message : String(error);
  fail(`Sync failed:\n${message}`);
});
