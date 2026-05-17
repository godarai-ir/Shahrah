import { Controller, Get } from '@nestjs/common';
import { ApiTags } from '@nestjs/swagger';
import { ModelBenchmarksService } from './model-benchmarks.service';

@ApiTags('model-benchmarks')
@Controller({ path: 'model-benchmarks', version: '1' })
export class ModelBenchmarksController {
  constructor(private readonly service: ModelBenchmarksService) {}

  @Get('public')
  health() {
    return { module: 'model-benchmarks', status: 'ok' };
  }
}
