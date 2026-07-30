import {
  DEFAULT_PROVIDERS_ROOT,
  disconnectProviderYamlImporter,
  importProviderYaml,
} from './import-models-from-yaml';
import {
  ANTHROPIC_PROVIDER_SLUG,
  ANTHROPIC_SUPPORTED_MODEL_IDS,
  validateAnthropicProviderDirectory,
} from './model-import/providers/anthropic';

function getArgValue(args: string[], flag: string): string | undefined {
  const index = args.indexOf(flag);
  return index < 0 ? undefined : args[index + 1];
}

function hasFlag(args: string[], flag: string): boolean {
  return args.includes(flag);
}

async function main(args: string[] = process.argv.slice(2)): Promise<void> {
  if (hasFlag(args, '--providers')) {
    throw new Error(
      'The Anthropic importer is locked to provider=anthropic; remove --providers.',
    );
  }

  const providersRoot =
    getArgValue(args, '--providers-root') ?? DEFAULT_PROVIDERS_ROOT;
  const validation = await validateAnthropicProviderDirectory(providersRoot);
  const dryRun = hasFlag(args, '--dry-run');
  const clearMissing = hasFlag(args, '--clear-missing');

  console.log(
    `[import-anthropic] validated models=${validation.modelIds.length} directory=${validation.providerDir}`,
  );

  const summary = await importProviderYaml({
    providersRoot,
    providerSlugs: [ANTHROPIC_PROVIDER_SLUG],
    dryRun,
    clearMissing,
  });

  if (
    summary.providersImported !== 1 ||
    summary.modelsImported !== ANTHROPIC_SUPPORTED_MODEL_IDS.length
  ) {
    throw new Error(
      `Anthropic import invariant failed: providers=${summary.providersImported} models=${summary.modelsImported}`,
    );
  }

  console.log(
    `[import-anthropic] completed models=${summary.modelsImported} deleted=${summary.modelsDeleted} dryRun=${summary.dryRun}`,
  );
}

void main()
  .catch((error) => {
    console.error(error);
    process.exitCode = 1;
  })
  .finally(disconnectProviderYamlImporter);
