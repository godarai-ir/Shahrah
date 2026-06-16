import { ServiceUnavailableException } from '@nestjs/common';
import { HealthController } from './health.controller';

describe('HealthController', () => {
  it('reports process health without checking dependencies', () => {
    const controller = new HealthController({} as any);

    expect(controller.health()).toEqual({
      status: 'healthy',
      service: 'model-discovery',
    });
  });

  it('reports readiness when the database query succeeds', async () => {
    const prisma = {
      $queryRaw: jest.fn().mockResolvedValue([{ '?column?': 1 }]),
    } as any;
    const controller = new HealthController(prisma);

    await expect(controller.ready()).resolves.toEqual({
      status: 'ready',
      service: 'model-discovery',
      database: 'available',
    });
    expect(prisma.$queryRaw).toHaveBeenCalledTimes(1);
  });

  it('returns service unavailable when the database query fails', async () => {
    const prisma = {
      $queryRaw: jest.fn().mockRejectedValue(new Error('db unavailable')),
    } as any;
    const controller = new HealthController(prisma);

    await expect(controller.ready()).rejects.toBeInstanceOf(ServiceUnavailableException);
  });
});
