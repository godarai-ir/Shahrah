import { cp, mkdtemp, readFile, rm, writeFile } from 'node:fs/promises';
import { tmpdir } from 'node:os';
import { join } from 'node:path';
import { parse } from 'yaml';
import {
  ZAI_SUPPORTED_MODEL_IDS,
  validateZAIProviderDirectory,
} from '../../scripts/model-import/providers/zai';

describe('Z.ai provider importer', () => {
  const providersRoot = join(__dirname, '..', '..', 'providers');
  const tempRoots: string[] = [];

  afterEach(async () => {
    await Promise.all(
      tempRoots
        .splice(0)
        .map((path) => rm(path, { recursive: true, force: true })),
    );
  });

  it('validates exactly the four curated Z.ai models and their prices', async () => {
    const result = await validateZAIProviderDirectory(providersRoot);
    expect(result.modelIds).toEqual([...ZAI_SUPPORTED_MODEL_IDS].sort());
    expect(result.modelFiles).toHaveLength(4);

    const ocr = parse(
      await readFile(join(providersRoot, 'zai', 'glm-ocr.yaml'), 'utf8'),
    );
    expect(ocr.runtimeProfile.endpoints).toEqual(['layout_parsing']);
    expect(ocr.costs[0]).toEqual(
      expect.objectContaining({
        input_cost_per_token: 0.00000003,
        output_cost_per_token: 0.00000003,
      }),
    );
  });

  it('keeps disputed GLM-5V tools and single-request OCR limits gated', async () => {
    const vision = parse(
      await readFile(join(providersRoot, 'zai', 'glm-5v-turbo.yaml'), 'utf8'),
    );
    expect(vision.runtimeProfile.tools.custom).toBe(false);
    expect(vision.runtimeProfile.unsupportedParameters).toContain('tools');
    expect(vision.runtimeProfile.tools.structuredOutput).toBe(false);
    expect(vision.runtimeProfile.unsupportedParameters).toContain(
      'response_format',
    );
    expect(vision.runtimeProfile.parameters.chat_completions).not.toContain(
      'response_format',
    );
  });

  it('rejects missing models, unofficial sources, and zero pricing', async () => {
    const root = await mkdtemp(join(tmpdir(), 'zai-import-'));
    const sourceDir = join(providersRoot, 'zai');
    const targetDir = join(root, 'zai');
    const modelPath = join(targetDir, 'glm-5.2.yaml');
    const defaultPath = join(targetDir, 'default.yaml');
    await cp(sourceDir, targetDir, { recursive: true });
    const modelDocument = await readFile(modelPath, 'utf8');
    tempRoots.push(root);

    await writeFile(defaultPath, 'documentation: [https://example.com/zai]\n');
    await expect(validateZAIProviderDirectory(root)).rejects.toThrow(
      'official Z.ai sources',
    );

    await cp(join(sourceDir, 'default.yaml'), defaultPath, { force: true });
    await writeFile(
      modelPath,
      modelDocument.replace(
        'input_cost_per_token: 0.0000014',
        'input_cost_per_token: 0',
      ),
    );
    await expect(validateZAIProviderDirectory(root)).rejects.toThrow(
      'pricing:input_cost_per_token',
    );

    await writeFile(modelPath, modelDocument);
    await cp(modelPath, join(targetDir, 'unexpected.yaml'));
    await expect(validateZAIProviderDirectory(root)).rejects.toThrow(
      'model allowlist validation failed',
    );
  });
});
