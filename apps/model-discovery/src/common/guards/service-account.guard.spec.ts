import { ExecutionContext, UnauthorizedException } from '@nestjs/common';
import { createHash } from 'crypto';
import { ServiceAccountGuard } from './service-account.guard';

function context(headers: Record<string, string>): ExecutionContext {
  const request = { headers };
  return {
    switchToHttp: () => ({ getRequest: () => request }),
  } as unknown as ExecutionContext;
}

describe('ServiceAccountGuard', () => {
  const secret = 'a'.repeat(43);
  const credential = {
    id: 'credential01',
    secretHash: createHash('sha256').update(secret).digest('hex'),
    revokedAt: null,
    expiresAt: null,
    lastUsedAt: new Date(),
    serviceAccount: {
      id: 'service-account-id', name: 'support-admin', active: true,
      scopes: ['catalog:read'],
    },
  };

  it('authenticates a valid service key and binds delegated actor context', async () => {
    const prisma = {
      serviceAccountCredential: {
        findUnique: jest.fn().mockResolvedValue(credential),
        update: jest.fn(),
      },
    };
    const guard = new ServiceAccountGuard(prisma as any);
    const ctx = context({
      authorization: `Bearer sa_credential01.${secret}`,
      'x-shahrah-actor-id': 'user_123',
      'x-shahrah-actor-role': 'catalog_admin',
    });

    await expect(guard.canActivate(ctx)).resolves.toBe(true);
    expect(ctx.switchToHttp().getRequest().user).toMatchObject({
      serviceAccountName: 'support-admin', actorId: 'user_123',
      scopes: ['catalog:read'],
    });
  });

  it('rejects an invalid secret', async () => {
    const prisma = {
      serviceAccountCredential: {
        findUnique: jest.fn().mockResolvedValue(credential),
      },
    };
    const guard = new ServiceAccountGuard(prisma as any);
    await expect(guard.canActivate(context({
      authorization: `Bearer sa_credential01.${'b'.repeat(43)}`,
    }))).rejects.toBeInstanceOf(UnauthorizedException);
  });
});
