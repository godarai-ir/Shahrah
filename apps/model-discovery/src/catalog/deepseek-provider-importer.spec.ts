import { cp, mkdtemp, readFile, rm, writeFile } from 'node:fs/promises';
import { tmpdir } from 'node:os';
import { join } from 'node:path';
import { parse } from 'yaml';
import {
  DEEPSEEK_SUPPORTED_MODEL_IDS,
  validateDeepSeekProviderDirectory,
} from '../../scripts/model-import/providers/deepseek';

describe('DeepSeek provider importer', () => {
  const providersRoot = join(__dirname, '..', '..', 'providers');
  const tempRoots: string[] = [];

  afterEach(async () => {
    await Promise.all(
      tempRoots.splice(0).map((path) =>
        rm(path, {
          recursive: true,
          force: true,
        }),
      ),
    );
  });

  it('validates exactly the two curated DeepSeek models', async () => {
    const result = await validateDeepSeekProviderDirectory(providersRoot);
    expect(result.modelIds).toEqual([...DEEPSEEK_SUPPORTED_MODEL_IDS].sort());
    expect(result.modelFiles).toHaveLength(2);
  });

  it('enables Responses create only for Flash', async () => {
    const flash = parse(
      await readFile(
        join(providersRoot, 'deepseek', 'deepseek-v4-flash.yaml'),
        'utf8',
      ),
    );
    const pro = parse(
      await readFile(
        join(providersRoot, 'deepseek', 'deepseek-v4-pro.yaml'),
        'utf8',
      ),
    );

    expect(flash.supportedModes).toContain('responses');
    expect(flash.runtimeProfile.endpoints).toContain('responses');
    expect(flash.runtimeProfile.responseResources).toEqual({
      store: false,
      retrieve: false,
      delete: false,
      compact: false,
      countInputTokens: false,
    });
    expect(pro.supportedModes).not.toContain('responses');
    expect(pro.runtimeProfile.endpoints).not.toContain('responses');
  });

  it('rejects missing models, zero pricing, and unofficial sources', async () => {
    const root = await mkdtemp(join(tmpdir(), 'deepseek-import-'));
    const sourceDir = join(providersRoot, 'deepseek');
    const targetDir = join(root, 'deepseek');
    const proPath = join(targetDir, 'deepseek-v4-pro.yaml');
    const defaultPath = join(targetDir, 'default.yaml');
    await cp(sourceDir, targetDir, {
      recursive: true,
    });
    const proDocument = await readFile(proPath, 'utf8');
    tempRoots.push(root);

    await writeFile(proPath, '');
    await expect(validateDeepSeekProviderDirectory(root)).rejects.toThrow(
      'DeepSeek model allowlist validation failed',
    );

    await writeFile(proPath, proDocument);
    await writeFile(
      defaultPath,
      'documentation: [https://example.com/deepseek]\n',
    );
    await expect(validateDeepSeekProviderDirectory(root)).rejects.toThrow(
      'official DeepSeek sources',
    );

    await cp(join(sourceDir, 'default.yaml'), defaultPath, { force: true });
    await writeFile(
      proPath,
      proDocument.replace(
        'input_cost_per_token: 0.000000435',
        'input_cost_per_token: 0',
      ),
    );
    await expect(validateDeepSeekProviderDirectory(root)).rejects.toThrow(
      'pricing:input_cost_per_token',
    );

    await writeFile(
      proPath,
      proDocument.replace('model: deepseek-v4-pro', 'model: deepseek-v4-flash'),
    );
    await expect(validateDeepSeekProviderDirectory(root)).rejects.toThrow(
      'identity',
    );

    await writeFile(proPath, proDocument);
    await cp(proPath, join(targetDir, 'unexpected.yaml'));
    await expect(validateDeepSeekProviderDirectory(root)).rejects.toThrow(
      'model allowlist validation failed',
    );
  });
});
