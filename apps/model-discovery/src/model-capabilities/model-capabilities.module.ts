import { Module } from '@nestjs/common';
import { ModelCapabilitiesController } from './model-capabilities.controller';
import { ModelCapabilitiesService } from './model-capabilities.service';

@Module({
  controllers: [ModelCapabilitiesController],
  providers: [ModelCapabilitiesService],
})
export class ModelCapabilitiesModule {}
