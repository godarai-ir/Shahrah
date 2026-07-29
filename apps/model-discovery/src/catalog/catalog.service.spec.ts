import { CatalogService } from './catalog.service';

describe('CatalogService', () => {
  it('fails the runtime snapshot when a stored profile is invalid', async () => {
    const prisma = {
      model: {
        findMany: jest.fn().mockResolvedValue([
          {
            canonicalId: 'anthropic/claude-opus-4-7',
            slug: 'claude-opus-4-7',
            status: 'active',
            deprecationDate: null,
            retirementDate: null,
            replacementCanonicalId: null,
            runtimeProfiles: [
              {
                version: 1,
                profileJson: { version: 1 },
                sourceUrl: 'https://platform.claude.com/docs',
                effectiveFrom: new Date(),
              },
            ],
            pricingRates: [],
            sources: [],
          },
        ]),
      },
    } as any;
    const service = new CatalogService(prisma, {} as any);

    await expect(service.runtimeSnapshot('anthropic')).rejects.toThrow();
  });

  it('applies enabled/provider-enabled filters', async () => {
    const prisma = {
      model: {
        count: jest.fn().mockResolvedValue(0),
        findMany: jest.fn().mockResolvedValue([]),
      },
    } as any;
    const indexing = {
      searchModels: jest.fn().mockResolvedValue(null),
      isReadPathEnabled: jest.fn().mockReturnValue(true),
    } as any;
    const service = new CatalogService(prisma, indexing);
    await service.listModels({ page: 1, limit: 24, sort: 'newest' } as any);
    expect(prisma.model.count).toHaveBeenCalledWith(
      expect.objectContaining({
        where: expect.objectContaining({ enabled: true, provider: expect.objectContaining({ enabled: true }) }),
      }),
    );
  });

  it('prefers the Typesense-backed read model when available', async () => {
    const indexedResponse = {
      data: [{ canonicalId: 'openai/gpt-4.1' }],
      pagination: { page: 1, limit: 24, total: 1, totalPages: 1 },
      meta: { source: 'typesense', fallbackUsed: false },
    };
    const indexing = {
      searchModels: jest.fn().mockResolvedValue(indexedResponse),
    } as any;
    const service = new CatalogService({} as any, indexing);

    await expect(
      service.listModels({ page: 1, limit: 24, sort: 'newest' } as any),
    ).resolves.toBe(indexedResponse);
  });

  it('falls back to Postgres when Typesense is empty but enabled models exist', async () => {
    const indexedResponse = {
      data: [],
      pagination: { page: 1, limit: 24, total: 0, totalPages: 1 },
      meta: { source: 'typesense', fallbackUsed: false },
    };
    const prisma = {
      model: {
        count: jest
          .fn()
          .mockResolvedValueOnce(12)
          .mockResolvedValueOnce(0),
        findMany: jest.fn().mockResolvedValue([]),
      },
    } as any;
    const indexing = {
      searchModels: jest.fn().mockResolvedValue(indexedResponse),
      isReadPathEnabled: jest.fn().mockReturnValue(true),
    } as any;
    const service = new CatalogService(prisma, indexing);

    const result = await service.listModels({
      page: 1,
      limit: 24,
      sort: 'newest',
    } as any);

    expect(prisma.model.count).toHaveBeenNthCalledWith(1, {
      where: { enabled: true, provider: { enabled: true } },
    });
    expect(prisma.model.findMany).toHaveBeenCalled();
    expect(result.meta).toEqual({
      source: 'database',
      fallbackUsed: true,
    });
  });
});
