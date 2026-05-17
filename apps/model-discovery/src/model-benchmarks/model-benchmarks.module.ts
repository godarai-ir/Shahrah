import { Module } from '@nestjs/common';
import { ModelBenchmarksController } from './model-benchmarks.controller';
import { ModelBenchmarksService } from './model-benchmarks.service';

@Module({
  controllers: [ModelBenchmarksController],
  providers: [ModelBenchmarksService],
})
export class ModelBenchmarksModule {}
