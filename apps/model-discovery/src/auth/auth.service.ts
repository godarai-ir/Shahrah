import { GoneException, Injectable } from '@nestjs/common';

@Injectable()
export class AuthService {
  login(_username: string, _password: string): never {
    throw new GoneException('Human login is owned by support-admin');
  }
}
