import { CatalogService } from './catalog.service';

describe('CatalogService', () => {
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
      data: [],
      pagination: { page: 1, limit: 24, total: 0, totalPages: 1 },
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
});
