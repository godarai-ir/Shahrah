import {
  DEFAULT_PROVIDERS_ROOT,
  disconnectProviderYamlImporter,
  importProviderYaml,
  verifyActiveProviderModelInvariant,
} from './import-models-from-yaml';
import {
  GOOGLE_GEMINI_PROVIDER_SLUG,
  GOOGLE_GEMINI_SUPPORTED_MODEL_IDS,
  parseGoogleGeminiImportArgs,
  validateGoogleGeminiProviderDirectory,
} from './model-import/providers/google-gemini';

async function main(args: string[] = process.argv.slice(2)): Promise<void> {
  const options = parseGoogleGeminiImportArgs(args);
  const providersRoot = options.providersRoot ?? DEFAULT_PROVIDERS_ROOT;
  const validation = await validateGoogleGeminiProviderDirectory(providersRoot);
  const { dryRun, clearMissing } = options;

  console.log(
    `[import-google-gemini] validated models=${validation.modelIds.length} directory=${validation.providerDir}`,
  );
  const summary = await importProviderYaml({
    providersRoot,
    providerSlugs: [GOOGLE_GEMINI_PROVIDER_SLUG],
    dryRun,
    clearMissing,
  });
  if (
    summary.providersImported !== 1 ||
    summary.modelsImported !== GOOGLE_GEMINI_SUPPORTED_MODEL_IDS.length
  ) {
    throw new Error(
      `Google Gemini import invariant failed: providers=${summary.providersImported} models=${summary.modelsImported}`,
    );
  }
  if (clearMissing && !dryRun) {
    await verifyActiveProviderModelInvariant(
      GOOGLE_GEMINI_PROVIDER_SLUG,
      GOOGLE_GEMINI_SUPPORTED_MODEL_IDS,
    );
  }
  console.log(
    `[import-google-gemini] completed models=${summary.modelsImported} reconciled=${summary.modelsDeleted} dryRun=${summary.dryRun}`,
  );
}

void main()
  .catch((error) => {
    console.error(error);
    process.exitCode = 1;
  })
  .finally(disconnectProviderYamlImporter);
