import { ProvidersService } from './providers.service';
import sharp from 'sharp';

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

  it('resizes and stores provider logos as database-backed PNG assets', async () => {
    const beforeProvider = {
      id: 'provider-1',
      slug: 'openai',
      name: 'OpenAI',
      logoUrl: null,
    };
    const afterProvider = {
      ...beforeProvider,
      logoUrl: '/api/v1/providers/assets/provider-logos/openai-new.png',
    };
    const tx = {
      provider: {
        findUniqueOrThrow: jest.fn().mockResolvedValue(beforeProvider),
        update: jest.fn().mockResolvedValue(afterProvider),
      },
      providerLogoAsset: {
        create: jest.fn().mockResolvedValue({ id: 'asset-1' }),
      },
      auditLog: {
        create: jest.fn().mockResolvedValue({ id: 'audit-1' }),
      },
    };
    const prisma = {
      $transaction: jest.fn(async (callback) => callback(tx)),
    } as any;
    const indexing = {
      enqueueProviderModelsReindex: jest.fn().mockResolvedValue(undefined),
    };
    const service = new ProvidersService(prisma, indexing as any);
    const upload = await sharp({
      create: {
        width: 160,
        height: 48,
        channels: 3,
        background: { r: 28, g: 110, b: 190 },
      },
    })
      .jpeg()
      .toBuffer();

    const result = await service.saveProviderLogo(
      'openai',
      { buffer: upload, mimetype: 'image/jpeg', originalname: 'logo.jpg' },
      {
        serviceAccountId: 'svc-1',
        serviceAccountName: 'support-admin',
        scopes: ['provider:write'],
      },
    );

    expect(result).toBe(afterProvider);
    expect(tx.providerLogoAsset.create).toHaveBeenCalledTimes(1);
    const createData = tx.providerLogoAsset.create.mock.calls[0][0].data;
    const metadata = await sharp(createData.data).metadata();
    expect(createData.providerId).toBe('provider-1');
    expect(createData.fileName).toMatch(
      /^openai-[a-f0-9]{16}-[a-f0-9-]{36}\.png$/,
    );
    expect(createData.contentType).toBe('image/png');
    expect(createData.width).toBe(64);
    expect(createData.height).toBe(64);
    expect(createData.sizeBytes).toBe(createData.data.byteLength);
    expect(metadata.format).toBe('png');
    expect(metadata.width).toBe(64);
    expect(metadata.height).toBe(64);
    expect(tx.provider.update).toHaveBeenCalledWith({
      where: { slug: 'openai' },
      data: {
        logoUrl: `/api/v1/providers/assets/provider-logos/${createData.fileName}`,
      },
    });
    expect(indexing.enqueueProviderModelsReindex).toHaveBeenCalledWith(
      'openai',
    );
  });

  it('reads provider logo assets from the database by file name', async () => {
    const prisma = {
      providerLogoAsset: {
        findUnique: jest.fn().mockResolvedValue({
          data: new Uint8Array([1, 2, 3]),
          contentType: 'image/png',
        }),
      },
    } as any;
    const service = new ProvidersService(prisma, {} as any);

    const result = await service.getProviderLogoAsset('openai-logo.png');

    expect(prisma.providerLogoAsset.findUnique).toHaveBeenCalledWith({
      where: { fileName: 'openai-logo.png' },
      select: {
        data: true,
        contentType: true,
      },
    });
    expect(result?.contentType).toBe('image/png');
    expect(result?.body).toEqual(Buffer.from([1, 2, 3]));
  });
});
