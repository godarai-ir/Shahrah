import {
  DEFAULT_PROVIDERS_ROOT,
  disconnectProviderYamlImporter,
  importProviderYaml,
} from './import-models-from-yaml';
import {
  OPENAI_PROVIDER_SLUG,
  OPENAI_SUPPORTED_MODEL_IDS,
  validateOpenAIProviderDirectory,
} from './model-import/providers/openai';

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
      'The OpenAI importer is locked to provider=openai; remove --providers.',
    );
  }

  const providersRoot =
    getArgValue(args, '--providers-root') ?? DEFAULT_PROVIDERS_ROOT;
  const validation = await validateOpenAIProviderDirectory(providersRoot);
  const dryRun = hasFlag(args, '--dry-run');
  const clearMissing = hasFlag(args, '--clear-missing');

  console.log(
    `[import-openai] validated models=${validation.modelIds.length} directory=${validation.providerDir}`,
  );

  const summary = await importProviderYaml({
    providersRoot,
    providerSlugs: [OPENAI_PROVIDER_SLUG],
    dryRun,
    clearMissing,
  });

  if (
    summary.providersImported !== 1 ||
    summary.modelsImported !== OPENAI_SUPPORTED_MODEL_IDS.length
  ) {
    throw new Error(
      `OpenAI import invariant failed: providers=${summary.providersImported} models=${summary.modelsImported}`,
    );
  }

  console.log(
    `[import-openai] completed models=${summary.modelsImported} deleted=${summary.modelsDeleted} dryRun=${summary.dryRun}`,
  );
}

void main()
  .catch((error) => {
    console.error(error);
    process.exitCode = 1;
  })
  .finally(disconnectProviderYamlImporter);
