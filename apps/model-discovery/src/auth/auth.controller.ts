import { Body, Controller, Post } from '@nestjs/common';
import { ApiBody, ApiOkResponse, ApiProperty, ApiTags } from '@nestjs/swagger';
import { IsString } from 'class-validator';
import { AuthService } from './auth.service';

class AdminLoginDto {
  @ApiProperty()
  @IsString()
  username!: string;

  @ApiProperty()
  @IsString()
  password!: string;
}

class AdminLoginResponse {
  @ApiProperty()
  accessToken!: string;
}

@ApiTags('auth')
@Controller({ path: 'auth', version: '1' })
export class AuthController {
  constructor(private readonly authService: AuthService) {}

  @Post('login')
  @ApiBody({ type: AdminLoginDto })
  @ApiOkResponse({ type: AdminLoginResponse })
  login(@Body() body: AdminLoginDto) {
    return this.authService.login(body.username, body.password);
  }
}
