import { ServiceUnavailableException } from '@nestjs/common';
import { HealthController } from './health.controller';

describe('HealthController', () => {
  it('reports process health without checking dependencies', () => {
    const controller = new HealthController({} as any, {} as any);

    expect(controller.health()).toEqual({
      status: 'healthy',
      service: 'model-discovery',
    });
  });

  it('reports readiness when the database query succeeds', async () => {
    const prisma = {
      $queryRaw: jest.fn().mockResolvedValue([{ '?column?': 1 }]),
    } as any;
    const indexing = {
      ping: jest.fn().mockResolvedValue(true),
      isReadyStrict: jest.fn().mockReturnValue(true),
    } as any;
    const controller = new HealthController(prisma, indexing);

    await expect(controller.ready()).resolves.toEqual({
      status: 'ready',
      service: 'model-discovery',
      database: 'available',
      typesense: 'available',
    });
    expect(prisma.$queryRaw).toHaveBeenCalledTimes(1);
  });

  it('returns service unavailable when the database query fails', async () => {
    const prisma = {
      $queryRaw: jest.fn().mockRejectedValue(new Error('db unavailable')),
    } as any;
    const controller = new HealthController(prisma, {} as any);

    await expect(controller.ready()).rejects.toBeInstanceOf(ServiceUnavailableException);
  });

  it('fails readiness in strict mode when Typesense is unavailable', async () => {
    const prisma = {
      $queryRaw: jest.fn().mockResolvedValue([{ '?column?': 1 }]),
    } as any;
    const indexing = {
      ping: jest.fn().mockResolvedValue(false),
      isReadyStrict: jest.fn().mockReturnValue(true),
    } as any;
    const controller = new HealthController(prisma, indexing);

    await expect(controller.ready()).rejects.toBeInstanceOf(
      ServiceUnavailableException,
    );
  });

  it('reports degraded readiness when Typesense fallback is allowed', async () => {
    const prisma = {
      $queryRaw: jest.fn().mockResolvedValue([{ '?column?': 1 }]),
    } as any;
    const indexing = {
      ping: jest.fn().mockResolvedValue(false),
      isReadyStrict: jest.fn().mockReturnValue(false),
    } as any;
    const controller = new HealthController(prisma, indexing);

    await expect(controller.ready()).resolves.toEqual({
      status: 'ready',
      service: 'model-discovery',
      database: 'available',
      typesense: 'degraded',
    });
  });
});
