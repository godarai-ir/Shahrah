import { cp, mkdtemp, rm, writeFile } from 'node:fs/promises';
import { tmpdir } from 'node:os';
import { join } from 'node:path';
import {
  MOONSHOT_SUPPORTED_MODEL_IDS,
  validateMoonshotProviderDirectory,
} from '../../scripts/model-import/providers/moonshot';

describe('Moonshot provider importer', () => {
  const providersRoot = join(__dirname, '..', '..', 'providers');

  it('validates exactly the four curated Kimi models', async () => {
    const result = await validateMoonshotProviderDirectory(providersRoot);
    expect(result.modelIds).toEqual([...MOONSHOT_SUPPORTED_MODEL_IDS].sort());
    expect(result.modelFiles).toHaveLength(4);
  });

  it('rejects missing models and non-official sources', async () => {
    const root = await mkdtemp(join(tmpdir(), 'moonshot-import-'));
    await cp(join(providersRoot, 'moonshot'), join(root, 'moonshot'), {
      recursive: true,
    });

    await writeFile(join(root, 'moonshot', 'kimi-k3.yaml'), '');
    await expect(validateMoonshotProviderDirectory(root)).rejects.toThrow(
      'Moonshot model allowlist validation failed',
    );

    await cp(
      join(providersRoot, 'moonshot', 'kimi-k3.yaml'),
      join(root, 'moonshot', 'kimi-k3.yaml'),
      { force: true },
    );
    await writeFile(
      join(root, 'moonshot', 'default.yaml'),
      'documentation: [https://example.com/moonshot]\n',
    );
    await expect(validateMoonshotProviderDirectory(root)).rejects.toThrow(
      'official Kimi sources',
    );

    await rm(root, { recursive: true, force: true });
  });
});
