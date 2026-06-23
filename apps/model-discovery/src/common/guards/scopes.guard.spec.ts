import { ExecutionContext, ForbiddenException } from '@nestjs/common';
import { ScopesGuard } from './scopes.guard';

describe('ScopesGuard', () => {
  const context = (scopes: string[]) => ({
    getHandler: () => null,
    getClass: () => null,
    switchToHttp: () => ({ getRequest: () => ({ user: { scopes } }) }),
  }) as unknown as ExecutionContext;

  it('allows principals containing every required scope', () => {
    const guard = new ScopesGuard({ getAllAndOverride: () => ['catalog:write'] } as any);
    expect(guard.canActivate(context(['catalog:read', 'catalog:write']))).toBe(true);
  });

  it('rejects a missing scope', () => {
    const guard = new ScopesGuard({ getAllAndOverride: () => ['catalog:write'] } as any);
    expect(() => guard.canActivate(context(['catalog:read']))).toThrow(ForbiddenException);
  });
});
