import { Injectable, UnauthorizedException } from '@nestjs/common';
import { ConfigService } from '@nestjs/config';
import { JwtService } from '@nestjs/jwt';

@Injectable()
export class AuthService {
  constructor(
    private readonly cfg: ConfigService,
    private readonly jwt: JwtService,
  ) {}

  login(username: string, password: string) {
    const expectedUser = this.cfg.get<string>('ADMIN_USERNAME', 'admin');
    const expectedPass = this.cfg.get<string>('ADMIN_PASSWORD', 'admin123');
    if (username !== expectedUser || password !== expectedPass) {
      throw new UnauthorizedException('Invalid credentials');
    }
    return { accessToken: this.jwt.sign({ sub: username, roles: ['admin'] }) };
  }
}
