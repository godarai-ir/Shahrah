import { BadRequestException } from '@nestjs/common';
import { ProvidersController } from './providers.controller';
import { PROVIDER_LOGO_CACHE_CONTROL } from './provider-logo-image';

describe('ProvidersController', () => {
  it('rejects provider logo uploads with unsupported MIME types', () => {
    const controller = new ProvidersController({} as any);

    expect(() =>
      controller.uploadProviderLogo(
        'openai',
        { mimetype: 'image/gif', size: 100, buffer: Buffer.from('gif') },
        { user: {} } as any,
      ),
    ).toThrow(BadRequestException);
  });

  it('rejects provider logo uploads larger than 2MB', () => {
    const controller = new ProvidersController({} as any);

    expect(() =>
      controller.uploadProviderLogo(
        'openai',
        {
          mimetype: 'image/png',
          size: 2 * 1024 * 1024 + 1,
          buffer: Buffer.from('png'),
        },
        { user: {} } as any,
      ),
    ).toThrow('file size must be <= 2MB');
  });

  it('delegates valid provider logo uploads to the provider service', async () => {
    const providers = {
      saveProviderLogo: jest.fn().mockResolvedValue({ slug: 'openai' }),
    };
    const controller = new ProvidersController(providers as any);
    const file = {
      mimetype: 'image/webp',
      size: 1024,
      buffer: Buffer.from('webp'),
    };
    const request = {
      user: {
        serviceAccountId: 'svc-1',
        serviceAccountName: 'support-admin',
        scopes: ['provider:write'],
      },
    } as any;

    await expect(
      controller.uploadProviderLogo('openai', file, request),
    ).resolves.toEqual({ slug: 'openai' });
    expect(providers.saveProviderLogo).toHaveBeenCalledWith(
      'openai',
      file,
      request.user,
    );
  });

  it('serves provider logo assets from the provider service with long cache headers', async () => {
    const providers = {
      getProviderLogoAsset: jest.fn().mockResolvedValue({
        body: Buffer.from([1, 2, 3]),
        contentType: 'image/png',
      }),
    };
    const controller = new ProvidersController(providers as any);
    const res = mockResponse();

    await controller.getAsset('provider-logos', 'openai-logo.png', res as any);

    expect(providers.getProviderLogoAsset).toHaveBeenCalledWith(
      'openai-logo.png',
    );
    expect(res.setHeader).toHaveBeenCalledWith('Content-Type', 'image/png');
    expect(res.setHeader).toHaveBeenCalledWith(
      'Cache-Control',
      PROVIDER_LOGO_CACHE_CONTROL,
    );
    expect(res.status).toHaveBeenCalledWith(200);
    expect(res.send).toHaveBeenCalledWith(Buffer.from([1, 2, 3]));
  });

  it('returns 404 when a provider logo asset is not in the database', async () => {
    const providers = {
      getProviderLogoAsset: jest.fn().mockResolvedValue(null),
    };
    const controller = new ProvidersController(providers as any);
    const res = mockResponse();

    await controller.getAsset('provider-logos', 'missing.png', res as any);

    expect(res.status).toHaveBeenCalledWith(404);
    expect(res.send).toHaveBeenCalledWith('Not found');
  });
});

function mockResponse() {
  return {
    setHeader: jest.fn().mockReturnThis(),
    status: jest.fn().mockReturnThis(),
    send: jest.fn().mockReturnThis(),
    sendFile: jest.fn().mockReturnThis(),
  };
}
