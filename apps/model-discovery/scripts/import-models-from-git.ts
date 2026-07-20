import 'dotenv/config';
import { PrismaPg } from '@prisma/adapter-pg';
import { PrismaClient } from '../src/generated/prisma/client';
import { mkdtemp, rm } from 'node:fs/promises';
import { tmpdir } from 'node:os';
import { join } from 'node:path';
import { randomUUID } from 'node:crypto';
import { spawn } from 'node:child_process';

type ImportSourceStateRow = {
  lastCommitSha: string | null;
};

const connectionString =
  process.env.DATABASE_URL ??
  'postgresql://postgres:postgres@127.0.0.1:5432/model_discovery?schema=public';

const prisma = new PrismaClient({
  adapter: new PrismaPg({ connectionString }),
});

function getArgValue(flag: string): string | undefined {
  const idx = process.argv.findIndex((x) => x === flag);
  if (idx < 0) return undefined;
  return process.argv[idx + 1];
}

function hasFlag(flag: string): boolean {
  return process.argv.includes(flag);
}

function commandName(base: string): string {
  return process.platform === 'win32' ? `${base}.cmd` : base;
}

function runCommand(
  command: string,
  args: string[],
  options: { cwd?: string; capture?: boolean } = {},
): Promise<string> {
  return new Promise((resolve, reject) => {
    const child = spawn(commandName(command), args, {
      cwd: options.cwd,
      stdio: options.capture ? ['ignore', 'pipe', 'pipe'] : 'inherit',
    });

    let stdout = '';
    let stderr = '';
    if (options.capture) {
      child.stdout?.on('data', (chunk) => {
        stdout += chunk.toString();
      });
      child.stderr?.on('data', (chunk) => {
        stderr += chunk.toString();
      });
    }

    child.on('error', reject);
    child.on('close', (code) => {
      if (code === 0) {
        resolve(stdout.trim());
        return;
      }
      reject(
        new Error(
          `${command} ${args.join(' ')} failed with exit code ${code}${
            stderr ? `\n${stderr.trim()}` : ''
          }`,
        ),
      );
    });
  });
}

async function resolveRemoteCommit(
  repoUrl: string,
  ref: string,
): Promise<string> {
  const output = await runCommand('git', ['ls-remote', repoUrl, ref], {
    capture: true,
  });
  const line = output
    .split(/\r?\n/)
    .find((entry) => /^[0-9a-f]{40}\s+/i.test(entry));
  const commitSha = line?.split(/\s+/)[0];
  if (!commitSha) {
    throw new Error(`Could not resolve ${ref} from ${repoUrl}`);
  }
  return commitSha;
}

async function getImportState(
  sourceKey: string,
): Promise<ImportSourceStateRow | undefined> {
  const rows = await prisma.$queryRaw<ImportSourceStateRow[]>`
    SELECT "lastCommitSha"
    FROM "ImportSourceState"
    WHERE "sourceKey" = ${sourceKey}
    LIMIT 1
  `;
  return rows[0];
}

async function markChecked(
  sourceKey: string,
  repoUrl: string,
  ref: string,
): Promise<void> {
  await prisma.$executeRaw`
    INSERT INTO "ImportSourceState" (
      "id",
      "sourceKey",
      "repoUrl",
      "ref",
      "lastCheckedAt",
      "createdAt",
      "updatedAt"
    )
    VALUES (
      ${randomUUID()},
      ${sourceKey},
      ${repoUrl},
      ${ref},
      NOW(),
      NOW(),
      NOW()
    )
    ON CONFLICT ("sourceKey") DO UPDATE SET
      "repoUrl" = EXCLUDED."repoUrl",
      "ref" = EXCLUDED."ref",
      "lastCheckedAt" = NOW(),
      "updatedAt" = NOW()
  `;
}

async function markImported(
  sourceKey: string,
  repoUrl: string,
  ref: string,
  commitSha: string,
): Promise<void> {
  await prisma.$executeRaw`
    INSERT INTO "ImportSourceState" (
      "id",
      "sourceKey",
      "repoUrl",
      "ref",
      "lastCommitSha",
      "lastImportedAt",
      "lastCheckedAt",
      "createdAt",
      "updatedAt"
    )
    VALUES (
      ${randomUUID()},
      ${sourceKey},
      ${repoUrl},
      ${ref},
      ${commitSha},
      NOW(),
      NOW(),
      NOW(),
      NOW()
    )
    ON CONFLICT ("sourceKey") DO UPDATE SET
      "repoUrl" = EXCLUDED."repoUrl",
      "ref" = EXCLUDED."ref",
      "lastCommitSha" = EXCLUDED."lastCommitSha",
      "lastImportedAt" = NOW(),
      "lastCheckedAt" = NOW(),
      "updatedAt" = NOW()
  `;
}

function packageRunner(): string {
  const execPath = process.env.npm_execpath?.toLowerCase() ?? '';
  if (execPath.includes('pnpm')) return 'pnpm';
  return 'npm';
}

async function runYamlImport(providersRoot: string): Promise<void> {
  const args = [
    'run',
    'import:providers:yaml',
    '--',
    '--providers-root',
    providersRoot,
  ];

  const providers = getArgValue('--providers');
  if (providers) args.push('--providers', providers);
  if (hasFlag('--clear-missing')) args.push('--clear-missing');
  if (hasFlag('--dry-run')) args.push('--dry-run');

  await runCommand(packageRunner(), args);
}

async function cloneSource(
  repoUrl: string,
  expectedCommitSha: string,
): Promise<string> {
  const cloneRoot = await mkdtemp(join(tmpdir(), 'model-discovery-import-'));

  try {
    await runCommand('git', ['init', cloneRoot]);
    await runCommand('git', ['remote', 'add', 'origin', repoUrl], {
      cwd: cloneRoot,
    });
    await runCommand(
      'git',
      ['fetch', '--depth', '1', 'origin', expectedCommitSha],
      {
        cwd: cloneRoot,
      },
    );
    await runCommand('git', ['checkout', '--detach', 'FETCH_HEAD'], {
      cwd: cloneRoot,
    });

    const actualCommitSha = await runCommand('git', ['rev-parse', 'HEAD'], {
      cwd: cloneRoot,
      capture: true,
    });
    if (actualCommitSha !== expectedCommitSha) {
      throw new Error(
        `Checked out ${actualCommitSha}, but expected ${expectedCommitSha}`,
      );
    }

    return cloneRoot;
  } catch (err) {
    await rm(cloneRoot, { recursive: true, force: true });
    throw err;
  }
}

async function main() {
  const repoUrl =
    getArgValue('--repo-url') ?? 'https://github.com/truefoundry/models.git';
  const ref = getArgValue('--ref') ?? 'main';
  const pinnedCommitSha = getArgValue('--commit-sha');
  const sourceKey = getArgValue('--source-key') ?? `${repoUrl}#${ref}`;
  const providersPath = getArgValue('--providers-path') ?? 'providers';
  const dryRun = hasFlag('--dry-run');
  const force = hasFlag('--force');

  const remoteCommitSha =
    pinnedCommitSha ?? (await resolveRemoteCommit(repoUrl, ref));
  if (!/^[0-9a-f]{40}$/i.test(remoteCommitSha)) {
    throw new Error(`Invalid commit SHA: ${remoteCommitSha}`);
  }
  const state = await getImportState(sourceKey);

  if (!force && state?.lastCommitSha === remoteCommitSha) {
    if (!dryRun) {
      await markChecked(sourceKey, repoUrl, ref);
    }
    console.log(
      `[import-git] skipped source=${sourceKey} ref=${ref} commit=${remoteCommitSha}`,
    );
    return;
  }

  let cloneRoot: string | undefined;
  try {
    cloneRoot = await cloneSource(repoUrl, remoteCommitSha);
    await runYamlImport(join(cloneRoot, providersPath));
    if (!dryRun) {
      await markImported(sourceKey, repoUrl, ref, remoteCommitSha);
    }
    console.log(
      `[import-git] completed source=${sourceKey} ref=${ref} commit=${remoteCommitSha} dryRun=${dryRun}`,
    );
  } finally {
    if (cloneRoot) {
      await rm(cloneRoot, { recursive: true, force: true });
    }
  }
}

main()
  .catch((err) => {
    console.error(err);
    process.exitCode = 1;
  })
  .finally(async () => {
    await prisma.$disconnect();
  });
