import { ProvidersService } from './providers.service';

describe('ProvidersService', () => {
  it('returns the enabled model count with each provider', async () => {
    const prisma = {
      provider: {
        findMany: jest.fn().mockResolvedValue([
          {
            id: 'provider-1',
            slug: 'anthropic',
            name: 'Anthropic',
            enabled: true,
            models: [{ id: 'model-1' }, { id: 'model-2' }],
            connectionProfile: { configJson: { api_key: 'must-not-leak' } },
          },
        ]),
      },
    } as any;
    const service = new ProvidersService(prisma, {} as any);

    const result = await service.list();
    expect(result).toEqual([
      expect.objectContaining({ slug: 'anthropic', modelCount: 2 }),
    ]);
    expect(JSON.stringify(result)).not.toContain('connectionProfile');
    expect(JSON.stringify(result)).not.toContain('must-not-leak');
    expect(prisma.provider.findMany).toHaveBeenCalledWith(
      expect.objectContaining({
        where: { enabled: true },
        include: expect.objectContaining({
          models: { where: { enabled: true } },
        }),
      }),
    );
  });

  it('lists enabled provider models in the control-plane catalog shape', async () => {
    const prisma = {
      model: {
        findMany: jest.fn().mockResolvedValue([
          {
            canonicalId: 'anthropic/claude-sonnet-4',
            slug: 'claude-sonnet-4',
            displayName: 'Claude Sonnet 4',
            mode: ['chat'],
            contextWindow: 200000,
            maxOutputTokens: 64000,
            inputModalities: [{ modality: { slug: 'text' } }],
            outputModalities: [{ modality: { slug: 'text' } }],
            supportedParams: [{ supportedParameter: { slug: 'tools' } }],
            capabilities: [{ capability: { slug: 'reasoning' } }],
            regions: [{ region: { slug: 'global' } }],
            pricing: [
              {
                currency: 'USD',
                inputPricePerMillionTokens: 3,
                outputPricePerMillionTokens: 15,
                cachedInputPricePerMillionTokens: 0.3,
              },
            ],
          },
        ]),
      },
    } as any;
    const service = new ProvidersService(prisma, {} as any);

    await expect(service.listModels('anthropic')).resolves.toEqual({
      data: [
        {
          canonicalId: 'anthropic/claude-sonnet-4',
          slug: 'claude-sonnet-4',
          name: 'Claude Sonnet 4',
          mode: ['chat'],
          contextWindow: 200000,
          maxOutputTokens: 64000,
          inputModalities: ['text'],
          outputModalities: ['text'],
          supportedParameters: ['tools'],
          capabilities: ['reasoning'],
          pricing: {
            currency: 'USD',
            inputPerMillionTokens: 3,
            outputPerMillionTokens: 15,
            cachedInputPerMillionTokens: 0.3,
          },
          regions: ['global'],
        },
      ],
    });
    expect(prisma.model.findMany).toHaveBeenCalledWith(
      expect.objectContaining({
        where: {
          enabled: true,
          provider: { slug: 'anthropic', enabled: true },
        },
      }),
    );
  });
});
