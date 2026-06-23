import { AuthService } from './auth.service';

// Intentionally has no @Controller decorator. Human login belongs to support-admin.
export class AuthController {
  constructor(private readonly authService: AuthService) {}

  login(username: string, password: string) {
    return this.authService.login(username, password);
  }
}
