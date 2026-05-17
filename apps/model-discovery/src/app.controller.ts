import { Controller, Get } from '@nestjs/common';

@Controller({ path: 'public/health', version: '1' })
export class AppController {
  @Get()
  health() {
    return { status: 'ok', service: 'model-discovery' };
  }
}
