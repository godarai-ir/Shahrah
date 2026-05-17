import { Body, Controller, Post } from '@nestjs/common';
import { ApiTags } from '@nestjs/swagger';
import { AuthService } from './auth.service';

class AdminTokenDto {
  userId!: string;
}

@ApiTags('auth')
@Controller({ path: 'auth', version: '1' })
export class AuthController {
  constructor(private readonly authService: AuthService) {}
  @Post('admin-token') issue(@Body() body: AdminTokenDto) {
    return { accessToken: this.authService.adminToken(body.userId) };
  }
}
