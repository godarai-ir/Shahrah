import { UnauthorizedException } from '@nestjs/common';
import { AuthService } from './auth.service';

describe('AuthService', () => {
  it('returns token for valid creds', () => {
    const service = new AuthService(
      { get: (key: string, fallback: string) => (key === 'ADMIN_USERNAME' ? 'admin' : key === 'ADMIN_PASSWORD' ? 'pass' : fallback) } as any,
      { sign: () => 'jwt-token' } as any,
    );
    expect(service.login('admin', 'pass')).toEqual({ accessToken: 'jwt-token' });
  });

  it('throws for invalid creds', () => {
    const service = new AuthService(
      { get: (_k: string, fallback: string) => fallback } as any,
      { sign: () => 'jwt-token' } as any,
    );
    expect(() => service.login('x', 'y')).toThrow(UnauthorizedException);
  });
});
