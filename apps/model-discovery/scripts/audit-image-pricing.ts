import { promises as fs } from 'node:fs';
import { basename, join, relative } from 'node:path';
import { parse } from 'yaml';
import { parsePricingProfiles } from '../src/pricing/pricing-parser';

type YAMLRecord = Record<string, unknown>;

type ProfileAudit = {
  region: string;
  strategy?: 'image' | 'image_token' | 'token';
};

type ModelAudit = {
  provider: string;
  model: string;
  pricingStatus: 'unknown' | 'priced' | 'free';
  resolved: boolean;
  profiles: ProfileAudit[];
  file: string;
};

const defaultProvidersRoot = 'D:\\source\\models\\providers';
const compatibleOutputUnits = ['image', 'image_token', 'token'] as const;

function getArgValue(flag: string): string | undefined {
  const index = process.argv.findIndex((value) => value === flag);
  return index >= 0 ? process.argv[index + 1] : undefined;
}

function asRecord(value: unknown): YAMLRecord | undefined {
  if (!value || typeof value !== 'object' || Array.isArray(value)) {
    return undefined;
  }
  return value as YAMLRecord;
}

function imageMode(doc: YAMLRecord): boolean {
  const supportedModes = Array.isArray(doc['supportedModes'])
    ? doc['supportedModes']
    : [];
  const values = [doc['mode'], ...supportedModes];

  return values.some((raw) => {
    const normalized = String(raw ?? '')
      .trim()
      .toLowerCase()
      .replace(/[\s-]+/g, '_');
    return [
      'image',
      'image_generation',
      'text_to_image',
      'image_edit',
    ].includes(normalized);
  });
}

async function collectYAMLFiles(directory: string): Promise<string[]> {
  const entries = await fs.readdir(directory, { withFileTypes: true });
  const files: string[] = [];

  for (const entry of entries) {
    const path = join(directory, entry.name);
    if (entry.isDirectory()) {
      files.push(...(await collectYAMLFiles(path)));
    } else if (
      entry.isFile() &&
      entry.name.endsWith('.yaml') &&
      entry.name !== 'default.yaml'
    ) {
      files.push(path);
    }
  }

  return files;
}

async function auditModel(
  providersRoot: string,
  file: string,
): Promise<ModelAudit | undefined> {
  const doc = asRecord(parse(await fs.readFile(file, 'utf8')));
  const model = typeof doc?.['model'] === 'string' ? doc['model'] : undefined;
  if (!doc || !model || doc['isDeprecated'] === true || !imageMode(doc)) {
    return undefined;
  }

  const sources = Array.isArray(doc['sources']) ? doc['sources'] : [];
  const pricing = parsePricingProfiles(doc['costs'], {
    explicitFree: doc['free'] === true,
    sourceUrl: typeof sources[0] === 'string' ? sources[0] : undefined,
  });
  const profiles: ProfileAudit[] = pricing.profiles.map((profile) => {
    const strategy = compatibleOutputUnits.find((unit) =>
      profile.rates.some(
        (rate) =>
          rate.usage === 'output' &&
          rate.unit === unit &&
          rate.serviceTier === 'standard',
      ),
    );
    return { region: profile.regionSlug, strategy };
  });
  const resolved =
    pricing.status === 'free' ||
    (pricing.status === 'priced' &&
      profiles.length > 0 &&
      profiles.every((profile) => profile.strategy !== undefined));

  return {
    provider: relative(providersRoot, file).split(/[\\/]/)[0] ?? basename(file),
    model,
    pricingStatus: pricing.status,
    resolved,
    profiles,
    file,
  };
}

async function main() {
  const providersRoot = getArgValue('--providers-root') ?? defaultProvidersRoot;
  const files = await collectYAMLFiles(providersRoot);
  const audited = await Promise.all(
    files.map((file) => auditModel(providersRoot, file)),
  );
  const models = audited.filter(
    (model): model is ModelAudit => model !== undefined,
  );
  const resolved = models.filter((model) => model.resolved);
  const unresolved = models.filter((model) => !model.resolved);
  const strategyModels = {
    image: 0,
    image_token: 0,
    token: 0,
    free: 0,
  };

  for (const model of resolved) {
    if (model.pricingStatus === 'free') {
      strategyModels.free += 1;
      continue;
    }
    const strategies = new Set(
      model.profiles.flatMap((profile) =>
        profile.strategy ? [profile.strategy] : [],
      ),
    );
    for (const strategy of strategies) {
      strategyModels[strategy] += 1;
    }
  }

  console.log(
    JSON.stringify(
      {
        total: models.length,
        resolved: resolved.length,
        unresolved: unresolved.length,
        strategyModels,
        unresolvedModels: unresolved,
      },
      null,
      2,
    ),
  );
}

main().catch((error) => {
  console.error(error);
  process.exitCode = 1;
});
