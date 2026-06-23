import { GoneException } from '@nestjs/common';
import { AuthService } from './auth.service';

describe('AuthService', () => {
  it('keeps human login permanently disabled', () => {
    const service = new AuthService();
    expect(() => service.login('admin', 'pass')).toThrow(GoneException);
  });
});
