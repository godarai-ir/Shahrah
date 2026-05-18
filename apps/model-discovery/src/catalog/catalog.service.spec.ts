import { CatalogService } from './catalog.service';

describe('CatalogService', () => {
  it('applies enabled/provider-enabled filters', async () => {
    const prisma = {
      model: {
        count: jest.fn().mockResolvedValue(0),
        findMany: jest.fn().mockResolvedValue([]),
      },
    } as any;
    const service = new CatalogService(prisma);
    await service.listModels({ page: 1, limit: 24, sort: 'newest' } as any);
    expect(prisma.model.count).toHaveBeenCalledWith(
      expect.objectContaining({
        where: expect.objectContaining({ enabled: true, provider: expect.objectContaining({ enabled: true }) }),
      }),
    );
  });
});
