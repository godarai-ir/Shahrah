import { findMissingProviderSlugs } from './provider-catalog-reconciliation';

describe('findMissingProviderSlugs', () => {
  it('returns database providers absent from the catalog', () => {
    expect(
      findMissingProviderSlugs(
        ['anthropic', 'cohere', 'openai'],
        ['anthropic', 'openai'],
      ),
    ).toEqual(['cohere']);
  });

  it('returns missing providers in stable sorted order', () => {
    expect(
      findMissingProviderSlugs(
        ['mistral-ai', 'anthropic', 'cohere'],
        ['anthropic'],
      ),
    ).toEqual(['cohere', 'mistral-ai']);
  });

  it('refuses to reconcile against an empty catalog', () => {
    expect(() => findMissingProviderSlugs(['anthropic', 'openai'], [])).toThrow(
      'Refusing to reconcile missing providers against an empty provider catalog',
    );
  });
});
