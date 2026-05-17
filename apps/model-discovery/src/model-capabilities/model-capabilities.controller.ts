import { Controller, Get } from '@nestjs/common';
import { ApiTags } from '@nestjs/swagger';
import { ModelCapabilitiesService } from './model-capabilities.service';

@ApiTags('model-capabilities')
@Controller({ path: 'model-capabilities', version: '1' })
export class ModelCapabilitiesController {
  constructor(private readonly service: ModelCapabilitiesService) {}

  @Get('public')
  health() {
    return { module: 'model-capabilities', status: 'ok' };
  }
}
